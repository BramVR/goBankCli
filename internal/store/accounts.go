package store

import (
	"context"
	"time"

	"gobankcli/internal/provider"
)

func (s *Store) UpsertAccount(ctx context.Context, account provider.Account) (string, error) {
	id := account.ID
	if id == "" {
		id = stableID("account", account.Provider, account.ProviderAccountID)
	}
	resourceID := account.ProviderResourceID
	if resourceID == "" {
		resourceID = account.ProviderAccountID
	}
	institutionID := account.InstitutionID
	if institutionID != "" {
		institutionID = stableID("institution", account.Provider, institutionID)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
insert into accounts(id, provider, provider_account_id, provider_resource_id, institution_id, connection_id, iban, name, currency, owner_name, raw_json, updated_at)
values(?,?,?,?,?,?,?,?,?,?,?,?)
on conflict(provider, provider_account_id) do update set
	provider_resource_id=excluded.provider_resource_id,
	institution_id=excluded.institution_id,
	connection_id=excluded.connection_id,
	iban=excluded.iban,
	name=excluded.name,
	currency=excluded.currency,
	owner_name=excluded.owner_name,
	raw_json=excluded.raw_json,
	updated_at=excluded.updated_at`,
		id, account.Provider, account.ProviderAccountID, resourceID, institutionID, account.ConnectionID, account.IBAN, account.Name, account.Currency, account.OwnerName, account.RawJSON, now)
	if err != nil {
		return "", err
	}
	return s.accountID(ctx, account.Provider, account.ProviderAccountID)
}

func (s *Store) accountID(ctx context.Context, providerName, providerAccountID string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `select id from accounts where provider = ? and provider_account_id = ?`, providerName, providerAccountID).Scan(&id)
	return id, err
}
