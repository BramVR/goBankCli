package cmd

import (
	"context"
	"errors"

	"gobankcli/internal/store"
)

type ConnectCmd struct {
	Provider    string `help:"Provider name." default:""`
	Institution string `help:"Provider institution ID." required:""`
	RedirectURL string `name:"redirect" help:"Redirect URL registered with the provider." required:""`
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
