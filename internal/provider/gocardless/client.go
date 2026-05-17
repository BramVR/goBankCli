package gocardless

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"gobankcli/internal/provider"
)

const (
	Name        = "gocardless"
	DefaultBase = "https://bankaccountdata.gocardless.com/api/v2"

	CredentialSecretID  = "secret_id"
	CredentialSecretKey = "secret_key"
)

var ErrMissingCredentials = errors.New("gocardless credentials missing")

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	secretID   string
	secretKey  string
	access     string
	accessExp  time.Time
	refresh    string
	refreshExp time.Time
}

func New(cfg provider.Config) (provider.Provider, error) {
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBase
	}
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	creds := cfg.Credentials
	return &Client{
		baseURL:    u,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		secretID:   creds[CredentialSecretID],
		secretKey:  creds[CredentialSecretKey],
	}, nil
}

func (c *Client) Name() string { return Name }

func (c *Client) ListInstitutions(ctx context.Context, country string) ([]provider.Institution, error) {
	var raw []institutionPayload
	q := url.Values{"country": []string{strings.ToUpper(strings.TrimSpace(country))}}
	if err := c.get(ctx, "/institutions/?"+q.Encode(), &raw); err != nil {
		return nil, err
	}
	return NormalizeInstitutions(raw)
}

func (c *Client) StartConnection(ctx context.Context, institutionID string, redirectURL string) (provider.ConnectionSession, error) {
	body := map[string]string{
		"institution_id": institutionID,
		"redirect":       redirectURL,
		"reference":      fmt.Sprintf("gobankcli-%d", time.Now().UnixNano()),
	}
	var raw requisitionPayload
	if err := c.post(ctx, "/requisitions/", body, &raw); err != nil {
		return provider.ConnectionSession{}, err
	}
	return provider.ConnectionSession{
		Connection:  NormalizeConnection(raw),
		RedirectURL: raw.Link,
	}, nil
}

func (c *Client) GetConnection(ctx context.Context, connectionID string) (provider.Connection, error) {
	var raw requisitionPayload
	if err := c.get(ctx, "/requisitions/"+url.PathEscape(connectionID)+"/", &raw); err != nil {
		return provider.Connection{}, err
	}
	return NormalizeConnection(raw), nil
}

func (c *Client) ListAccounts(ctx context.Context, connectionID string) ([]provider.Account, error) {
	var requisition requisitionPayload
	if err := c.get(ctx, "/requisitions/"+url.PathEscape(connectionID)+"/", &requisition); err != nil {
		return nil, err
	}
	accounts := make([]provider.Account, 0, len(requisition.Accounts))
	for _, accountID := range requisition.Accounts {
		var details accountDetailsPayload
		if err := c.get(ctx, "/accounts/"+url.PathEscape(accountID)+"/details/", &details); err != nil {
			return nil, err
		}
		accounts = append(accounts, NormalizeAccountDetails(accountID, requisition.InstitutionID, connectionID, details))
	}
	return accounts, nil
}

