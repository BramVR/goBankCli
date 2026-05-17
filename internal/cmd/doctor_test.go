package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobankcli/internal/config"
	"gobankcli/internal/provider"
)

func TestDoctorJSONMissingCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--json", "doctor"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	for _, want := range []string{`"config_exists": false`, `"gocardless_secret_id": "missing"`, `"gocardless_secret_key": "missing"`, `"gocardless_configured": false`} {
		if !strings.Contains(got, want) {
			t.Fatalf("doctor output missing %s:\n%s", want, got)
		}
	}
	if strings.Contains(got, " (missing)") {
		t.Fatalf("doctor JSON path should not be annotated:\n%s", got)
	}
	if strings.Contains(got, "secret") && strings.Contains(got, "id\n") {
		t.Fatalf("doctor output appears to print secret data:\n%s", got)
	}
}

func TestVersionWithoutCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--version"}, "test-version", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "test-version" {
		t.Fatalf("version output = %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestInitForceOverwritesInvalidConfig(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, ".config", "gobankcli", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("not = [toml\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--config", configPath, "init", "--force"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `default_provider = 'gocardless'`) {
		t.Fatalf("config was not overwritten with defaults:\n%s", b)
	}
}

func TestStatusPlainOmitsEmptyMessage(t *testing.T) {
	home := t.TempDir()
	dbPath := filepath.Join(home, "archive.db")
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "--plain", "status"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	if !strings.Contains(got, "archive_open: true") {
		t.Fatalf("status output missing archive_open:\n%s", got)
	}
	if strings.Contains(got, "message:") {
		t.Fatalf("status output should not include empty message:\n%s", got)
	}
}

func TestExportWritesDefaultCSV(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"export"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	exportPath := filepath.Join(home, "Finance", "gobankcli", "exports", "normalized.csv")
	b, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "date,value_date,account_id,iban,institution,counterparty_name,counterparty_account,description,amount,currency,transaction_id,provider,category\n") {
		t.Fatalf("export CSV header mismatch:\n%s", b)
	}
	if !strings.Contains(stdout.String(), "rows: 0") {
		t.Fatalf("export report missing row count:\n%s", stdout.String())
	}
}

func TestQueryPlainOutputsTSV(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "archive.db")
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "--plain", "query", "select ';' as separator, 1 as one"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "separator\tone\n;\t1\n" {
		t.Fatalf("query output = %q", got)
	}
}

func TestQueryRejectsWriteSQL(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "archive.db")
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--db", dbPath, "sql", "delete from accounts"}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "only read-only select queries are allowed") {
		t.Fatalf("Run error = %v, want read-only rejection", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestInstitutionsRejectsUnsupportedProvider(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"institutions", "--provider", "unknown", "--country", "BE"}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "unsupported provider: unknown") {
		t.Fatalf("Run error = %v, want unsupported provider", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestInstitutionsMissingGoCardlessCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(config.EnvGoCardlessSecretID, "")
	t.Setenv(config.EnvGoCardlessSecretKey, "")
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"institutions", "--country", "BE"}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "gocardless credentials missing") {
		t.Fatalf("Run error = %v, want missing credentials", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestAccountReportsOmitRawJSON(t *testing.T) {
	report := accountsReport{Accounts: accountReports([]provider.Account{{
		ID:                "account_local",
		Provider:          "gocardless",
		ProviderAccountID: "account_provider",
		RawJSON:           []byte(`{"ownerName":"Sensitive"}`),
	}})}
	b, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if strings.Contains(got, "RawJSON") || strings.Contains(got, "Sensitive") {
		t.Fatalf("account report leaked raw JSON: %s", got)
	}
}

func TestInstitutionReportsOmitRawJSON(t *testing.T) {
	report := institutionReports([]provider.Institution{{
		Provider:              "gocardless",
		ProviderInstitutionID: "BELFIUS_GKCCBEBB",
		Name:                  "Belfius",
		RawJSON:               []byte(`{"logo":"raw"}`),
	}})
	b, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if strings.Contains(got, "RawJSON") || strings.Contains(got, "raw") {
		t.Fatalf("institution report leaked raw JSON: %s", got)
	}
	if !strings.Contains(got, "provider_institution_id") {
		t.Fatalf("institution report missing stable JSON field: %s", got)
	}
}

func TestSyncReportSeparatesProviderAndLocalConnectionIDs(t *testing.T) {
	report := syncReport{
		ProviderConnectionID: "req_provider",
		ConnectionID:         "connection_local",
	}
	b, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"provider_connection_id":"req_provider"`) || !strings.Contains(got, `"connection_id":"connection_local"`) {
		t.Fatalf("sync report IDs = %s", got)
	}
}

func TestInstitutionArchiveCountriesPreferMatchingConnection(t *testing.T) {
	cfg := config.Config{
		DefaultCountry: "BE",
		Connections: []config.Connection{
			{Provider: "gocardless", InstitutionID: "BUNQ_NL", Country: "NL"},
			{Provider: "gocardless", InstitutionID: "OTHER_DE", Country: "DE"},
		},
	}
	got := institutionArchiveCountries(cfg, "gocardless", "BUNQ_NL")
	want := []string{"NL", "BE", "DE"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("countries = %v, want %v", got, want)
	}
}
