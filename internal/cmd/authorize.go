package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gobankcli/internal/provider/enablebanking"
	"gobankcli/internal/store"
)

type AuthorizeCmd struct {
	Provider    string `help:"Provider name." default:""`
	Code        string `help:"Callback authorization code."`
	State       string `help:"Callback state."`
	CallbackURL string `name:"url" help:"Full callback URL containing the authorization code."`
	Institution string `help:"Provider institution ID, needed if the provider session response omits it."`
}

type authorizeReport struct {
	ProviderConnectionID string `json:"provider_connection_id"`
	ConnectionID         string `json:"connection_id"`
	Status               string `json:"status"`
	Accounts             int    `json:"accounts"`
}

func (c AuthorizeCmd) Run(ctx context.Context, app *App) error {
	callback, err := authorizationCallback(c.Code, c.State, c.CallbackURL)
	if err != nil {
		return err
	}
	p, err := newProvider(firstString(c.Provider, app.Config.DefaultProvider))
	if err != nil {
		return err
	}
	exchanger, ok := p.(enablebanking.SessionExchanger)
	if !ok {
		return fmt.Errorf("provider %s does not support authorization code exchange", p.Name())
	}

	s, err := store.Open(ctx, app.Config.Paths.DB)
	if err != nil {
		return err
	}
	defer s.Close()
	if callback.State == "" {
		return errors.New("state is required")
	}
	found, err := s.ConnectionExists(ctx, p.Name(), callback.State)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("state does not match a pending connection")
	}

	session, accounts, err := exchanger.ExchangeSession(ctx, callback.Code)
	if err != nil {
		return err
	}
	if session.Connection.InstitutionID == "" {
		session.Connection.InstitutionID = strings.TrimSpace(c.Institution)
	}
	if session.Connection.InstitutionID == "" {
		return errors.New("institution is required when the provider session response omits it")
	}
	connectionID, err := s.UpsertConnection(ctx, session.Connection)
	if err != nil {
		return err
	}
	archivedInstitutions := map[string]bool{}
	for i := range accounts {
		if accounts[i].InstitutionID == "" {
			accounts[i].InstitutionID = session.Connection.InstitutionID
		}
		if !archivedInstitutions[accounts[i].InstitutionID] {
			countries := institutionArchiveCountries(app.Config, p.Name(), accounts[i].InstitutionID)
			if err := archiveInstitutionByID(ctx, p, s, countries, accounts[i].InstitutionID); err != nil {
				return err
			}
			archivedInstitutions[accounts[i].InstitutionID] = true
		}
		accounts[i].ConnectionID = connectionID
		if _, err := s.UpsertAccount(ctx, accounts[i]); err != nil {
			return err
		}
	}
	return app.Out.Write(authorizeReport{
		ProviderConnectionID: session.Connection.ProviderConnectionID,
		ConnectionID:         connectionID,
		Status:               session.Connection.Status,
		Accounts:             len(accounts),
	})
}

type callbackParams struct {
	Code  string
	State string
}

func authorizationCallback(code, state, callbackURL string) (callbackParams, error) {
	if strings.TrimSpace(code) != "" && strings.TrimSpace(callbackURL) != "" {
		return callbackParams{}, errors.New("use either --code or --url, not both")
	}
	if strings.TrimSpace(code) != "" {
		return callbackParams{Code: strings.TrimSpace(code), State: strings.TrimSpace(state)}, nil
	}
	if strings.TrimSpace(callbackURL) == "" {
		return callbackParams{}, errors.New("code or url is required")
	}
	u, err := url.Parse(strings.TrimSpace(callbackURL))
	if err != nil {
		return callbackParams{}, fmt.Errorf("parse callback url: %w", err)
	}
	code = strings.TrimSpace(u.Query().Get("code"))
	if code == "" {
		return callbackParams{}, errors.New("callback url missing code")
	}
	return callbackParams{Code: code, State: strings.TrimSpace(u.Query().Get("state"))}, nil
}

func authorizationCode(code, callbackURL string) (string, error) {
	callback, err := authorizationCallback(code, "", callbackURL)
	return callback.Code, err
}
