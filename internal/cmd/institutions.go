package cmd

import (
	"context"
	"strings"

	"gobankcli/internal/config"
	"gobankcli/internal/provider"
	"gobankcli/internal/store"
)

type InstitutionsCmd struct {
	Provider string `help:"Provider name." default:""`
	Country  string `help:"ISO country code."`
	Query    string `help:"Case-insensitive name or BIC filter."`
}

type institutionReport struct {
	ID                    string `json:"id"`
	Provider              string `json:"provider"`
	ProviderInstitutionID string `json:"provider_institution_id"`
	Name                  string `json:"name"`
	Country               string `json:"country"`
	BIC                   string `json:"bic"`
}

func (c InstitutionsCmd) Run(ctx context.Context, app *App) error {
	providerName := firstString(c.Provider, app.Config.DefaultProvider)
	country := firstString(c.Country, app.Config.DefaultCountry)
	p, err := newProvider(providerName)
	if err != nil {
		return err
	}
	institutions, err := p.ListInstitutions(ctx, country)
	if err != nil {
		return err
	}
	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()
	for i := range institutions {
		id, err := s.UpsertInstitution(ctx, institutions[i])
		if err != nil {
			return err
		}
		institutions[i].ID = id
	}
	if c.Query != "" {
		institutions = filterInstitutions(institutions, c.Query)
	}
	return app.Out.Write(institutionReports(institutions))
}

func archiveInstitutionByID(ctx context.Context, p provider.Provider, s *store.Store, countries []string, providerInstitutionID string) error {
	if providerInstitutionID == "" {
		return nil
	}
	exists, err := s.HasInstitution(ctx, p.Name(), providerInstitutionID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	for _, country := range countries {
		institutions, err := p.ListInstitutions(ctx, country)
		if err != nil {
			return err
		}
		for _, institution := range institutions {
			if institution.ProviderInstitutionID != providerInstitutionID {
				continue
			}
			_, err := s.UpsertInstitution(ctx, institution)
			return err
		}
	}
	return nil
}

func institutionArchiveCountries(cfg config.Config, providerName, providerInstitutionID string) []string {
	var countries []string
	addCountry := func(country string) {
		country = strings.ToUpper(strings.TrimSpace(country))
		if country == "" {
			return
		}
		for _, existing := range countries {
			if existing == country {
				return
			}
		}
		countries = append(countries, country)
	}
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	for _, connection := range cfg.Connections {
		if !sameConfigProvider(connection.Provider, providerName) || strings.TrimSpace(connection.InstitutionID) != providerInstitutionID {
			continue
		}
		addCountry(connection.Country)
	}
	addCountry(cfg.DefaultCountry)
	for _, connection := range cfg.Connections {
		if !sameConfigProvider(connection.Provider, providerName) {
			continue
		}
		addCountry(connection.Country)
	}
	return countries
}

func sameConfigProvider(configProvider, providerName string) bool {
	configProvider = strings.ToLower(strings.TrimSpace(configProvider))
	return configProvider == "" || configProvider == providerName
}

func filterInstitutions(institutions []provider.Institution, query string) []provider.Institution {
	q := strings.ToLower(strings.TrimSpace(query))
	filtered := make([]provider.Institution, 0, len(institutions))
	for _, institution := range institutions {
		if strings.Contains(strings.ToLower(institution.Name), q) || strings.Contains(strings.ToLower(institution.BIC), q) || strings.Contains(strings.ToLower(institution.ProviderInstitutionID), q) {
			filtered = append(filtered, institution)
		}
	}
	return filtered
}

func firstString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func institutionReports(institutions []provider.Institution) []institutionReport {
	reports := make([]institutionReport, 0, len(institutions))
	for _, institution := range institutions {
		reports = append(reports, institutionReport{
			ID:                    institution.ID,
			Provider:              institution.Provider,
			ProviderInstitutionID: institution.ProviderInstitutionID,
			Name:                  institution.Name,
			Country:               institution.Country,
			BIC:                   institution.BIC,
		})
	}
	return reports
}
