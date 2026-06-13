package authflow

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gobankcli/internal/archive"
	"gobankcli/internal/config"
	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

type SessionExchanger interface {
	ExchangeSession(ctx context.Context, code string) (provider.ConnectionSession, []provider.Account, error)
}

type Flow struct {
	Config        config.Config
	Provider      provider.Provider
	Exchanger     SessionExchanger
	Store         *store.Store
	InstitutionID string
	ListenAddress string
	HTTPS         bool
	CertPath      string
	KeyPath       string
	Timeout       time.Duration
	Stderr        io.Writer
}

type Report struct {
	ProviderConnectionID string `json:"provider_connection_id"`
	ConnectionID         string `json:"connection_id"`
	Status               string `json:"status"`
	Accounts             int    `json:"accounts"`
}

func (f Flow) Run(ctx context.Context) (Report, error) {
	if f.Provider == nil {
		return Report{}, errors.New("provider is required")
	}
	if f.Exchanger == nil {
		return Report{}, errors.New("session exchanger is required")
	}
	if f.Store == nil {
		return Report{}, errors.New("store is required")
	}
	listener, callbackAddress, err := listenLoopback(f.ListenAddress)
	if err != nil {
		return Report{}, err
	}
	defer listener.Close()

	callbackPath := "/" + f.Provider.Name() + "/callback"
	redirectURL := localCallbackURL(callbackAddress, callbackPath, f.HTTPS)
	callbackTLS := localCallbackTLSConfig{
		Enabled:  f.HTTPS,
		CertPath: f.CertPath,
		KeyPath:  f.KeyPath,
	}
	if f.HTTPS {
		cert, err := localCallbackCertificate(f.CertPath, f.KeyPath)
		if err != nil {
			return Report{}, err
		}
		callbackTLS.Certificate = &cert
	}

	session, err := f.Provider.StartConnection(ctx, f.InstitutionID, redirectURL)
	if err != nil {
		return Report{}, err
	}
	if _, err := f.Store.UpsertConnection(ctx, session.Connection); err != nil {
		return Report{}, err
	}
	if f.Stderr != nil {
		fmt.Fprintf(f.Stderr, "Open this URL: %s\n", session.RedirectURL)
		fmt.Fprintf(f.Stderr, "Waiting for callback on %s\n", redirectURL)
		if f.HTTPS && f.CertPath == "" && f.KeyPath == "" {
			fmt.Fprintln(f.Stderr, "The browser may show a warning for the temporary localhost certificate.")
		}
	}

	callback, err := waitForLocalCallback(ctx, listener, callbackPath, session.Connection.ProviderConnectionID, f.Timeout, callbackTLS)
	if err != nil {
		return Report{}, err
	}
	return f.completeAuthorization(ctx, callback)
}

func (f Flow) completeAuthorization(ctx context.Context, callback callbackParams) (Report, error) {
	if callback.State == "" {
		return Report{}, errors.New("state is required")
	}
	found, err := f.Store.ConnectionExists(ctx, f.Provider.Name(), callback.State)
	if err != nil {
		return Report{}, err
	}
	if !found {
		return Report{}, errors.New("state does not match a pending connection")
	}

	session, accounts, err := f.Exchanger.ExchangeSession(ctx, callback.Code)
	if err != nil {
		return Report{}, err
	}
	if session.Connection.InstitutionID == "" {
		session.Connection.InstitutionID = strings.TrimSpace(f.InstitutionID)
	}
	if session.Connection.InstitutionID == "" {
		return Report{}, errors.New("institution is required when the provider session response omits it")
	}
	connectionID, err := f.Store.UpsertConnection(ctx, session.Connection)
	if err != nil {
		return Report{}, err
	}
	for i := range accounts {
		if accounts[i].InstitutionID == "" {
			accounts[i].InstitutionID = session.Connection.InstitutionID
		}
	}
	if _, err := archive.NewManager(f.Config, f.Provider, f.Store).ArchiveAccounts(ctx, connectionID, accounts); err != nil {
		return Report{}, err
	}
	return Report{
		ProviderConnectionID: session.Connection.ProviderConnectionID,
		ConnectionID:         connectionID,
		Status:               session.Connection.Status,
		Accounts:             len(accounts),
	}, nil
}

func listenLoopback(address string) (net.Listener, string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, "", errors.New("listen address is required")
	}
	if strings.HasPrefix(address, ":") {
		address = "127.0.0.1" + address
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, "", fmt.Errorf("listen address must be host:port: %w", err)
	}
	if port == "" {
		return nil, "", errors.New("listen address port is required")
	}
	if host == "" {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, "", errors.New("listen address must be loopback")
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return nil, "", err
	}
	actualPort := port
	if port == "0" {
		_, actualPort, err = net.SplitHostPort(listener.Addr().String())
		if err != nil {
			_ = listener.Close()
			return nil, "", err
		}
	}
	return listener, net.JoinHostPort(host, actualPort), nil
}

func localCallbackURL(address, callbackPath string, https bool) string {
	scheme := "http"
	if https {
		scheme = "https"
	}
	return (&url.URL{Scheme: scheme, Host: address, Path: callbackPath}).String()
}

type localCallbackTLSConfig struct {
	Enabled     bool
	CertPath    string
	KeyPath     string
	Certificate *tls.Certificate
}

func waitForLocalCallback(ctx context.Context, listener net.Listener, callbackPath, expectedState string, timeout time.Duration, tlsConfig localCallbackTLSConfig) (callbackParams, error) {
	result := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		callbackState := strings.TrimSpace(query.Get("state"))
		if providerError := strings.TrimSpace(query.Get("error")); providerError != "" {
			if callbackState == "" || callbackState != expectedState {
				http.Error(w, "state mismatch", http.StatusBadRequest)
				return
			}
			http.Error(w, providerError, http.StatusBadRequest)
			select {
			case result <- callbackResult{err: fmt.Errorf("provider callback error: %s", providerError)}:
			default:
			}
			return
		}
		callback := callbackParams{
			Code:  strings.TrimSpace(query.Get("code")),
			State: callbackState,
		}
		if callback.Code == "" || callback.State == "" {
			http.Error(w, "missing code or state", http.StatusBadRequest)
			return
		}
		if callback.State != expectedState {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Authorization received. You may close this page.")
		select {
		case result <- callbackResult{callback: callback}:
		default:
		}
	})
	serveListener := listener
	if tlsConfig.Enabled {
		cert := tlsConfig.Certificate
		if cert == nil {
			loaded, err := localCallbackCertificate(tlsConfig.CertPath, tlsConfig.KeyPath)
			if err != nil {
				return callbackParams{}, err
			}
			cert = &loaded
		}
		serveListener = tls.NewListener(listener, &tls.Config{
			Certificates: []tls.Certificate{*cert},
			MinVersion:   tls.VersionTLS12,
		})
	}
	go func() {
		if err := server.Serve(serveListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case result <- callbackResult{err: err}:
			default:
			}
		}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	select {
	case item := <-result:
		return item.callback, item.err
	case <-timer.C:
		return callbackParams{}, errors.New("timed out waiting for callback")
	case <-ctx.Done():
		return callbackParams{}, ctx.Err()
	}
}

type callbackParams struct {
	Code  string
	State string
}

type callbackResult struct {
	callback callbackParams
	err      error
}
