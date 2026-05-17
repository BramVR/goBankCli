package enablebanking

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gobankcli/internal/provider"
)

func TestCreateJWTUsesEnableBankingClaimsAndKeyID(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	token, err := createJWT("app-123", key, now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT parts = %d, want 3", len(parts))
	}
	header := decodeJWTPart(t, parts[0])
	if header["alg"] != "RS256" || header["kid"] != "app-123" {
		t.Fatalf("JWT header = %+v", header)
	}
	payload := decodeJWTPart(t, parts[1])
	if payload["iss"] != "enablebanking.com" || payload["aud"] != "api.enablebanking.com" || payload["iat"].(float64) != float64(now.Unix()) || payload["exp"].(float64) != float64(now.Add(time.Hour).Unix()) {
		t.Fatalf("JWT payload = %+v", payload)
	}
}

func TestClientStartConnectionPostsAuthRequest(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aspsps":
			writeJSON(t, w, map[string]any{"aspsps": []map[string]any{{
				"name":                     "Belfius",
				"country":                  "BE",
				"maximum_consent_validity": 86400,
			}}})
		case "/auth":
			gotAuth = r.Header.Get("Authorization")
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
			writeJSON(t, w, map[string]string{"url": "https://bank.example/auth"})
		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer server.Close()

	p := newTestProvider(t, server.URL)
	c := p.(*Client)
	c.now = func() time.Time { return time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC) }
	session, err := p.StartConnection(context.Background(), "BE:Belfius", "https://app.example/callback")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(gotAuth, "Bearer ") {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	aspsp := gotBody["aspsp"].(map[string]any)
	access := gotBody["access"].(map[string]any)
	if aspsp["name"] != "Belfius" || aspsp["country"] != "BE" || gotBody["redirect_url"] != "https://app.example/callback" || gotBody["psu_type"] != "personal" || access["valid_until"] != "2026-05-18T12:00:00Z" {
		t.Fatalf("auth body = %#v", gotBody)
	}
	if session.Connection.Provider != Name || session.Connection.ProviderConnectionID == "" || session.Connection.InstitutionID != "BE:Belfius" || session.Connection.Status != "PENDING" || session.RedirectURL != "https://bank.example/auth" {
		t.Fatalf("session = %+v", session)
	}
}

func TestClientListsInstitutionsAndExchangesSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/aspsps":
			if r.URL.Query().Get("country") != "BE" {
				t.Fatalf("country query = %q", r.URL.RawQuery)
			}
			writeJSON(t, w, map[string]any{"aspsps": []map[string]any{{
				"name":                     "Belfius",
				"country":                  "BE",
				"psu_types":                []string{"personal"},
				"maximum_consent_validity": 7776000,
			}}})
		case "/sessions":
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["code"] != "callback-code" {
				t.Fatalf("session code body = %#v", body)
			}
			writeJSON(t, w, sessionFixture())
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	p := newTestProvider(t, server.URL)
	institutions, err := p.ListInstitutions(context.Background(), "be")
	if err != nil {
		t.Fatal(err)
	}
	if len(institutions) != 1 || institutions[0].ProviderInstitutionID != "BE:Belfius" || institutions[0].Name != "Belfius" || institutions[0].Country != "BE" {
		t.Fatalf("institutions = %+v", institutions)
	}

	exchanger := p.(SessionExchanger)
	session, accounts, err := exchanger.ExchangeSession(context.Background(), "callback-code")
	if err != nil {
		t.Fatal(err)
	}
	if session.Connection.ProviderConnectionID != "session-1" || session.Connection.InstitutionID != "BE:Belfius" || session.Connection.ExpiresAt == nil {
		t.Fatalf("connection = %+v", session.Connection)
	}
	if len(accounts) != 1 || accounts[0].ProviderAccountID != "stable-identification-hash" || accounts[0].ProviderResourceID != "session-account-uid" || accounts[0].IBAN != "BE00000000000000" || accounts[0].Name != "Current | Main account" {
		t.Fatalf("accounts = %+v", accounts)
	}
}

func TestClientListAccountsRejectsUIDOnlySessionAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/session-1" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, map[string]any{
			"session_id": "session-1",
			"status":     "AUTHORIZED",
			"aspsp":      map[string]string{"name": "Belfius", "country": "BE"},
			"accounts":   []string{"session-account-uid"},
		})
	}))
	defer server.Close()

	p := newTestProvider(t, server.URL)
	if _, err := p.ListAccounts(context.Background(), "session-1"); err != ErrMissingStableAccountID {
		t.Fatalf("err = %v, want missing stable account id", err)
	}
}

