package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gobankcli/internal/provider"
	"gobankcli/internal/provider/enablebanking"
	"gobankcli/internal/store"
)

type ConnectCmd struct {
	Provider        string        `help:"Provider name." default:""`
	Institution     string        `help:"Provider institution ID." required:""`
	RedirectURL     string        `name:"redirect" help:"Redirect URL registered with the provider."`
	Listen          string        `help:"Listen on a loopback address for one provider callback, e.g. 127.0.0.1:8787."`
	ListenHTTPS     bool          `name:"listen-https" help:"Serve the local callback listener over HTTPS."`
	ListenCert      string        `name:"listen-cert" help:"TLS certificate path for --listen-https."`
	ListenKey       string        `name:"listen-key" help:"TLS private key path for --listen-https."`
	CallbackTimeout time.Duration `name:"callback-timeout" help:"How long --listen waits for the callback." default:"5m"`
}

type connectReport struct {
	ProviderConnectionID string `json:"provider_connection_id"`
	ConnectionID         string `json:"connection_id"`
	Status               string `json:"status"`
	RedirectURL          string `json:"redirect_url"`
}

func (c ConnectCmd) Run(ctx context.Context, app *App) error {
	if c.Institution == "" {
		return errors.New("institution is required")
	}
	p, err := newProvider(firstString(c.Provider, app.Config.DefaultProvider))
	if err != nil {
		return err
	}
	if c.ListenHTTPS && strings.TrimSpace(c.Listen) == "" {
		return errors.New("listen-https requires listen")
	}
	if !c.ListenHTTPS && (strings.TrimSpace(c.ListenCert) != "" || strings.TrimSpace(c.ListenKey) != "") {
		return errors.New("listen-cert and listen-key require listen-https")
	}
	if strings.TrimSpace(c.Listen) != "" {
		exchanger, ok := p.(enablebanking.SessionExchanger)
		if !ok {
			return fmt.Errorf("provider %s does not support local callback authorization", p.Name())
		}
		return c.runWithLocalCallback(ctx, app, p, exchanger)
	}
	if strings.TrimSpace(c.RedirectURL) == "" {
		return errors.New("redirect is required unless listen is set")
	}
	session, err := p.StartConnection(ctx, c.Institution, c.RedirectURL)
	if err != nil {
		return err
	}
	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()
	connectionID, err := s.UpsertConnection(ctx, session.Connection)
	if err != nil {
		return err
	}
	return app.Out.Write(connectReport{
		ProviderConnectionID: session.Connection.ProviderConnectionID,
		ConnectionID:         connectionID,
		Status:               session.Connection.Status,
		RedirectURL:          session.RedirectURL,
	})
}

func (c ConnectCmd) runWithLocalCallback(ctx context.Context, app *App, p provider.Provider, exchanger enablebanking.SessionExchanger) error {
	if app.NoInput {
		return errors.New("listen is not supported with --no-input")
	}
	if strings.TrimSpace(c.RedirectURL) != "" {
		return errors.New("use either redirect or listen, not both")
	}
	listener, callbackAddress, err := listenLoopback(c.Listen)
	if err != nil {
		return err
	}
	defer listener.Close()
	redirectURL := localCallbackURL(callbackAddress, c.ListenHTTPS)
	callbackTLS := localCallbackTLSConfig{
		Enabled:  c.ListenHTTPS,
		CertPath: c.ListenCert,
		KeyPath:  c.ListenKey,
	}
	if c.ListenHTTPS {
		cert, err := localCallbackCertificate(c.ListenCert, c.ListenKey)
		if err != nil {
			return err
		}
		callbackTLS.Certificate = &cert
	}

	session, err := p.StartConnection(ctx, c.Institution, redirectURL)
	if err != nil {
		return err
	}
	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()
	if _, err := s.UpsertConnection(ctx, session.Connection); err != nil {
		return err
	}
	fmt.Fprintf(app.Stderr, "Open this URL: %s\n", session.RedirectURL)
	fmt.Fprintf(app.Stderr, "Waiting for callback on %s\n", redirectURL)
	if c.ListenHTTPS {
		if c.ListenCert == "" && c.ListenKey == "" {
			fmt.Fprintln(app.Stderr, "The browser may show a warning for the temporary localhost certificate.")
		}
	}

	callback, err := waitForLocalCallback(ctx, listener, session.Connection.ProviderConnectionID, c.CallbackTimeout, callbackTLS)
	if err != nil {
		return err
	}
	report, err := completeAuthorization(ctx, app, p, exchanger, s, callback, c.Institution)
	if err != nil {
		return err
	}
	return app.Out.Write(report)
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

func localCallbackURL(address string, https bool) string {
	scheme := "http"
	if https {
		scheme = "https"
	}
	return (&url.URL{Scheme: scheme, Host: address, Path: "/enablebanking/callback"}).String()
}

type localCallbackTLSConfig struct {
	Enabled     bool
	CertPath    string
	KeyPath     string
	Certificate *tls.Certificate
}

func waitForLocalCallback(ctx context.Context, listener net.Listener, expectedState string, timeout time.Duration, tlsConfig localCallbackTLSConfig) (callbackParams, error) {
	result := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	mux.HandleFunc("/enablebanking/callback", func(w http.ResponseWriter, r *http.Request) {
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

type callbackResult struct {
	callback callbackParams
	err      error
}
