package store

import (
	"context"
	"database/sql"
	"time"

	"gobankcli/internal/provider"
)

func (s *Store) UpsertInstitution(ctx context.Context, institution provider.Institution) (string, error) {
	id := institution.ID
	if id == "" {
		id = stableID("institution", institution.Provider, institution.ProviderInstitutionID)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
insert into institutions(id, provider, provider_institution_id, name, country, bic, raw_json, updated_at)
values(?,?,?,?,?,?,?,?)
on conflict(provider, provider_institution_id) do update set
	name=excluded.name,
	country=excluded.country,
	bic=excluded.bic,
	raw_json=excluded.raw_json,
	updated_at=excluded.updated_at`,
		id, institution.Provider, institution.ProviderInstitutionID, institution.Name, institution.Country, institution.BIC, institution.RawJSON, now)
	if err != nil {
		return "", err
	}
	return s.institutionID(ctx, institution.Provider, institution.ProviderInstitutionID)
}

func (s *Store) institutionID(ctx context.Context, providerName, providerInstitutionID string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `select id from institutions where provider = ? and provider_institution_id = ?`, providerName, providerInstitutionID).Scan(&id)
	return id, err
}

func (s *Store) HasInstitution(ctx context.Context, providerName, providerInstitutionID string) (bool, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `select id from institutions where provider = ? and provider_institution_id = ?`, providerName, providerInstitutionID).Scan(&id)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}
