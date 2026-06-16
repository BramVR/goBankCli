package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gobankcli/internal/provider"
)

type TransactionUpsertResult struct {
	ID       string
	Inserted bool
}

func TransactionDedupeKey(tx provider.Transaction) string {
	switch {
	case strings.TrimSpace(tx.ProviderTransactionID) != "":
		return "provider_transaction_id:" + tx.Provider + ":" + tx.AccountID + ":" + strings.TrimSpace(tx.ProviderTransactionID)
	case strings.TrimSpace(tx.Reference) != "":
		return "reference:" + tx.Provider + ":" + tx.AccountID + ":" + strings.TrimSpace(tx.Reference)
	default:
		return "hash:" + stableID(
			tx.Provider,
			tx.AccountID,
			tx.BookingDate.Format("2006-01-02"),
			dateString(tx.ValueDate),
			tx.Amount,
			tx.Currency,
			tx.CounterpartyName,
			tx.CounterpartyAccount,
			tx.Description,
			tx.RemittanceInfo,
		)
	}
}

func (s *Store) UpsertTransaction(ctx context.Context, tx provider.Transaction) (string, error) {
	result, err := s.UpsertTransactionResult(ctx, tx)
	if err != nil {
		return "", err
	}
	return result.ID, nil
}

func (s *Store) UpsertTransactionResult(ctx context.Context, tx provider.Transaction) (TransactionUpsertResult, error) {
	if strings.TrimSpace(tx.AccountID) == "" {
		return TransactionUpsertResult{}, fmt.Errorf("transaction account id is required")
	}
	if err := s.requireAccount(ctx, tx.AccountID); err != nil {
		return TransactionUpsertResult{}, err
	}
	dedupeKey := TransactionDedupeKey(tx)
	id := tx.ID
	if id == "" {
		id = stableID("transaction", dedupeKey)
	}
	inserted, err := s.transactionMissing(ctx, dedupeKey)
	if err != nil {
		return TransactionUpsertResult{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	bookingDate := tx.BookingDate.Format("2006-01-02")
	valueDate := dateString(tx.ValueDate)

	_, err = s.db.ExecContext(ctx, `
insert into transactions(id, dedupe_key, provider, provider_transaction_id, account_id, booking_date, value_date, amount, currency, counterparty_name, counterparty_account, description, remittance_info, reference, raw_json, created_at, updated_at)
values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
on conflict(dedupe_key) do update set
	provider_transaction_id=excluded.provider_transaction_id,
	booking_date=excluded.booking_date,
	value_date=excluded.value_date,
	amount=excluded.amount,
	currency=excluded.currency,
	counterparty_name=excluded.counterparty_name,
	counterparty_account=excluded.counterparty_account,
	description=excluded.description,
	remittance_info=excluded.remittance_info,
	reference=excluded.reference,
	raw_json=excluded.raw_json,
	updated_at=excluded.updated_at`,
		id, dedupeKey, tx.Provider, tx.ProviderTransactionID, tx.AccountID, bookingDate, valueDate, tx.Amount, tx.Currency, tx.CounterpartyName, tx.CounterpartyAccount, tx.Description, tx.RemittanceInfo, tx.Reference, tx.RawJSON, now, now)
	if err != nil {
		return TransactionUpsertResult{}, err
	}
	persistedID, err := s.transactionID(ctx, dedupeKey)
	if err != nil {
		return TransactionUpsertResult{}, err
	}
	return TransactionUpsertResult{ID: persistedID, Inserted: inserted}, nil
}

func (s *Store) requireAccount(ctx context.Context, accountID string) error {
	var id string
	err := s.db.QueryRowContext(ctx, `select id from accounts where id = ?`, accountID).Scan(&id)
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return fmt.Errorf("transaction account %q is not archived", accountID)
	}
	return err
}

func dateString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func (s *Store) transactionID(ctx context.Context, dedupeKey string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `select id from transactions where dedupe_key = ?`, dedupeKey).Scan(&id)
	return id, err
}

func (s *Store) transactionMissing(ctx context.Context, dedupeKey string) (bool, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `select id from transactions where dedupe_key = ?`, dedupeKey).Scan(&id)
	if err == nil {
		return false, nil
	}
	if err == sql.ErrNoRows {
		return true, nil
	}
	return false, err
}