func TestClientFetchTransactionsPaginatesAndNormalizesBookedOnly(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accounts/session-account-uid/transactions" {
			http.NotFound(w, r)
			return
		}
		calls++
		if r.URL.Query().Get("transaction_status") != "BOOK" || r.URL.Query().Get("date_from") != "2026-01-01" || r.URL.Query().Get("date_to") != "2026-01-31" {
			t.Fatalf("transaction query = %q", r.URL.RawQuery)
		}
		if calls == 1 {
			writeJSON(t, w, map[string]any{
				"continuation_key": "next",
				"transactions": []map[string]any{
					debitTransaction(),
					{"status": "PDNG", "transaction_amount": map[string]string{"amount": "1.00", "currency": "EUR"}, "booking_date": "2026-01-01"},
				},
			})
			return
		}
		if r.URL.Query().Get("continuation_key") != "next" {
			t.Fatalf("continuation query = %q", r.URL.RawQuery)
		}
		writeJSON(t, w, map[string]any{"transactions": []map[string]any{creditTransaction()}})
	}))
	defer server.Close()

	p := newTestProvider(t, server.URL)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	transactions, err := p.FetchTransactions(context.Background(), "session-account-uid", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(transactions) != 2 {
		t.Fatalf("calls=%d transactions=%+v", calls, transactions)
	}
	if transactions[0].Amount != "-12.34" || transactions[0].CounterpartyName != "Shop" || transactions[0].CounterpartyAccount != "BE11111111111111" || transactions[0].RemittanceInfo != "Invoice 1 | Card payment" || transactions[0].Reference != "ref-1" {
		t.Fatalf("debit transaction = %+v", transactions[0])
	}
	if transactions[1].Amount != "50.00" || transactions[1].CounterpartyName != "Employer" || transactions[1].CounterpartyAccount != "BE22222222222222" {
		t.Fatalf("credit transaction = %+v", transactions[1])
	}
}

func TestClientMissingCredentialsFailClearly(t *testing.T) {
	p, err := New(provider.Config{BaseURL: "https://example.test"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.ListInstitutions(context.Background(), "BE")
	if err == nil || !strings.Contains(err.Error(), "enablebanking credentials missing") {
		t.Fatalf("error = %v, want missing credentials", err)
	}
}

func newTestProvider(t *testing.T, baseURL string) provider.Provider {
	t.Helper()
	keyPath := filepath.Join(t.TempDir(), "key.pem")
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	b := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(keyPath, b, 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := New(provider.Config{
		BaseURL: baseURL,
		Credentials: map[string]string{
			CredentialApplicationID: "app-123",
			CredentialPrivateKey:    keyPath,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func sessionFixture() map[string]any {
	return map[string]any{
		"session_id": "session-1",
		"status":     "AUTHORIZED",
		"access": map[string]string{
			"valid_until": "2026-08-15T12:00:00Z",
		},
		"aspsp": map[string]string{"name": "Belfius", "country": "BE"},
		"accounts": []map[string]any{{
			"uid":                 "session-account-uid",
			"identification_hash": "stable-identification-hash",
			"name":                "Current",
			"details":             "Main account",
			"currency":            "EUR",
			"account_id": map[string]string{
				"iban": "BE00000000000000",
			},
		}},
	}
}

func debitTransaction() map[string]any {
	return map[string]any{
		"transaction_id":         "tx-1",
		"entry_reference":        "entry-1",
		"reference_number":       "ref-1",
		"status":                 "BOOK",
		"credit_debit_indicator": "DBIT",
		"booking_date":           "2026-01-05",
		"value_date":             "2026-01-06",
		"transaction_amount":     map[string]string{"amount": "12.34", "currency": "EUR"},
		"creditor":               map[string]string{"name": "Shop"},
		"creditor_account":       map[string]string{"iban": "BE11111111111111"},
		"remittance_information": []string{"Invoice 1"},
		"note":                   "Card payment",
	}
}

func creditTransaction() map[string]any {
	return map[string]any{
		"transaction_id":         "tx-2",
		"status":                 "BOOK",
		"credit_debit_indicator": "CRDT",
		"booking_date":           "2026-01-10",
		"transaction_amount":     map[string]string{"amount": "50.00", "currency": "EUR"},
		"debtor":                 map[string]string{"name": "Employer"},
		"debtor_account":         map[string]string{"iban": "BE22222222222222"},
	}
}

func decodeJWTPart(t *testing.T, part string) map[string]any {
	t.Helper()
	b, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatal(err)
	}
}
