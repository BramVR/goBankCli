package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
