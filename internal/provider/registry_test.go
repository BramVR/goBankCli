package provider

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeProvider struct {
	name string
}

func (p fakeProvider) Name() string { return p.name }

func (p fakeProvider) ListInstitutions(context.Context, string) ([]Institution, error) {
	return nil, nil
}

func (p fakeProvider) StartConnection(context.Context, string, string) (ConnectionSession, error) {
	return ConnectionSession{}, nil
}

func (p fakeProvider) GetConnection(context.Context, string) (Connection, error) {
	return Connection{}, nil
}

func (p fakeProvider) ListAccounts(context.Context, string) ([]Account, error) {
	return nil, nil
}

func (p fakeProvider) FetchTransactions(context.Context, string, time.Time, time.Time) ([]Transaction, error) {
	return nil, nil
}

func TestRegistryCreateNormalizesName(t *testing.T) {
	reg := NewRegistry()
	err := reg.Register("GoCardless", func(Config) (Provider, error) {
		return fakeProvider{name: "gocardless"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	provider, err := reg.Create(" gocardless ", Config{})
	if err != nil {
		t.Fatal(err)
	}
	if provider.Name() != "gocardless" {
		t.Fatalf("provider name = %q", provider.Name())
	}
}

func TestRegistryRejectsDuplicate(t *testing.T) {
	reg := NewRegistry()
	factory := func(Config) (Provider, error) { return fakeProvider{name: "x"}, nil }
	if err := reg.Register("x", factory); err != nil {
		t.Fatal(err)
	}
	err := reg.Register(" X ", factory)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestRegistryNamesSorted(t *testing.T) {
	reg := NewRegistry()
	factory := func(Config) (Provider, error) { return fakeProvider{name: "x"}, nil }
	for _, name := range []string{"z", "a", "m"} {
		if err := reg.Register(name, factory); err != nil {
			t.Fatal(err)
		}
	}
	got := reg.Names()
	want := []string{"a", "m", "z"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("names = %v, want %v", got, want)
		}
	}
}
