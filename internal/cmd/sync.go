package cmd

import (
	"context"
	"errors"
	"time"

	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

type SyncCmd struct {
	Provider   string `help:"Provider name." default:""`
	Connection string `help:"Provider connection/requisition ID." required:""`
	From       string `help:"Start booking date, inclusive, as YYYY-MM-DD."`
	To         string `help:"End booking date, inclusive, as YYYY-MM-DD."`
}

type syncReport struct {
	ConnectionID     string `json:"connection_id"`
	Accounts         int    `json:"accounts"`
	TransactionsSeen int    `json:"transactions_seen"`
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
	accounts, err := p.ListAccounts(ctx, c.Connection)
	if err != nil {
		return err
	}

	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	seen := 0
	localConnectionID := store.LocalConnectionID(providerName, c.Connection)
	archivedInstitutions := map[string]bool{}
	for _, account := range accounts {
		started := time.Now().UTC()
		providerAccountID := account.ProviderAccountID
		if !archivedInstitutions[account.InstitutionID] {
			if err := archiveInstitutionByID(ctx, p, s, app.Config.DefaultCountry, account.InstitutionID); err != nil {
				return err
			}
			archivedInstitutions[account.InstitutionID] = true
		}
		account.ConnectionID = localConnectionID
		localID, err := s.UpsertAccount(ctx, account)
		if err != nil {
			return err
		}
		transactions, err := p.FetchTransactions(ctx, providerAccountID, valueOrZero(from), valueOrZero(to))
		if err != nil {
			return err
		}
		newCount := 0
		for _, tx := range transactions {
			tx.AccountID = localID
			result, err := s.UpsertTransactionResult(ctx, tx)
			if err != nil {
				return err
			}
			if result.Inserted {
				newCount++
			}
			seen++
		}
		finished := time.Now().UTC()
		if _, err := s.InsertSyncRun(ctx, provider.SyncRun{
			Provider:         providerName,
			ConnectionID:     localConnectionID,
			AccountID:        localID,
			StartedAt:        started,
			FinishedAt:       &finished,
			Status:           "ok",
			TransactionsNew:  int64(newCount),
			TransactionsSeen: int64(len(transactions)),
		}); err != nil {
			return err
		}
	}
	return app.Out.Write(syncReport{ConnectionID: c.Connection, Accounts: len(accounts), TransactionsSeen: seen})
}

func valueOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
