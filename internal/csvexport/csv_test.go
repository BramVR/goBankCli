package csvexport

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteStableNormalizedCSV(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, []Row{{
		Date:                "2026-05-17",
		ValueDate:           "2026-05-18",
		AccountID:           "acct_1",
		IBAN:                "BE00000000000000",
		Institution:         "Belfius",
		CounterpartyName:    "Example BV",
		CounterpartyAccount: "BE11111111111111",
		Description:         "Invoice, May",
		Amount:              "-12.34",
		Currency:            "EUR",
		TransactionID:       "tx_1",
		Provider:            "gocardless",
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	wantHeader := "date,value_date,account_id,iban,institution,counterparty_name,counterparty_account,description,amount,currency,transaction_id,provider,category\n"
	if !strings.HasPrefix(got, wantHeader) {
		t.Fatalf("CSV header = %q", got)
	}
	if !strings.Contains(got, `"Invoice, May"`) {
		t.Fatalf("CSV should quote comma-containing descriptions:\n%s", got)
	}
	if !strings.HasSuffix(strings.TrimSpace(got), ",gocardless,") {
		t.Fatalf("CSV should keep empty category column:\n%s", got)
	}
}
