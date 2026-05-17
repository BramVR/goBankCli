package cmd

import (
	"context"
	"errors"

	"gobankcli/internal/provider"
	"gobankcli/internal/provider/enablebanking"
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
	accounts, fresh, err := accountsForConnection(ctx, p, s, providerName, c.Connection)
	if err != nil {
		return err
	}
	if !fresh {
		return app.Out.Write(accountsReport{Accounts: accountReports(accounts), Count: len(accounts)})
	}
	archivedInstitutions := map[string]bool{}
	for i := range accounts {
		if !archivedInstitutions[accounts[i].InstitutionID] {
			countries := institutionArchiveCountries(app.Config, providerName, accounts[i].InstitutionID)
			if err := archiveInstitutionByID(ctx, p, s, countries, accounts[i].InstitutionID); err != nil {
				return err
			}
			archivedInstitutions[accounts[i].InstitutionID] = true
		}
		accounts[i].ConnectionID = localConnectionID
		id, err := s.UpsertAccount(ctx, accounts[i])
		if err != nil {
			return err
		}
		accounts[i].ID = id
	}
	return app.Out.Write(accountsReport{Accounts: accountReports(accounts), Count: len(accounts)})
}

func accountsForConnection(ctx context.Context, p provider.Provider, s *store.Store, providerName, providerConnectionID string) ([]provider.Account, bool, error) {
	accounts, err := p.ListAccounts(ctx, providerConnectionID)
	if err == nil {
		return accounts, true, nil
	}
	if providerName != enablebanking.Name || !errors.Is(err, enablebanking.ErrMissingStableAccountID) {
		return nil, false, err
	}
	stored, storedErr := s.AccountsByConnection(ctx, store.LocalConnectionID(providerName, providerConnectionID))
	if storedErr != nil {
		return nil, false, storedErr
	}
	if len(stored) == 0 {
		return nil, false, err
	}
	return stored, false, nil
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
