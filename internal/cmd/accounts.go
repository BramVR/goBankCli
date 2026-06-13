package cmd

import (
	"context"
	"errors"

	"gobankcli/internal/archive"
	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

type AccountsCmd struct {
	Provider   string `help:"Provider name." default:""`
	Connection string `help:"Provider connection/requisition ID." required:""`
}

type accountsReport struct {
	Accounts []accountReport `json:"accounts"`
	Count    int             `json:"count"`
}

type accountReport struct {
	ID                 string `json:"id"`
	Provider           string `json:"provider"`
	ProviderAccountID  string `json:"provider_account_id"`
	ProviderResourceID string `json:"provider_resource_id"`
	InstitutionID      string `json:"institution_id"`
	ConnectionID       string `json:"connection_id"`
	IBAN               string `json:"iban"`
	Name               string `json:"name"`
	Currency           string `json:"currency"`
	OwnerName          string `json:"owner_name"`
}

func (c AccountsCmd) Run(ctx context.Context, app *App) error {
	if c.Connection == "" {
		return errors.New("connection is required")
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
	localConnectionID := store.LocalConnectionID(providerName, c.Connection)
	manager := archive.NewManager(app.Config, p, s)
	result, err := manager.AccountsForConnection(ctx, c.Connection)
	if err != nil {
		return err
	}
	accounts := result.Accounts
	if !result.Fresh {
		return app.Out.Write(accountsReport{Accounts: accountReports(accounts), Count: len(accounts)})
	}
	accounts, err = manager.ArchiveAccounts(ctx, localConnectionID, accounts)
	if err != nil {
		return err
	}
	return app.Out.Write(accountsReport{Accounts: accountReports(accounts), Count: len(accounts)})
}

func accountReports(accounts []provider.Account) []accountReport {
	reports := make([]accountReport, 0, len(accounts))
	for _, account := range accounts {
		reports = append(reports, accountReport{
			ID:                 account.ID,
			Provider:           account.Provider,
			ProviderAccountID:  account.ProviderAccountID,
			ProviderResourceID: account.ProviderResourceID,
			InstitutionID:      account.InstitutionID,
			ConnectionID:       account.ConnectionID,
			IBAN:               account.IBAN,
			Name:               account.Name,
			Currency:           account.Currency,
			OwnerName:          account.OwnerName,
		})
	}
	return reports
}
