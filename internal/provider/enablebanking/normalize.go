package enablebanking

import (
	"strings"
	"time"

	"gobankcli/internal/provider"
)

func NormalizeInstitutions(raw []aspspPayload) []provider.Institution {
	out := make([]provider.Institution, 0, len(raw))
	for _, item := range raw {
		out = append(out, NormalizeInstitution(item))
	}
	return out
}

func NormalizeInstitution(raw aspspPayload) provider.Institution {
	country := strings.ToUpper(strings.TrimSpace(raw.Country))
	name := strings.TrimSpace(raw.Name)
	return provider.Institution{
		Provider:              Name,
		ProviderInstitutionID: InstitutionID(country, name),
		Name:                  name,
		Country:               country,
		RawJSON:               copyRaw(raw.Raw),
	}
}

func NormalizeConnection(session sessionPayload) provider.Connection {
	expiresAt := parseTimePtr(session.Access.ValidUntil)
	status := firstNonEmpty(session.Status, "AUTHORIZED")
	now := time.Now().UTC()
	return provider.Connection{
		Provider:             Name,
		ProviderConnectionID: session.SessionID,
		InstitutionID:        InstitutionID(session.ASPSP.Country, session.ASPSP.Name),
		Status:               status,
		CreatedAt:            now,
		UpdatedAt:            now,
		ExpiresAt:            expiresAt,
		RawJSON:              copyRaw(session.Raw),
	}
}

var ErrMissingStableAccountID = provider.ErrMissingStableAccountID

func NormalizeAccounts(session sessionPayload) ([]provider.Account, error) {
	out := make([]provider.Account, 0, len(session.Accounts))
	for _, account := range session.Accounts {
		normalized, err := NormalizeAccount(session, account)
		if err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
}

func NormalizeAccount(session sessionPayload, raw accountPayload) (provider.Account, error) {
	stableID := stableAccountID(raw)
	if stableID == "" {
		return provider.Account{}, ErrMissingStableAccountID
	}
	name := joinNonEmpty([]string{raw.Name, raw.Details}, " | ")
	return provider.Account{
		Provider:           Name,
		ProviderAccountID:  stableID,
		ProviderResourceID: raw.UID,
		InstitutionID:      InstitutionID(session.ASPSP.Country, session.ASPSP.Name),
		ConnectionID:       session.SessionID,
		IBAN:               raw.AccountID.IBAN,
		Name:               name,
		Currency:           raw.Currency,
		RawJSON:            copyRaw(raw.Raw),
	}, nil
}

func stableAccountID(raw accountPayload) string {
	return firstNonEmpty(raw.IdentificationHash, raw.AccountID.reference())
}

func NormalizeTransactions(accountID string, raw transactionsPayload) ([]provider.Transaction, error) {
	out := make([]provider.Transaction, 0, len(raw.Transactions))
	for _, item := range raw.Transactions {
		if strings.ToUpper(strings.TrimSpace(item.Status)) != "BOOK" {
			continue
		}
		tx, err := NormalizeTransaction(accountID, item)
		if err != nil {
			return nil, err
		}
		out = append(out, tx)
	}
	return out, nil
}

func NormalizeTransaction(accountID string, raw transactionPayload) (provider.Transaction, error) {
	bookingDate, err := parseDate(firstNonEmpty(raw.BookingDate, raw.ValueDate, raw.TransactionDate))
	if err != nil {
		return provider.Transaction{}, err
	}
	var valueDate *time.Time
	if strings.TrimSpace(raw.ValueDate) != "" {
		parsed, err := parseDate(raw.ValueDate)
		if err != nil {
			return provider.Transaction{}, err
		}
		valueDate = &parsed
	}
	counterpartyName, counterpartyAccount := counterparty(raw)
	remittance := joinNonEmpty(append(raw.RemittanceInformation, raw.Note), " | ")
	return provider.Transaction{
		Provider:              Name,
		ProviderTransactionID: firstNonEmpty(raw.TransactionID, raw.EntryReference),
		AccountID:             accountID,
		BookingDate:           bookingDate,
		ValueDate:             valueDate,
		Amount:                signedAmount(raw.TransactionAmount.Amount, raw.CreditDebitIndicator),
		Currency:              raw.TransactionAmount.Currency,
		CounterpartyName:      counterpartyName,
		CounterpartyAccount:   counterpartyAccount,
		Description:           remittance,
		RemittanceInfo:        remittance,
		Reference:             firstNonEmpty(raw.ReferenceNumber, raw.EntryReference),
		RawJSON:               copyRaw(raw.Raw),
	}, nil
}

func InstitutionID(country, name string) string {
	country = strings.ToUpper(strings.TrimSpace(country))
	name = strings.TrimSpace(name)
	if country == "" || name == "" {
		return ""
	}
	return country + ":" + name
}

func ParseInstitutionID(id string) (string, string, error) {
	country, name, ok := strings.Cut(strings.TrimSpace(id), ":")
	if !ok || strings.TrimSpace(country) == "" || strings.TrimSpace(name) == "" {
		return "", "", ErrInvalidInstitutionID
	}
	return strings.ToUpper(strings.TrimSpace(country)), strings.TrimSpace(name), nil
}

func signedAmount(amount, indicator string) string {
	amount = strings.TrimSpace(amount)
	if strings.ToUpper(strings.TrimSpace(indicator)) != "DBIT" || amount == "" || strings.HasPrefix(amount, "-") {
		return amount
	}
	return "-" + amount
}

func counterparty(raw transactionPayload) (string, string) {
	if strings.ToUpper(strings.TrimSpace(raw.CreditDebitIndicator)) == "DBIT" {
		return raw.Creditor.Name, raw.CreditorAccount.reference()
	}
	return raw.Debtor.Name, raw.DebtorAccount.reference()
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func parseTimePtr(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func joinNonEmpty(values []string, sep string) string {
	var out []string
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, sep)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func copyRaw(raw []byte) []byte {
	if len(raw) == 0 {
		return nil
	}
	return append([]byte(nil), raw...)
}
