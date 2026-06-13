package cmd

import (
	"context"
	"errors"
	"time"

	"gobankcli/internal/archive"
	"gobankcli/internal/store"
)

type SyncCmd struct {
	Provider   string `help:"Provider name." default:""`
	Connection string `help:"Provider connection/requisition ID." required:""`
	From       string `help:"Start booking date, inclusive, as YYYY-MM-DD."`
	To         string `help:"End booking date, inclusive, as YYYY-MM-DD."`
}

type syncReport struct {
	ProviderConnectionID string `json:"provider_connection_id"`
	ConnectionID         string `json:"connection_id"`
	Accounts             int    `json:"accounts"`
	TransactionsSeen     int    `json:"transactions_seen"`
}

func (c SyncCmd) Run(ctx context.Context, app *App) error {
	if c.Connection == "" {
		return errors.New("connection is required")
	}
	from, err := parseOptionalDate(c.From, "from")
	if err != nil {
		return err
	}
	to, err := parseOptionalDate(c.To, "to")
	if err != nil {
		return err
	}
	if from != nil && to != nil && from.After(*to) {
		return errors.New("from date must be on or before to date")
	}

	providerName := firstString(c.Provider, app.Config.DefaultProvider)
	p, err := newProvider(providerName)
	if err != nil {
		return err
	}
	providerName = p.Name()

	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()
	manager := archive.NewManager(app.Config, p, s)
	result, err := manager.SyncConnection(ctx, c.Connection, valueOrZero(from), valueOrZero(to))
	if err != nil {
		return err
	}
	localConnectionID := store.LocalConnectionID(providerName, c.Connection)
	return app.Out.Write(syncReport{
		ProviderConnectionID: c.Connection,
		ConnectionID:         localConnectionID,
		Accounts:             result.Accounts,
		TransactionsSeen:     result.TransactionsSeen,
	})
}

func valueOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
