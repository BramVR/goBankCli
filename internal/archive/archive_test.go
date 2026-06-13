package archive

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gobankcli/internal/config"
	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

func TestManagerArchivesFreshAccountsWithInstitutions(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	p := &fakeProvider{
		name: "example",
		institutions: map[string][]provider.Institution{
			"NL": {{Provider: "example", ProviderInstitutionID: "BANK_NL", Name: "Bank NL", Country: "NL"}},
		},
	}
	manager := NewManager(config.Config{
		DefaultCountry: "BE",
		Connections: []config.Connection{
			{Provider: "example", InstitutionID: "BANK_NL", Country: "NL"},
		},
	}, p, s)

	archived, err := manager.ArchiveAccounts(ctx, store.LocalConnectionID("example", "conn-1"), []provider.Account{{
		Provider:          "example",
		ProviderAccountID: "acct-1",
		InstitutionID:     "BANK_NL",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(archived) != 1 || archived[0].ID == "" {
		t.Fatalf("archived accounts = %+v", archived)
	}
	rows, err := s.Query(ctx, "select i.country, a.connection_id from accounts a join institutions i on i.id = a.institution_id")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows.Rows) != 1 || rows.Rows[0][0] != "NL" || rows.Rows[0][1] != store.LocalConnectionID("example", "conn-1") {
		t.Fatalf("archive rows = %+v", rows.Rows)
	}
	if strings.Join(p.countriesSeen, ",") != "NL" {
		t.Fatalf("institution countries = %v, want matching connection country first only", p.countriesSeen)
	}
}

func TestManagerReusesStoredAccountsWhenStableAccountIDMissing(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(ctx, filepath.Join(t.TempDir(), "gobankcli.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	localConnectionID := store.LocalConnectionID("example", "conn-1")
	storedID, err := s.UpsertAccount(ctx, provider.Account{
		Provider:          "example",
		ProviderAccountID: "stable-account",
		ConnectionID:      localConnectionID,
	})
	if err != nil {
		t.Fatal(err)
	}

	manager := NewManager(config.Config{}, &fakeProvider{
		name:            "example",
		listAccountsErr: provider.ErrMissingStableAccountID,
	}, s)
	result, err := manager.AccountsForConnection(ctx, "conn-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Fresh {
		t.Fatal("stored accounts should not be marked fresh")
	}
	if len(result.Accounts) != 1 || result.Accounts[0].ID != storedID {
		t.Fatalf("accounts = %+v, want stored id %q", result.Accounts, storedID)
	}
}

func TestArchiveCountriesPreferMatchingConnectionBeforeDefault(t *testing.T) {
	cfg := config.Config{
		DefaultCountry: "BE",
		Connections: []config.Connection{
			{Provider: "example", InstitutionID: "BANK_NL", Country: "NL"},
			{Provider: "example", InstitutionID: "OTHER_DE", Country: "DE"},
		},
	}
	got := archiveCountries(cfg, "example", "BANK_NL")
	want := []string{"NL", "BE", "DE"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("countries = %v, want %v", got, want)
	}
}

type fakeProvider struct {
	name            string
	institutions    map[string][]provider.Institution
	accounts        []provider.Account
	listAccountsErr error
	countriesSeen   []string
}

func (p *fakeProvider) Name() string { return p.name }

func (p *fakeProvider) ListInstitutions(_ context.Context, country string) ([]provider.Institution, error) {
	p.countriesSeen = append(p.countriesSeen, country)
	if p.institutions == nil {
		return nil, nil
	}
	return p.institutions[country], nil
}

func (p *fakeProvider) StartConnection(context.Context, string, string) (provider.ConnectionSession, error) {
	return provider.ConnectionSession{}, errors.New("not implemented")
}

func (p *fakeProvider) GetConnection(context.Context, string) (provider.Connection, error) {
	return provider.Connection{}, errors.New("not implemented")
}

func (p *fakeProvider) ListAccounts(context.Context, string) ([]provider.Account, error) {
	return p.accounts, p.listAccountsErr
}

func (p *fakeProvider) FetchTransactions(_ context.Context, accountID string, _, _ time.Time) ([]provider.Transaction, error) {
	return nil, errors.New("not implemented")
}
