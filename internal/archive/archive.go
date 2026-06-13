package archive

import (
	"context"
	"errors"
	"strings"

	"gobankcli/internal/config"
	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

type Manager struct {
	Config   config.Config
	Provider provider.Provider
	Store    *store.Store
}

type AccountsResult struct {
	Accounts []provider.Account
	Fresh    bool
}

func NewManager(cfg config.Config, p provider.Provider, s *store.Store) *Manager {
	return &Manager{Config: cfg, Provider: p, Store: s}
}

func (m *Manager) AccountsForConnection(ctx context.Context, providerConnectionID string) (AccountsResult, error) {
	providerName := m.Provider.Name()
	accounts, err := m.Provider.ListAccounts(ctx, providerConnectionID)
	if err == nil {
		return AccountsResult{Accounts: accounts, Fresh: true}, nil
	}
	if !errors.Is(err, provider.ErrMissingStableAccountID) {
		return AccountsResult{}, err
	}
	stored, storedErr := m.Store.AccountsByConnection(ctx, store.LocalConnectionID(providerName, providerConnectionID))
	if storedErr != nil {
		return AccountsResult{}, storedErr
	}
	if len(stored) == 0 {
		return AccountsResult{}, err
	}
	return AccountsResult{Accounts: stored}, nil
}

func (m *Manager) ArchiveAccounts(ctx context.Context, localConnectionID string, accounts []provider.Account) ([]provider.Account, error) {
	archived := make([]provider.Account, len(accounts))
	copy(archived, accounts)
	archivedInstitutions := map[string]bool{}
	for i := range archived {
		institutionID := archived[i].InstitutionID
		if !archivedInstitutions[institutionID] {
			if err := m.ArchiveInstitutionByID(ctx, institutionID); err != nil {
				return nil, err
			}
			archivedInstitutions[institutionID] = true
		}
		archived[i].ConnectionID = localConnectionID
		id, err := m.Store.UpsertAccount(ctx, archived[i])
		if err != nil {
			return nil, err
		}
		archived[i].ID = id
	}
	return archived, nil
}

func (m *Manager) ArchiveConnectionAccounts(ctx context.Context, localConnectionID string, connection provider.Connection, accounts []provider.Account) ([]provider.Account, error) {
	archived := make([]provider.Account, len(accounts))
	copy(archived, accounts)
	institutionID := strings.TrimSpace(connection.InstitutionID)
	for i := range archived {
		if strings.TrimSpace(archived[i].InstitutionID) == "" {
			archived[i].InstitutionID = institutionID
		}
	}
	return m.ArchiveAccounts(ctx, localConnectionID, archived)
}

func (m *Manager) ArchiveInstitutionByID(ctx context.Context, providerInstitutionID string) error {
	if providerInstitutionID == "" {
		return nil
	}
	providerName := m.Provider.Name()
	exists, err := m.Store.HasInstitution(ctx, providerName, providerInstitutionID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	for _, country := range archiveCountries(m.Config, providerName, providerInstitutionID) {
		institutions, err := m.Provider.ListInstitutions(ctx, country)
		if err != nil {
			return err
		}
		for _, institution := range institutions {
			if institution.ProviderInstitutionID != providerInstitutionID {
				continue
			}
			_, err := m.Store.UpsertInstitution(ctx, institution)
			return err
		}
	}
	return nil
}

func archiveCountries(cfg config.Config, providerName, providerInstitutionID string) []string {
	var countries []string
	addCountry := func(country string) {
		country = strings.ToUpper(strings.TrimSpace(country))
		if country == "" {
			return
		}
		for _, existing := range countries {
			if existing == country {
				return
			}
		}
		countries = append(countries, country)
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	for _, connection := range cfg.Connections {
		if !sameConfigProvider(connection.Provider, providerName) || strings.TrimSpace(connection.InstitutionID) != providerInstitutionID {
			continue
		}
		addCountry(connection.Country)
	}
	addCountry(cfg.DefaultCountry)
	for _, connection := range cfg.Connections {
		if !sameConfigProvider(connection.Provider, providerName) {
			continue
		}
		addCountry(connection.Country)
	}
	return countries
}

func sameConfigProvider(configProvider, providerName string) bool {
	configProvider = strings.ToLower(strings.TrimSpace(configProvider))
	return configProvider == "" || configProvider == providerName
}
