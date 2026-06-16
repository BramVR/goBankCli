package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gobankcli/internal/provider"
)

func TestTransactionDedupeKeyStableHash(t *testing.T) {
	valueDate := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	tx := provider.Transaction{
		Provider:            "gocardless",
		AccountID:           "acct_1",
		BookingDate:         time.Date(2026, 5, 17, 12, 30, 0, 0, time.UTC),
		ValueDate:           &valueDate,
		Amount:              "-12.34",
		Currency:            "EUR",
		CounterpartyName:    "Example BV",
		CounterpartyAccount: "BE11111111111111",
		Description:         "Card payment",
		RemittanceInfo:      "structured info",
	}
	first := TransactionDedupeKey(tx)
	second := TransactionDedupeKey(tx)
	if first == "" || first != second {
		t.Fatalf("dedupe key not stable: %q %q", first, second)
	}
	tx.Description = "Other"
	if changed := TransactionDedupeKey(tx); changed == first {
		t.Fatalf("dedupe key should change when fallback identity changes")
	}
}

func TestTransactionDedupeKeyUsesFallbackDifferentiators(t *testing.T) {
	tx := provider.Transaction{
		Provider:         "gocardless",
		AccountID:        "acct_1",
		BookingDate:      time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		Amount:           "-12.34",
		Currency:         "EUR",
		CounterpartyName: "Example BV",
		Description:      "Payment",
	}
	first := TransactionDedupeKey(tx)
	tx.CounterpartyAccount = "BE11111111111111"
	if changed := TransactionDedupeKey(tx); changed == first {
		t.Fatal("dedupe key should include counterparty account")
	}
	tx.CounterpartyAccount = ""
	tx.RemittanceInfo = "invoice 1"
	if changed := TransactionDedupeKey(tx); changed == first {
		t.Fatal("dedupe key should include remittance info")
	}
}

func TestTransactionDedupeKeyScopesProviderIDToAccount(t *testing.T) {
	tx := provider.Transaction{
		Provider:              "gocardless",
		AccountID:             "acct_1",
		ProviderTransactionID: "same",
	}
	first := TransactionDedupeKey(tx)
	tx.AccountID = "acct_2"
	second := TransactionDedupeKey(tx)
	if first == second {
		t.Fatalf("provider transaction dedupe key should include account: %q", first)
	}
}

func TestOpenEscapesSQLitePath(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "archive?with#delims.db")
	s, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected sqlite db at requested path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "archive")); !os.IsNotExist(err) {
		t.Fatalf("unexpected unescaped sqlite path exists or stat failed: %v", err)
	}
}

func TestOpenRelativeSQLitePath(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	}()

	dbPath := "relative?with#delims.db"
	s, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, dbPath)); err != nil {
		t.Fatalf("expected sqlite db at requested relative path: %v", err)
	}
}

func TestOpenRestrictsDBPermissions(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "gobankcli.db")
	s, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("db permissions = %o, want 600", got)
	}
}

func TestOpenRefusesNewerSchemaVersion(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "gobankcli.db")
	db, err := sql.Open("sqlite", sqliteDSN(dbPath))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "pragma user_version = 99"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = Open(ctx, dbPath)
	if err == nil || !strings.Contains(err.Error(), "newer than supported") {
		t.Fatalf("Open error = %v, want newer schema error", err)
	}

	db, err = sql.Open("sqlite", sqliteDSN(dbPath))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var version int
	if err := db.QueryRowContext(ctx, "pragma user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 99 {
		t.Fatalf("user_version = %d, want 99", version)
	}
}

