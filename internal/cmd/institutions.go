package cmd

import (
	"context"
	"strings"

	"gobankcli/internal/provider"
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
	if c.Query != "" {
		institutions = filterInstitutions(institutions, c.Query)
	}
	return app.Out.Write(institutionReports(institutions))
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
