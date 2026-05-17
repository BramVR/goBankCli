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

func TestStoreUpsertTransactionReturnsPersistedID(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	tx := provider.Transaction{
		ID:                    "stored",
		Provider:              "gocardless",
		ProviderTransactionID: "tx_provider",
		AccountID:             "acct_1",
		BookingDate:           time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC),
		Amount:                "42.00",
		Currency:              "EUR",
	}
	firstID, err := s.UpsertTransaction(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.ID = "ignored"
	secondID, err := s.UpsertTransaction(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if firstID != "stored" || secondID != "stored" {
		t.Fatalf("ids = %q, %q; want persisted id", firstID, secondID)
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
