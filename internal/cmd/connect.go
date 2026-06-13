package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gobankcli/internal/authflow"
	"gobankcli/internal/provider"
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
		exchanger, ok := p.(authflow.SessionExchanger)
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

func (c ConnectCmd) runWithLocalCallback(ctx context.Context, app *App, p provider.Provider, exchanger authflow.SessionExchanger) error {
	if app.NoInput {
		return errors.New("listen is not supported with --no-input")
	}
	if strings.TrimSpace(c.RedirectURL) != "" {
		return errors.New("use either redirect or listen, not both")
	}
	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()
	report, err := authflow.Flow{
		Config:        app.Config,
		Provider:      p,
		Exchanger:     exchanger,
		Store:         s,
		InstitutionID: c.Institution,
		ListenAddress: c.Listen,
		HTTPS:         c.ListenHTTPS,
		CertPath:      c.ListenCert,
		KeyPath:       c.ListenKey,
		Timeout:       c.CallbackTimeout,
		Stderr:        app.Stderr,
	}.Run(ctx)
	if err != nil {
		return err
	}
	return app.Out.Write(report)
}
