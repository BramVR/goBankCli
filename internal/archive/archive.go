package archive

import (
	"context"
	"errors"
	"strings"
	"time"

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

type SyncResult struct {
	Accounts             int
	TransactionsInserted int
	TransactionsSeen     int
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

func (m *Manager) SyncConnection(ctx context.Context, providerConnectionID string, from, to time.Time) (SyncResult, error) {
	providerName := m.Provider.Name()
	localConnectionID := store.LocalConnectionID(providerName, providerConnectionID)
	result, err := m.AccountsForConnection(ctx, providerConnectionID)
	if err != nil {
		return SyncResult{}, err
	}
	accounts := result.Accounts
	if result.Fresh {
		accounts, err = m.ArchiveAccounts(ctx, localConnectionID, accounts)
		if err != nil {
			return SyncResult{}, err
		}
	}

	syncResult := SyncResult{Accounts: len(accounts)}
	for _, account := range accounts {
		if account.ID == "" {
			return SyncResult{}, errors.New("stored account missing local id")
		}
		started := time.Now().UTC()
		transactions, err := m.Provider.FetchTransactions(ctx, transactionFetchAccountID(account), from, to)
		if err != nil {
			return SyncResult{}, err
		}
		inserted := 0
		for _, tx := range transactions {
			tx.AccountID = account.ID
			upsert, err := m.Store.UpsertTransactionResult(ctx, tx)
			if err != nil {
				return SyncResult{}, err
			}
			if upsert.Inserted {
				inserted++
			}
		}
		finished := time.Now().UTC()
		if _, err := m.Store.InsertSyncRun(ctx, provider.SyncRun{
			Provider:         providerName,
			ConnectionID:     localConnectionID,
			AccountID:        account.ID,
			StartedAt:        started,
			FinishedAt:       &finished,
			Status:           "ok",
			TransactionsNew:  int64(inserted),
			TransactionsSeen: int64(len(transactions)),
		}); err != nil {
			return SyncResult{}, err
		}
		syncResult.TransactionsInserted += inserted
		syncResult.TransactionsSeen += len(transactions)
	}
	return syncResult, nil
}

func transactionFetchAccountID(account provider.Account) string {
	if account.ProviderResourceID != "" {
		return account.ProviderResourceID
	}
	return account.ProviderAccountID
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
