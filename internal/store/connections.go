package store

import (
	"context"
	"time"

	"gobankcli/internal/provider"
)

func (s *Store) UpsertConnection(ctx context.Context, connection provider.Connection) (string, error) {
	id := connection.ID
	if id == "" {
		id = LocalConnectionID(connection.Provider, connection.ProviderConnectionID)
	}
	createdAt := timeString(connection.CreatedAt)
	updatedAt := timeString(connection.UpdatedAt)
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if updatedAt == "" {
		updatedAt = createdAt
	}
	institutionID := connection.InstitutionID
	if institutionID != "" {
		institutionID = stableID("institution", connection.Provider, institutionID)
	}
	expiresAt := timeStringPtr(connection.ExpiresAt)
	_, err := s.db.ExecContext(ctx, `
insert into connections(id, provider, provider_connection_id, institution_id, status, redirect_url, created_at, updated_at, expires_at, raw_json)
values(?,?,?,?,?,?,?,?,?,?)
on conflict(provider, provider_connection_id) do update set
	institution_id=excluded.institution_id,
	status=excluded.status,
	redirect_url=excluded.redirect_url,
	updated_at=excluded.updated_at,
	expires_at=excluded.expires_at,
	raw_json=excluded.raw_json`,
		id, connection.Provider, connection.ProviderConnectionID, institutionID, connection.Status, connection.RedirectURL, createdAt, updatedAt, expiresAt, connection.RawJSON)
	if err != nil {
		return "", err
	}
	return s.connectionID(ctx, connection.Provider, connection.ProviderConnectionID)
}

func LocalConnectionID(providerName, providerConnectionID string) string {
	return stableID("connection", providerName, providerConnectionID)
}

func (s *Store) connectionID(ctx context.Context, providerName, providerConnectionID string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `select id from connections where provider = ? and provider_connection_id = ?`, providerName, providerConnectionID).Scan(&id)
	return id, err
}

func timeString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func timeStringPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return timeString(*t)
}