func (c *Client) FetchTransactions(ctx context.Context, accountID string, from, to time.Time) ([]provider.Transaction, error) {
	q := url.Values{}
	if !from.IsZero() {
		q.Set("date_from", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		q.Set("date_to", to.Format("2006-01-02"))
	}
	endpoint := "/accounts/" + url.PathEscape(accountID) + "/transactions/"
	if encoded := q.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	var raw transactionsPayload
	if err := c.get(ctx, endpoint, &raw); err != nil {
		return nil, err
	}
	return NormalizeTransactions(accountID, raw)
}

func (c *Client) get(ctx context.Context, endpoint string, out any) error {
	return c.do(ctx, http.MethodGet, endpoint, nil, true, out)
}

func (c *Client) post(ctx context.Context, endpoint string, body any, out any) error {
	return c.do(ctx, http.MethodPost, endpoint, body, true, out)
}

func (c *Client) do(ctx context.Context, method, endpoint string, body any, auth bool, out any) error {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = b
	}
	for attempt := 0; attempt < 2; attempt++ {
		req, err := c.request(ctx, method, endpoint, bodyBytes, auth)
		if err != nil {
			return err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusUnauthorized && auth && attempt == 0 && c.refresh != "" {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			_ = resp.Body.Close()
			if _, err := c.refreshAccess(ctx); err != nil {
				c.clearTokens()
				if _, err := c.newAccess(ctx); err != nil {
					return err
				}
			}
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			return fmt.Errorf("gocardless %s %s: %s: %s", method, endpoint, resp.Status, strings.TrimSpace(string(b)))
		}
		if out == nil {
			return nil
		}
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return errors.New("gocardless request retry exhausted")
}

func (c *Client) request(ctx context.Context, method, endpoint string, body []byte, auth bool) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.resolve(endpoint), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth {
		token, err := c.accessToken(ctx)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	if c.access != "" && time.Now().Before(c.accessExp) {
		return c.access, nil
	}
	if c.refresh != "" && (c.refreshExp.IsZero() || time.Now().Before(c.refreshExp)) {
		access, err := c.refreshAccess(ctx)
		if err == nil {
			return access, nil
		}
		c.clearTokens()
	}
	return c.newAccess(ctx)
}

func (c *Client) newAccess(ctx context.Context) (string, error) {
	if c.secretID == "" || c.secretKey == "" {
		return "", ErrMissingCredentials
	}
	var token tokenPayload
	if err := c.do(ctx, http.MethodPost, "/token/new/", map[string]string{
		"secret_id":  c.secretID,
		"secret_key": c.secretKey,
	}, false, &token); err != nil {
		return "", err
	}
	c.setToken(token)
	if c.access != "" {
		return c.access, nil
	}
	if c.refresh == "" {
		return "", errors.New("gocardless token response missing access token")
	}
	return c.refreshAccess(ctx)
}

func (c *Client) refreshAccess(ctx context.Context) (string, error) {
	if c.refresh == "" {
		return "", errors.New("gocardless refresh token missing")
	}
	oldRefresh := c.refresh
	oldRefreshExp := c.refreshExp
	var refreshed tokenPayload
	if err := c.do(ctx, http.MethodPost, "/token/refresh/", map[string]string{
		"refresh": c.refresh,
	}, false, &refreshed); err != nil {
		return "", err
	}
	if refreshed.Refresh == "" {
		refreshed.Refresh = oldRefresh
	}
	c.setToken(refreshed)
	if refreshed.RefreshExpires <= 0 {
		c.refreshExp = oldRefreshExp
	}
	if c.access == "" {
		return "", errors.New("gocardless refresh response missing access token")
	}
	return c.access, nil
}

func (c *Client) setToken(token tokenPayload) {
	c.access = token.Access
	c.accessExp = token.accessExpiry()
	c.refresh = token.Refresh
	c.refreshExp = token.refreshExpiry()
}

func (c *Client) clearTokens() {
	c.access = ""
	c.accessExp = time.Time{}
	c.refresh = ""
	c.refreshExp = time.Time{}
}

func (c *Client) resolve(endpoint string) string {
	u := *c.baseURL
	endpointPath, rawQuery, _ := strings.Cut(endpoint, "?")
	rawPath := path.Join(strings.TrimRight(c.baseURL.EscapedPath(), "/"), endpointPath)
	if strings.HasSuffix(endpointPath, "/") && !strings.HasSuffix(rawPath, "/") {
		rawPath += "/"
	}
	unescapedPath, err := url.PathUnescape(rawPath)
	if err != nil {
		u.Path = rawPath
		u.RawPath = ""
	} else {
		u.Path = unescapedPath
		u.RawPath = rawPath
	}
	u.RawQuery = rawQuery
	return u.String()
}
