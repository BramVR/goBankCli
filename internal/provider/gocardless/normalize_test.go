package gocardless

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeInstitutions(t *testing.T) {
	var raw []institutionPayload
	readJSON(t, "institutions_be.json", &raw)
	institutions, err := NormalizeInstitutions(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(institutions) != 2 {
		t.Fatalf("institutions = %d, want 2", len(institutions))
	}
	got := institutions[0]
	if got.Provider != Name || got.ProviderInstitutionID != "BELFIUS_GKCCBEBB" || got.Name != "Belfius" || got.Country != "BE" || got.BIC == "" {
		t.Fatalf("institution = %+v", got)
	}
	if len(got.RawJSON) == 0 || !json.Valid(got.RawJSON) || !containsJSONField(got.RawJSON, "transaction_total_days") {
		t.Fatalf("institution raw JSON not preserved: %s", got.RawJSON)
	}
}

func TestNormalizeAccountDetails(t *testing.T) {
	var raw accountDetailsPayload
	readJSON(t, "account_details.json", &raw)
	account := NormalizeAccountDetails("account-1", "BELFIUS_GKCCBEBB", "req-1", raw)
	if account.Provider != Name || account.ProviderAccountID != "account-1" || account.InstitutionID != "BELFIUS_GKCCBEBB" || account.IBAN != "BE00000000000000" || account.Name != "Main checking" || account.OwnerName != "Test User" {
		t.Fatalf("account = %+v", account)
	}
	if len(account.RawJSON) == 0 {
		t.Fatal("account raw JSON missing")
	}
}

func TestNormalizeTransactions(t *testing.T) {
	var raw transactionsPayload
	readJSON(t, "transactions.json", &raw)
	transactions, err := NormalizeTransactions("account-1", raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 2 {
		t.Fatalf("transactions = %d, want 2 booked transactions", len(transactions))
	}
	got := transactions[0]
	if got.Provider != Name || got.ProviderTransactionID != "2020103000624289-1" || got.AccountID != "account-1" || got.Amount != "45.00" || got.Currency != "EUR" || got.CounterpartyName != "MON MOTHMA" || got.CounterpartyAccount != "GL53SAFI055151515" {
		t.Fatalf("transaction = %+v", got)
	}
	if got.BookingDate.Format("2006-01-02") != "2020-10-30" || got.ValueDate == nil || got.ValueDate.Format("2006-01-02") != "2020-10-30" {
		t.Fatalf("dates = booking %s value %v", got.BookingDate, got.ValueDate)
	}
	if got.Description != "For the support of Restoration of the Republic foundation" || got.RemittanceInfo == "" {
		t.Fatalf("remittance fields = %+v", got)
	}
	if len(got.RawJSON) == 0 {
		t.Fatal("transaction raw JSON missing")
	}
}

func TestNormalizeTransactionDirectionAndStructuredReference(t *testing.T) {
	raw := transactionPayload{
		TransactionID:     "tx-in",
		DebtorName:        "Employer",
		DebtorAccount:     accountRef{IBAN: "BE11111111111111"},
		CreditorName:      "Own Account",
		CreditorAccount:   accountRef{IBAN: "BE00000000000000"},
		BookingDate:       "2026-05-17",
		TransactionAmount: amountPayload{Amount: "100.00", Currency: "EUR"},
		RemittanceInformationStructuredArray: []string{
			"",
			"STRUCTURED-REF",
		},
	}
	tx, err := NormalizeTransaction("account-1", raw)
	if err != nil {
		t.Fatal(err)
	}
	if tx.CounterpartyName != "Employer" || tx.CounterpartyAccount != "BE11111111111111" || tx.Reference != "STRUCTURED-REF" {
		t.Fatalf("incoming transaction = %+v", tx)
	}

	raw.TransactionID = "tx-out"
	raw.TransactionAmount.Amount = "-25.00"
	tx, err = NormalizeTransaction("account-1", raw)
	if err != nil {
		t.Fatal(err)
	}
	if tx.CounterpartyName != "Own Account" || tx.CounterpartyAccount != "BE00000000000000" {
		t.Fatalf("outgoing transaction = %+v", tx)
	}
}

func TestNormalizeTransactionUsesDateTimeFallback(t *testing.T) {
	raw := transactionPayload{
		BookingDateTime:   "2026-05-17T10:11:12Z",
		ValueDateTime:     "2026-05-18T00:00:00Z",
		TransactionAmount: amountPayload{Amount: "-25.00", Currency: "EUR"},
	}
	tx, err := NormalizeTransaction("account-1", raw)
	if err != nil {
		t.Fatal(err)
	}
	if tx.BookingDate.Format("2006-01-02") != "2026-05-17" || tx.ValueDate == nil || tx.ValueDate.Format("2006-01-02") != "2026-05-18" {
		t.Fatalf("datetime fallback dates = booking %s value %v", tx.BookingDate, tx.ValueDate)
	}
}

func TestNormalizeTransactionDropsPlaceholderReference(t *testing.T) {
	raw := transactionPayload{
		BookingDate:       "2026-05-17",
		EndToEndID:        "NOTPROVIDED",
		TransactionAmount: amountPayload{Amount: "-25.00", Currency: "EUR"},
	}
	tx, err := NormalizeTransaction("account-1", raw)
	if err != nil {
		t.Fatal(err)
	}
	if tx.Reference != "" {
		t.Fatalf("reference = %q, want empty placeholder", tx.Reference)
	}
}

func containsJSONField(raw []byte, field string) bool {
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return false
	}
	_, ok := fields[field]
	return ok
}

func readJSON(t *testing.T, name string, dst any) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatal(err)
	}
}
