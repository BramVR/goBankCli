package gocardless

import (
	"strings"
	"time"

	"gobankcli/internal/provider"
)

func NormalizeInstitutions(raw []institutionPayload) ([]provider.Institution, error) {
	out := make([]provider.Institution, 0, len(raw))
	for _, item := range raw {
		out = append(out, NormalizeInstitution(item))
	}
	return out, nil
}

func NormalizeInstitution(raw institutionPayload) provider.Institution {
	country := ""
	if len(raw.Countries) > 0 {
		country = raw.Countries[0]
	}
	return provider.Institution{
		Provider:              Name,
		ProviderInstitutionID: raw.ID,
		Name:                  raw.Name,
		Country:               country,
		BIC:                   raw.BIC,
		RawJSON:               copyRaw(raw.Raw),
	}
}

func NormalizeConnection(raw requisitionPayload) provider.Connection {
	created, _ := time.Parse(time.RFC3339Nano, raw.Created)
	return provider.Connection{
		Provider:             Name,
		ProviderConnectionID: raw.ID,
		InstitutionID:        raw.InstitutionID,
		Status:               raw.Status,
		RedirectURL:          firstNonEmpty(raw.Link, raw.Redirect),
		CreatedAt:            created,
		UpdatedAt:            created,
		RawJSON:              copyRaw(raw.Raw),
	}
}

func NormalizeAccountDetails(accountID, institutionID, connectionID string, raw accountDetailsPayload) provider.Account {
	name := firstNonEmpty(raw.Account.DisplayName, raw.Account.Name)
	return provider.Account{
		Provider:          Name,
		ProviderAccountID: accountID,
		InstitutionID:     institutionID,
		ConnectionID:      connectionID,
		IBAN:              raw.Account.IBAN,
		Name:              name,
		Currency:          raw.Account.Currency,
		OwnerName:         raw.Account.OwnerName,
		RawJSON:           copyRaw(raw.Raw),
	}
}

func NormalizeTransactions(accountID string, raw transactionsPayload) ([]provider.Transaction, error) {
	out := make([]provider.Transaction, 0, len(raw.Transactions.Booked))
	for _, tx := range raw.Transactions.Booked {
		normalized, err := NormalizeTransaction(accountID, tx)
		if err != nil {
			return nil, err
		}
		out = append(out, normalized)
	}
	return out, nil
}

func NormalizeTransaction(accountID string, raw transactionPayload) (provider.Transaction, error) {
	bookingDate, err := time.Parse("2006-01-02", raw.BookingDate)
	if err != nil {
		return provider.Transaction{}, err
	}
	var valueDate *time.Time
	if raw.ValueDate != "" {
		parsed, err := time.Parse("2006-01-02", raw.ValueDate)
		if err != nil {
			return provider.Transaction{}, err
		}
		valueDate = &parsed
	}
	counterpartyName, counterpartyAccount := counterparty(raw)
	remittanceInfo := remittanceText(raw)
	return provider.Transaction{
		Provider:              Name,
		ProviderTransactionID: raw.TransactionID,
		AccountID:             accountID,
		BookingDate:           bookingDate,
		ValueDate:             valueDate,
		Amount:                raw.TransactionAmount.Amount,
		Currency:              raw.TransactionAmount.Currency,
		CounterpartyName:      counterpartyName,
		CounterpartyAccount:   counterpartyAccount,
		Description:           firstNonEmpty(raw.AdditionalInformation, remittanceInfo),
		RemittanceInfo:        remittanceInfo,
		Reference:             firstReference(raw.EntryReference, raw.EndToEndID, raw.RemittanceInformationStructured, firstStructuredReference(raw.RemittanceInformationStructuredArray)),
		RawJSON:               copyRaw(raw.Raw),
	}, nil
}

func counterparty(raw transactionPayload) (string, string) {
	if strings.HasPrefix(strings.TrimSpace(raw.TransactionAmount.Amount), "-") {
		return firstNonEmpty(raw.CreditorName, raw.DebtorName), firstNonEmpty(raw.CreditorAccount.IBAN, raw.CreditorAccount.BBAN, raw.DebtorAccount.IBAN, raw.DebtorAccount.BBAN)
	}
	return firstNonEmpty(raw.DebtorName, raw.CreditorName), firstNonEmpty(raw.DebtorAccount.IBAN, raw.DebtorAccount.BBAN, raw.CreditorAccount.IBAN, raw.CreditorAccount.BBAN)
}

func remittanceText(raw transactionPayload) string {
	if len(raw.RemittanceInformationUnstructuredArray) > 0 {
		return strings.Join(raw.RemittanceInformationUnstructuredArray, " ")
	}
	return raw.RemittanceInformationUnstructured
}

func firstStructuredReference(items []string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
}

func firstReference(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" && !isPlaceholderReference(trimmed) {
			return trimmed
		}
	}
	return ""
}

func isPlaceholderReference(value string) bool {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	switch normalized {
	case "NOTPROVIDED", "NOTAVAILABLE", "NONREF", "NOREF", "N/A", "NA", "NONE", "NULL", "-":
		return true
	default:
		return false
	}
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
