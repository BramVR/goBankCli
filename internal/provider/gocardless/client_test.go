package gocardless

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"gobankcli/internal/provider"
)

func TestClientRefreshesExpiredAccessToken(t *testing.T) {
	var institutionsCalls int
	var sawFreshToken bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/token/new/":
			writeJSON(t, w, tokenPayload{Access: "stale", AccessExpires: 60, Refresh: "refresh"})
		case "/api/v2/token/refresh/":
			writeJSON(t, w, tokenPayload{Access: "fresh", AccessExpires: 60, Refresh: "refresh"})
		case "/api/v2/institutions/":
			institutionsCalls++
			if r.Header.Get("Authorization") == "Bearer stale" {
				http.Error(w, "expired", http.StatusUnauthorized)
				return
			}
			if r.Header.Get("Authorization") == "Bearer fresh" {
				sawFreshToken = true
			}
			writeJSON(t, w, []institutionPayload{{ID: "SANDBOX", Name: "Sandbox", Countries: []string{"BE"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	p, err := New(provider.Config{
		BaseURL: server.URL + "/api/v2",
		Credentials: map[string]string{
			CredentialSecretID:  "id",
			CredentialSecretKey: "key",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	institutions, err := p.ListInstitutions(context.Background(), "BE")
	if err != nil {
		t.Fatal(err)
	}
	if len(institutions) != 1 || institutionsCalls != 2 || !sawFreshToken {
		t.Fatalf("institutions=%v calls=%d sawFresh=%v", institutions, institutionsCalls, sawFreshToken)
	}
}

func TestClientFallsBackWhenRefreshTokenIsRejected(t *testing.T) {
	var tokenNewCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/token/new/":
			tokenNewCalls++
			writeJSON(t, w, tokenPayload{Access: "access", AccessExpires: 60, Refresh: "refresh", RefreshExpires: 60})
		case "/api/v2/token/refresh/":
			http.Error(w, "refresh expired", http.StatusUnauthorized)
		case "/api/v2/institutions/":
			writeJSON(t, w, []institutionPayload{{ID: "SANDBOX", Name: "Sandbox", Countries: []string{"BE"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	p, err := New(provider.Config{
		BaseURL: server.URL + "/api/v2",
		Credentials: map[string]string{
			CredentialSecretID:  "id",
			CredentialSecretKey: "key",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	c := p.(*Client)
	c.refresh = "refresh"
	c.refreshExp = time.Now().Add(time.Hour)

	if _, err := p.ListInstitutions(context.Background(), "BE"); err != nil {
		t.Fatal(err)
	}
	if tokenNewCalls != 1 {
		t.Fatalf("token/new calls = %d, want 1", tokenNewCalls)
	}
}

func TestResolvePreservesEscapedPathSegments(t *testing.T) {
	p, err := New(provider.Config{BaseURL: "https://example.test/api/v2"})
	if err != nil {
		t.Fatal(err)
	}
	c := p.(*Client)
	got := c.resolve("/accounts/" + url.PathEscape("account/with/slash") + "/details/?x=1")
	want := "https://example.test/api/v2/accounts/account%2Fwith%2Fslash/details/?x=1"
	if got != want {
		t.Fatalf("resolve() = %q, want %q", got, want)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatal(err)
	}
}
