package provider

import (
	"context"
	"time"
)

type Config struct {
	Credentials map[string]string
	BaseURL     string
}

type Provider interface {
	Name() string
	ListInstitutions(ctx context.Context, country string) ([]Institution, error)
	StartConnection(ctx context.Context, institutionID string, redirectURL string) (ConnectionSession, error)
	GetConnection(ctx context.Context, connectionID string) (Connection, error)
	ListAccounts(ctx context.Context, connectionID string) ([]Account, error)
	FetchTransactions(ctx context.Context, accountID string, from, to time.Time) ([]Transaction, error)
}

type Factory func(Config) (Provider, error)