func TestStoreUpsertTransactionDeduplicates(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	account := provider.Account{
		Provider:          "gocardless",
		ProviderAccountID: "acct_provider",
		IBAN:              "BE00000000000000",
		Currency:          "EUR",
	}
	accountID, err := s.UpsertAccount(ctx, account)
	if err != nil {
		t.Fatal(err)
	}

	tx := provider.Transaction{
		Provider:              "gocardless",
		ProviderTransactionID: "tx_provider",
		AccountID:             accountID,
		BookingDate:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		Amount:                "42.00",
		Currency:              "EUR",
		Description:           "Initial",
	}
	firstID, err := s.UpsertTransaction(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Description = "Updated"
	secondID, err := s.UpsertTransaction(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if firstID != secondID {
		t.Fatalf("upsert ids differ: %q %q", firstID, secondID)
	}
	status, err := s.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status.Accounts != 1 || status.Transactions != 1 {
		t.Fatalf("status = %+v", status)
	}
}

func TestStoreRejectsTransactionWithoutArchivedAccount(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	baseTx := provider.Transaction{
		Provider:              "gocardless",
		ProviderTransactionID: "tx_provider",
		BookingDate:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		Amount:                "42.00",
		Currency:              "EUR",
	}
	for _, tc := range []struct {
		name      string
		accountID string
	}{
		{name: "blank", accountID: ""},
		{name: "unknown", accountID: "acct_missing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := baseTx
			tx.AccountID = tc.accountID
			if _, err := s.UpsertTransaction(ctx, tx); err == nil || !strings.Contains(err.Error(), "account") {
				t.Fatalf("UpsertTransaction error = %v, want missing account error", err)
			}
		})
	}
	status, err := s.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status.Transactions != 0 {
		t.Fatalf("transactions = %d, want 0", status.Transactions)
	}
}

func TestStoreUpsertAccountReturnsPersistedID(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	firstID, err := s.UpsertAccount(ctx, provider.Account{
		ID:                "stored",
		Provider:          "gocardless",
		ProviderAccountID: "acct_provider",
	})
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := s.UpsertAccount(ctx, provider.Account{
		ID:                "ignored",
		Provider:          "gocardless",
		ProviderAccountID: "acct_provider",
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstID != "stored" || secondID != "stored" {
		t.Fatalf("ids = %q, %q; want persisted id", firstID, secondID)
	}
}

func TestStoreUpsertAccountKeepsStableIDAndUpdatesResourceID(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	firstID, err := s.UpsertAccount(ctx, provider.Account{
		Provider:           "enablebanking",
		ProviderAccountID:  "stable-hash",
		ProviderResourceID: "session-uid-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := s.UpsertAccount(ctx, provider.Account{
		Provider:           "enablebanking",
		ProviderAccountID:  "stable-hash",
		ProviderResourceID: "session-uid-2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstID != secondID {
		t.Fatalf("ids = %q, %q; want same stable account", firstID, secondID)
	}
	var resourceID string
	if err := s.db.QueryRowContext(ctx, `select provider_resource_id from accounts where id = ?`, firstID).Scan(&resourceID); err != nil {
		t.Fatal(err)
	}
	if resourceID != "session-uid-2" {
		t.Fatalf("provider_resource_id = %q, want latest session UID", resourceID)
	}
}

func TestStoreMigratesV1AccountsProviderResourceID(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "gobankcli.db")
	db, err := sql.Open("sqlite", sqliteDSN(dbPath))
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx, `
create table accounts (
	id text primary key,
	provider text not null,
	provider_account_id text not null,
	institution_id text,
	connection_id text,
	iban text,
	name text,
	currency text,
	owner_name text,
	raw_json blob,
	updated_at text not null,
	unique(provider, provider_account_id)
);
insert into accounts(id, provider, provider_account_id, updated_at)
values('account-local', 'gocardless', 'account-provider', '2026-05-17T00:00:00Z');
pragma user_version = 1;
`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	s, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var resourceID string
	if err := s.db.QueryRowContext(ctx, `select provider_resource_id from accounts where id = 'account-local'`).Scan(&resourceID); err != nil {
		t.Fatal(err)
	}
	if resourceID != "account-provider" {
		t.Fatalf("provider_resource_id = %q, want provider account id", resourceID)
	}
	var version int
	if err := s.db.QueryRowContext(ctx, "pragma user_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != schemaVersion {
		t.Fatalf("schema version = %d, want %d", version, schemaVersion)
	}
}

func TestStoreUpsertTransactionReturnsPersistedID(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := s.UpsertAccount(ctx, provider.Account{
		ID:                "acct_1",
		Provider:          "gocardless",
		ProviderAccountID: "acct_provider",
	}); err != nil {
		t.Fatal(err)
	}
	tx := provider.Transaction{
		ID:                    "stored",
		Provider:              "gocardless",
		ProviderTransactionID: "tx_provider",
		AccountID:             "acct_1",
		BookingDate:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		Amount:                "42.00",
		Currency:              "EUR",
	}
	first, err := s.UpsertTransactionResult(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.ID = "ignored"
	second, err := s.UpsertTransactionResult(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "stored" || second.ID != "stored" {
		t.Fatalf("ids = %q, %q; want persisted id", first.ID, second.ID)
	}
	if !first.Inserted || second.Inserted {
		t.Fatalf("inserted flags = %v, %v; want true, false", first.Inserted, second.Inserted)
	}
}

func TestStoreUpsertConnectionAndSyncRun(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	firstID, err := s.UpsertConnection(ctx, provider.Connection{
		ID:                   "connection_local",
		Provider:             "gocardless",
		ProviderConnectionID: "req_provider",
		InstitutionID:        "BELFIUS_GKCCBEBB",
		Status:               "linked",
		RedirectURL:          "https://example.test/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := s.UpsertConnection(ctx, provider.Connection{
		ID:                   "ignored",
		Provider:             "gocardless",
		ProviderConnectionID: "req_provider",
		InstitutionID:        "BELFIUS_GKCCBEBB",
		Status:               "expired",
		RedirectURL:          "https://example.test/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstID != "connection_local" || secondID != "connection_local" {
		t.Fatalf("connection ids = %q, %q; want persisted id", firstID, secondID)
	}

	finished := time.Date(2026, 5, 17, 13, 0, 0, 0, time.UTC)
	syncRunID, err := s.InsertSyncRun(ctx, provider.SyncRun{
		ID:               "sync_local",
		Provider:         "gocardless",
		ConnectionID:     firstID,
		AccountID:        "account_local",
		StartedAt:        finished.Add(-time.Minute),
		FinishedAt:       &finished,
		Status:           "ok",
		TransactionsSeen: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if syncRunID != "sync_local" {
		t.Fatalf("sync run id = %q, want sync_local", syncRunID)
	}
	status, err := s.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status.Connections != 1 || status.SyncRuns != 1 {
		t.Fatalf("status = %+v", status)
	}
}

func TestStoreAllowsSameProviderTransactionIDAcrossAccounts(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	for _, accountID := range []string{"acct_1", "acct_2"} {
		if _, err := s.UpsertAccount(ctx, provider.Account{
			ID:                accountID,
			Provider:          "gocardless",
			ProviderAccountID: accountID + "_provider",
		}); err != nil {
			t.Fatal(err)
		}
		_, err := s.UpsertTransaction(ctx, provider.Transaction{
			Provider:              "gocardless",
			ProviderTransactionID: "same-provider-id",
			AccountID:             accountID,
			BookingDate:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
			Amount:                "1.00",
			Currency:              "EUR",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	status, err := s.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if status.Transactions != 2 {
		t.Fatalf("transactions = %d, want 2", status.Transactions)
	}
}

func TestStoreQueryReadOnlySelect(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	result, err := s.Query(ctx, "select 'ok' as value, 2 as value;")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Columns) != 2 || result.Columns[0] != "value" || result.Columns[1] != "value" || len(result.Rows) != 1 || result.Rows[0][0] != "ok" || result.Rows[0][1] != "2" {
		t.Fatalf("query result = %+v", result)
	}
	empty, err := s.Query(ctx, "select 'x' as value where 0")
	if err != nil {
		t.Fatal(err)
	}
	if empty.Rows == nil || len(empty.Rows) != 0 {
		t.Fatalf("empty query rows = %#v, want empty non-nil slice", empty.Rows)
	}
	if _, err := s.Query(ctx, "insert into accounts(id, provider, provider_account_id, updated_at) values('a','p','pa','now')"); err == nil {
		t.Fatal("write query should be rejected")
	}
	if _, err := s.UpsertAccount(ctx, provider.Account{Provider: "gocardless", ProviderAccountID: "acct_after_query"}); err != nil {
		t.Fatalf("store should allow writes after query: %v", err)
	}
}

func TestListTransactionsForCSVExport(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = s.UpsertInstitution(ctx, provider.Institution{
		Provider:              "gocardless",
		ProviderInstitutionID: "BELFIUS_GKCCBEBB",
		Name:                  "Belfius",
		Country:               "BE",
	})
	if err != nil {
		t.Fatal(err)
	}
	accountID, err := s.UpsertAccount(ctx, provider.Account{
		Provider:          "gocardless",
		ProviderAccountID: "acct_provider",
		InstitutionID:     "BELFIUS_GKCCBEBB",
		IBAN:              "BE00000000000000",
		Currency:          "EUR",
	})
	if err != nil {
		t.Fatal(err)
	}
	valueDate := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	_, err = s.UpsertTransaction(ctx, provider.Transaction{
		ID:                    "tx_export",
		Provider:              "gocardless",
		ProviderTransactionID: "tx_provider",
		AccountID:             accountID,
		BookingDate:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		ValueDate:             &valueDate,
		Amount:                "-12.34",
		Currency:              "EUR",
		CounterpartyName:      "Example BV",
		CounterpartyAccount:   "BE11111111111111",
		Description:           "Invoice",
	})
	if err != nil {
		t.Fatal(err)
	}
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	rows, err := s.ListTransactions(ctx, TransactionFilter{From: &from, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.Date != "2026-05-17" || row.ValueDate != "2026-05-18" || row.IBAN != "BE00000000000000" || row.Institution != "Belfius" || row.TransactionID != "tx_export" {
		t.Fatalf("export row = %+v", row)
	}
}
