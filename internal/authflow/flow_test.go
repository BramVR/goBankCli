package authflow

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"gobankcli/internal/config"
	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

func TestFlowRunExchangesCallbackAndArchivesAccounts(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(ctx, t.TempDir()+"/archive.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p := &fakeProvider{
		started: make(chan startedConnection, 1),
	}
	flow := Flow{
		Config:        config.Default(),
		Provider:      p,
		Exchanger:     p,
		Store:         s,
		InstitutionID: "BE:Belfius",
		ListenAddress: "127.0.0.1:0",
		Timeout:       2 * time.Second,
		Stderr:        io.Discard,
	}
	done := make(chan flowResult, 1)
	go func() {
		report, err := flow.Run(ctx)
		done <- flowResult{report: report, err: err}
	}()

	var started startedConnection
	select {
	case started = <-p.started:
	case <-time.After(time.Second):
		t.Fatal("start connection was not called")
	}
	if !strings.HasPrefix(started.redirectURL, "http://127.0.0.1:") {
		t.Fatalf("redirect url = %q, want loopback URL", started.redirectURL)
	}
	resp, err := http.Get(started.redirectURL + "?code=callback-code&state=" + started.state)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("callback status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	select {
	case result := <-done:
		if result.err != nil {
			t.Fatal(result.err)
		}
		if result.report.ProviderConnectionID != "session-1" || result.report.Accounts != 1 {
			t.Fatalf("report = %+v", result.report)
		}
	case <-time.After(time.Second):
		t.Fatal("flow did not finish")
	}
	if p.exchangedCode != "callback-code" {
		t.Fatalf("exchanged code = %q", p.exchangedCode)
	}
	accounts, err := s.AccountsByConnection(ctx, store.LocalConnectionID("testbank", "session-1"))
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].ConnectionID == "" {
		t.Fatalf("archived accounts = %+v", accounts)
	}
}

type flowResult struct {
	report Report
	err    error
}

type startedConnection struct {
	redirectURL string
	state       string
}

type fakeProvider struct {
	started       chan startedConnection
	exchangedCode string
}

func (p *fakeProvider) Name() string { return "testbank" }

func (p *fakeProvider) ListInstitutions(ctx context.Context, country string) ([]provider.Institution, error) {
	return []provider.Institution{{
		Provider:              p.Name(),
		ProviderInstitutionID: "BE:Belfius",
		Name:                  "Belfius",
		Country:               "BE",
	}}, nil
}

func (p *fakeProvider) StartConnection(ctx context.Context, institutionID string, redirectURL string) (provider.ConnectionSession, error) {
	state := "pending-state"
	p.started <- startedConnection{redirectURL: redirectURL, state: state}
	return provider.ConnectionSession{
		Connection: provider.Connection{
			Provider:             p.Name(),
			ProviderConnectionID: state,
			InstitutionID:        institutionID,
			Status:               "PENDING",
			RedirectURL:          "https://bank.example/auth",
		},
		RedirectURL: "https://bank.example/auth",
	}, nil
}

func (p *fakeProvider) ExchangeSession(ctx context.Context, code string) (provider.ConnectionSession, []provider.Account, error) {
	p.exchangedCode = code
	return provider.ConnectionSession{
			Connection: provider.Connection{
				Provider:             p.Name(),
				ProviderConnectionID: "session-1",
				InstitutionID:        "BE:Belfius",
				Status:               "AUTHORIZED",
			},
		}, []provider.Account{{
			Provider:           p.Name(),
			ProviderAccountID:  "account-1",
			ProviderResourceID: "resource-1",
			InstitutionID:      "BE:Belfius",
			Name:               "Current",
		}}, nil
}

func (p *fakeProvider) GetConnection(ctx context.Context, connectionID string) (provider.Connection, error) {
	return provider.Connection{}, nil
}

func (p *fakeProvider) ListAccounts(ctx context.Context, connectionID string) ([]provider.Account, error) {
	return nil, nil
}

func (p *fakeProvider) FetchTransactions(ctx context.Context, accountID string, from, to time.Time) ([]provider.Transaction, error) {
	return nil, nil
}
