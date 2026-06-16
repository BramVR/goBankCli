package enablebanking

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"gobankcli/internal/provider"
)

const (
	Name        = "enablebanking"
	DefaultBase = "https://api.enablebanking.com"

	CredentialApplicationID = "application_id"
	CredentialPrivateKey    = "private_key_path"

	defaultConsentValidity = 90 * 24 * time.Hour
)

var (
	ErrMissingCredentials   = errors.New("enablebanking credentials missing")
	ErrInvalidInstitutionID = errors.New("enablebanking institution id must be COUNTRY:NAME")
	ErrInsecureBaseURL      = errors.New("enablebanking API override must use https unless the host is loopback")
)

type SessionExchanger interface {
	ExchangeSession(ctx context.Context, code string) (provider.ConnectionSession, []provider.Account, error)
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	appID      string
	privateKey *rsa.PrivateKey
	now        func() time.Time
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
	if err := validateBaseURL(u); err != nil {
		return nil, err
	}
	creds := cfg.Credentials
	c := &Client{
		baseURL:    u,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		appID:      creds[CredentialApplicationID],
		now:        func() time.Time { return time.Now().UTC() },
	}
	keyPath := creds[CredentialPrivateKey]
	if c.appID != "" && keyPath != "" {
		key, err := loadPrivateKey(keyPath)
		if err != nil {
			return nil, err
		}
		c.privateKey = key
	}
	return c, nil
}

func validateBaseURL(u *url.URL) error {
	if u.Scheme != "http" {
		return nil
	}
	if isLoopbackHost(u.Hostname()) {
		return nil
	}
	return ErrInsecureBaseURL
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (c *Client) Name() string { return Name }

func (c *Client) ListInstitutions(ctx context.Context, country string) ([]provider.Institution, error) {
	q := url.Values{}
	if strings.TrimSpace(country) != "" {
		q.Set("country", strings.ToUpper(strings.TrimSpace(country)))
	}
	endpoint := "/aspsps"
	if encoded := q.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	var raw aspspsPayload
	if err := c.get(ctx, endpoint, &raw); err != nil {
		return nil, err
	}
	return NormalizeInstitutions(raw.ASPSPs), nil
}

func (c *Client) StartConnection(ctx context.Context, institutionID string, redirectURL string) (provider.ConnectionSession, error) {
	country, name, err := ParseInstitutionID(institutionID)
	if err != nil {
		return provider.ConnectionSession{}, err
	}
	state, err := randomState()
	if err != nil {
		return provider.ConnectionSession{}, err
	}
	consentValidity, err := c.consentValidity(ctx, country, name)
	if err != nil {
		return provider.ConnectionSession{}, err
	}
	validUntil := c.now().Add(consentValidity).Format(time.RFC3339)
	body := map[string]any{
		"access":       map[string]string{"valid_until": validUntil},
		"aspsp":        map[string]string{"name": name, "country": country},
		"state":        state,
		"redirect_url": redirectURL,
		"psu_type":     "personal",
	}
	var raw authPayload
	if err := c.post(ctx, "/auth", body, &raw); err != nil {
		return provider.ConnectionSession{}, err
	}
	return provider.ConnectionSession{
		Connection: provider.Connection{
			Provider:             Name,
			ProviderConnectionID: state,
			InstitutionID:        InstitutionID(country, name),
			Status:               "PENDING",
			RedirectURL:          raw.URL,
			CreatedAt:            c.now(),
			UpdatedAt:            c.now(),
			RawJSON:              copyRaw(raw.Raw),
		},
		RedirectURL: raw.URL,
	}, nil
}

func (c *Client) ExchangeSession(ctx context.Context, code string) (provider.ConnectionSession, []provider.Account, error) {
	var raw sessionPayload
	if err := c.post(ctx, "/sessions", map[string]string{"code": strings.TrimSpace(code)}, &raw); err != nil {
		return provider.ConnectionSession{}, nil, err
	}
	connection := NormalizeConnection(raw)
	accounts, err := NormalizeAccounts(raw)
	if err != nil {
		return provider.ConnectionSession{}, nil, err
	}
	return provider.ConnectionSession{Connection: connection, RedirectURL: connection.RedirectURL}, accounts, nil
}

func (c *Client) GetConnection(ctx context.Context, connectionID string) (provider.Connection, error) {
	var raw sessionPayload
	if err := c.get(ctx, "/sessions/"+url.PathEscape(connectionID), &raw); err != nil {
		return provider.Connection{}, err
	}
	if raw.SessionID == "" {
		raw.SessionID = connectionID
	}
	return NormalizeConnection(raw), nil
}

func (c *Client) ListAccounts(ctx context.Context, connectionID string) ([]provider.Account, error) {
	var raw sessionPayload
	if err := c.get(ctx, "/sessions/"+url.PathEscape(connectionID), &raw); err != nil {
		return nil, err
	}
	if raw.SessionID == "" {
		raw.SessionID = connectionID
	}
	return NormalizeAccounts(raw)
}

func (c *Client) FetchTransactions(ctx context.Context, accountID string, from, to time.Time) ([]provider.Transaction, error) {
	q := url.Values{}
	if !from.IsZero() {
		q.Set("date_from", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		q.Set("date_to", to.Format("2006-01-02"))
	}
	q.Set("transaction_status", "BOOK")
	var transactions []provider.Transaction
	for {
		endpoint := "/accounts/" + url.PathEscape(accountID) + "/transactions?" + q.Encode()
		var raw transactionsPayload
		if err := c.get(ctx, endpoint, &raw); err != nil {
			return nil, err
		}
		page, err := NormalizeTransactions(accountID, raw)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, page...)
		if raw.ContinuationKey == "" {
			return transactions, nil
		}
		q.Set("continuation_key", raw.ContinuationKey)
	}
}

func (c *Client) get(ctx context.Context, endpoint string, out any) error {
	return c.do(ctx, http.MethodGet, endpoint, nil, out)
}

func (c *Client) post(ctx context.Context, endpoint string, body any, out any) error {
	return c.do(ctx, http.MethodPost, endpoint, body, out)
}

func (c *Client) consentValidity(ctx context.Context, country, name string) (time.Duration, error) {
	var raw aspspsPayload
	if err := c.get(ctx, "/aspsps?country="+url.QueryEscape(country), &raw); err != nil {
		return 0, err
	}
	maximum := defaultConsentValidity
	id := InstitutionID(country, name)
	for _, item := range raw.ASPSPs {
		if NormalizeInstitution(item).ProviderInstitutionID != id || item.MaximumConsentValidity <= 0 {
			continue
		}
		aspspMaximum := time.Duration(item.MaximumConsentValidity) * time.Second
		if aspspMaximum < maximum {
			maximum = aspspMaximum
		}
		break
	}
	return maximum, nil
}

func (c *Client) do(ctx context.Context, method, endpoint string, body any, out any) error {
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyBytes = b
	}
	req, err := c.request(ctx, method, endpoint, bodyBytes)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("enablebanking %s %s: %s", method, endpoint, resp.Status)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) request(ctx context.Context, method, endpoint string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.resolve(endpoint), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	token, err := c.jwt()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req, nil
}

func (c *Client) jwt() (string, error) {
	if c.appID == "" || c.privateKey == nil {
		return "", ErrMissingCredentials
	}
	return createJWT(c.appID, c.privateKey, c.now())
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

func randomState() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "gobankcli-" + hex.EncodeToString(b[:]), nil
}
