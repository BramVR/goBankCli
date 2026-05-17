package store

import (
	"context"
	"time"

	"gobankcli/internal/provider"
)

func (s *Store) InsertSyncRun(ctx context.Context, run provider.SyncRun) (string, error) {
	id := run.ID
	if id == "" {
		id = stableID("sync_run", run.Provider, run.ConnectionID, run.AccountID, run.StartedAt.Format(time.RFC3339Nano))
	}
	startedAt := timeString(run.StartedAt)
	if startedAt == "" {
		startedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	_, err := s.db.ExecContext(ctx, `
insert into sync_runs(id, provider, connection_id, account_id, started_at, finished_at, status, error, transactions_new, transactions_seen)
values(?,?,?,?,?,?,?,?,?,?)`,
		id, run.Provider, run.ConnectionID, run.AccountID, startedAt, timeStringPtr(run.FinishedAt), run.Status, run.Error, run.TransactionsNew, run.TransactionsSeen)
	if err != nil {
		return "", err
	}
	return id, nil
}
