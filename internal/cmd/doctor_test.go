package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gobankcli/internal/config"
	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

func TestDoctorJSONMissingCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--json", "doctor"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	for _, want := range []string{`"config_exists": false`, `"gocardless_secret_id": "missing"`, `"gocardless_secret_key": "missing"`, `"gocardless_configured": false`, `"enablebanking_application_id": "missing"`, `"enablebanking_private_key_path": "missing"`, `"enablebanking_api": "default"`, `"enablebanking_configured": false`} {
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
	if !strings.Contains(string(b), `default_provider = 'enablebanking'`) {
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

func TestInstitutionsMissingDefaultEnableBankingCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(config.EnvEnableBankingApplicationID, "")
	t.Setenv(config.EnvEnableBankingPrivateKey, "")
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"institutions", "--country", "BE"}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "enablebanking credentials missing") {
		t.Fatalf("Run error = %v, want missing credentials", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestInstitutionsMissingEnableBankingCredentials(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(config.EnvEnableBankingApplicationID, "")
	t.Setenv(config.EnvEnableBankingPrivateKey, "")
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"institutions", "--provider", "enablebanking", "--country", "BE"}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "enablebanking credentials missing") {
		t.Fatalf("Run error = %v, want missing credentials", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestAccountsEnableBankingArchivesAndKeepsOutputShape(t *testing.T) {
	var sawConfiguredCountry bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions/session-1":
			writeCmdJSON(t, w, map[string]any{
				"session_id": "session-1",
				"status":     "AUTHORIZED",
				"aspsp":      map[string]string{"name": "Belfius", "country": "NL"},
				"accounts": []map[string]any{{
					"uid":                 "session-account-uid",
					"identification_hash": "stable-hash",
					"name":                "Current",
					"currency":            "EUR",
					"account_id":          map[string]string{"iban": "NL00BANK0000000000"},
				}},
			})
		case "/aspsps":
			if got := r.URL.Query().Get("country"); got != "NL" {
				http.Error(w, "unexpected country "+got, http.StatusInternalServerError)
				return
			}
			sawConfiguredCountry = true
			writeCmdJSON(t, w, map[string]any{"aspsps": []map[string]any{{
				"name":    "Belfius",
				"country": "NL",
			}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	dbPath := filepath.Join(home, "archive.db")
	configPath := filepath.Join(home, "config.toml")
	if err := os.WriteFile(configPath, []byte(`
default_provider = 'enablebanking'
default_country = 'BE'

[paths]
db = '`+dbPath+`'
exports = '`+filepath.Join(home, "exports")+`'

[[connections]]
provider = 'enablebanking'
institution_id = 'NL:Belfius'
country = 'NL'
`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv(config.EnvEnableBankingApplicationID, "app-123")
	t.Setenv(config.EnvEnableBankingPrivateKey, writeTestRSAKey(t))
	t.Setenv(config.EnvEnableBankingAPI, server.URL)

	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--config", configPath, "--json", "accounts", "--provider", "enablebanking", "--connection", "session-1"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	gotJSON := stdout.String()
	for _, want := range []string{`"provider_account_id": "stable-hash"`, `"provider_resource_id": "session-account-uid"`, `"institution_id": "NL:Belfius"`, `"count": 1`} {
		if !strings.Contains(gotJSON, want) {
			t.Fatalf("accounts JSON missing %s:\n%s", want, gotJSON)
		}
	}
	if strings.Contains(gotJSON, "raw_json") || strings.Contains(gotJSON, "RawJSON") {
		t.Fatalf("accounts JSON leaked raw JSON:\n%s", gotJSON)
	}
	if !sawConfiguredCountry {
		t.Fatal("institution archive did not use configured country")
	}

	stdout.Reset()
	err = Run(context.Background(), []string{"--config", configPath, "--plain", "accounts", "--provider", "enablebanking", "--connection", "session-1"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	gotPlain := stdout.String()
	for _, want := range []string{"accounts[0].provider_account_id: stable-hash", "accounts[0].provider_resource_id: session-account-uid", "accounts[0].institution_id: NL:Belfius", "count: 1"} {
		if !strings.Contains(gotPlain, want) {
			t.Fatalf("accounts plain missing %s:\n%s", want, gotPlain)
		}
	}
}

func TestAuthorizeEnableBankingExchangesCodeAndArchivesAccounts(t *testing.T) {
	var sawSessionCode bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sessions":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["code"] != "callback-code" {
				t.Fatalf("session body = %#v", body)
			}
			sawSessionCode = true
			writeCmdJSON(t, w, map[string]any{
				"session_id": "session-1",
				"status":     "AUTHORIZED",
				"access":     map[string]string{"valid_until": "2026-08-15T12:00:00Z"},
				"accounts": []map[string]any{{
					"uid":                 "session-account-uid",
					"identification_hash": "stable-hash",
					"name":                "Current",
					"currency":            "EUR",
					"account_id":          map[string]string{"iban": "BE00000000000000"},
				}},
			})
		case "/sessions/session-1":
			writeCmdJSON(t, w, map[string]any{
				"session_id": "session-1",
				"status":     "AUTHORIZED",
				"aspsp":      map[string]string{"name": "Belfius", "country": "BE"},
				"accounts":   []string{"session-account-uid"},
			})
		case "/accounts/session-account-uid/transactions":
			writeCmdJSON(t, w, map[string]any{"transactions": []map[string]any{{
				"transaction_id":         "tx-1",
				"status":                 "BOOK",
				"credit_debit_indicator": "DBIT",
				"booking_date":           "2026-05-01",
				"transaction_amount":     map[string]string{"amount": "12.34", "currency": "EUR"},
				"creditor":               map[string]string{"name": "Shop"},
			}}})
		case "/aspsps":
			writeCmdJSON(t, w, map[string]any{"aspsps": []map[string]any{{
				"name":    "Belfius",
				"country": "BE",
			}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	dbPath := filepath.Join(home, "archive.db")
	ctx := context.Background()
	s, err := store.Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpsertConnection(ctx, provider.Connection{
		Provider:             "enablebanking",
		ProviderConnectionID: "pending-state",
		InstitutionID:        "BE:Belfius",
		Status:               "PENDING",
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv(config.EnvEnableBankingApplicationID, "app-123")
	t.Setenv(config.EnvEnableBankingPrivateKey, writeTestRSAKey(t))
	t.Setenv(config.EnvEnableBankingAPI, server.URL)
	var stdout, stderr bytes.Buffer
	err = Run(ctx, []string{"--db", dbPath, "--json", "authorize", "--provider", "enablebanking", "--url", "https://app.example/callback?code=callback-code&state=pending-state", "--institution", "BE:Belfius"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !sawSessionCode {
		t.Fatal("session exchange was not called")
	}
	got := stdout.String()
	for _, want := range []string{`"provider_connection_id": "session-1"`, `"status": "AUTHORIZED"`, `"accounts": 1`} {
		if !strings.Contains(got, want) {
			t.Fatalf("authorize output missing %s:\n%s", want, got)
		}
	}

	stdout.Reset()
	err = Run(context.Background(), []string{"--db", dbPath, "--plain", "query", "select provider, provider_account_id, provider_resource_id from accounts"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); !strings.Contains(got, "enablebanking\tstable-hash\tsession-account-uid") {
		t.Fatalf("archived account row = %q", got)
	}

	stdout.Reset()
	err = Run(ctx, []string{"--db", dbPath, "--json", "sync", "--provider", "enablebanking", "--connection", "session-1", "--from", "2026-05-01", "--to", "2026-05-31"}, "test", &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); !strings.Contains(got, `"transactions_seen": 1`) {
		t.Fatalf("sync output = %q", got)
	}
}

func TestConnectEnableBankingListenExchangesCallback(t *testing.T) {
	type authRequest struct {
		redirectURL string
		state       string
	}
	authRequests := make(chan authRequest, 1)
	var sawSessionCode bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aspsps":
			writeCmdJSON(t, w, map[string]any{"aspsps": []map[string]any{{
				"name":                     "Belfius",
				"country":                  "BE",
				"maximum_consent_validity": 86400,
			}}})
		case "/auth":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			authRequests <- authRequest{
				redirectURL: body["redirect_url"].(string),
				state:       body["state"].(string),
			}
			writeCmdJSON(t, w, map[string]string{"url": "https://bank.example/auth"})
		case "/sessions":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["code"] != "callback-code" {
				t.Fatalf("session body = %#v", body)
			}
			sawSessionCode = true
			writeCmdJSON(t, w, map[string]any{
				"session_id": "session-1",
				"status":     "AUTHORIZED",
				"access":     map[string]string{"valid_until": "2026-08-15T12:00:00Z"},
				"accounts": []map[string]any{{
					"uid":                 "session-account-uid",
					"identification_hash": "stable-hash",
					"name":                "Current",
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	dbPath := filepath.Join(home, "archive.db")
	t.Setenv("HOME", home)
	t.Setenv(config.EnvEnableBankingApplicationID, "app-123")
	t.Setenv(config.EnvEnableBankingPrivateKey, writeTestRSAKey(t))
	t.Setenv(config.EnvEnableBankingAPI, server.URL)
	var stdout, stderr bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- Run(context.Background(), []string{"--db", dbPath, "--json", "connect", "--provider", "enablebanking", "--institution", "BE:Belfius", "--listen", "127.0.0.1:0", "--callback-timeout", "3s"}, "test", &stdout, &stderr)
	}()

	var auth authRequest
	select {
	case auth = <-authRequests:
	case <-time.After(2 * time.Second):
		t.Fatal("auth request was not sent")
	}
	callbackURL := auth.redirectURL + "?code=callback-code&state=" + auth.state
	deadline := time.After(2 * time.Second)
	for {
		resp, err := http.Get(callbackURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		select {
		case <-time.After(25 * time.Millisecond):
		case <-deadline:
			t.Fatal("callback server did not accept request")
		}
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("connect did not finish")
	}
	if !sawSessionCode {
		t.Fatal("session exchange was not called")
	}
	if got := stdout.String(); !strings.Contains(got, `"provider_connection_id": "session-1"`) || !strings.Contains(got, `"accounts": 1`) {
		t.Fatalf("connect output = %s", got)
	}
	if got := stderr.String(); !strings.Contains(got, "Open this URL: https://bank.example/auth") || !strings.Contains(got, "Waiting for callback") {
		t.Fatalf("stderr = %q", got)
	}
}

func TestConnectListenRejectsNoInput(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(config.EnvEnableBankingApplicationID, "app-123")
	t.Setenv(config.EnvEnableBankingPrivateKey, writeTestRSAKey(t))
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"--no-input", "connect", "--provider", "enablebanking", "--institution", "BE:Belfius", "--listen", "127.0.0.1:0"}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "listen is not supported with --no-input") {
		t.Fatalf("Run error = %v, want no-input listen rejection", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestConnectListenHTTPSRequiresListen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv(config.EnvEnableBankingApplicationID, "app-123")
	t.Setenv(config.EnvEnableBankingPrivateKey, writeTestRSAKey(t))
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"connect", "--provider", "enablebanking", "--institution", "BE:Belfius", "--redirect", "https://example.test/callback", "--listen-https"}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "listen-https requires listen") {
		t.Fatalf("Run error = %v, want listen-https rejection", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestConnectListenHTTPSValidatesCertBeforeAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected provider request before TLS validation: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	t.Setenv("HOME", t.TempDir())
	t.Setenv(config.EnvEnableBankingApplicationID, "app-123")
	t.Setenv(config.EnvEnableBankingPrivateKey, writeTestRSAKey(t))
	t.Setenv(config.EnvEnableBankingAPI, server.URL)
	var stdout, stderr bytes.Buffer
	err := Run(context.Background(), []string{"connect", "--provider", "enablebanking", "--institution", "BE:Belfius", "--listen", "127.0.0.1:0", "--listen-https", "--listen-cert", filepath.Join(t.TempDir(), "missing.crt"), "--listen-key", filepath.Join(t.TempDir(), "missing.key")}, "test", &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("Run error = %v, want missing cert rejection", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestAuthorizationCodeParsing(t *testing.T) {
	got, err := authorizationCode("", "https://app.example/callback?state=s&code=abc")
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc" {
		t.Fatalf("code = %q", got)
	}
	callback, err := authorizationCallback("", "", "https://app.example/callback?state=s&code=abc")
	if err != nil {
		t.Fatal(err)
	}
	if callback.Code != "abc" || callback.State != "s" {
		t.Fatalf("callback = %+v", callback)
	}
	if _, err := authorizationCode("abc", "https://app.example/callback?code=def"); err == nil || !strings.Contains(err.Error(), "either --code or --url") {
		t.Fatalf("err = %v, want exclusivity error", err)
	}
	if _, err := authorizationCode("", "https://app.example/callback?state=s"); err == nil || !strings.Contains(err.Error(), "missing code") {
		t.Fatalf("err = %v, want missing code", err)
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

func writeTestRSAKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "enablebanking.pem")
	b := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeCmdJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatal(err)
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
