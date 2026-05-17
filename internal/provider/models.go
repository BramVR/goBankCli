package provider

import "time"

type Institution struct {
	ID                    string
	Provider              string
	ProviderInstitutionID string
	Name                  string
	Country               string
	BIC                   string
	RawJSON               []byte
}

type Connection struct {
	ID                   string
	Provider             string
	ProviderConnectionID string
	InstitutionID        string
	Status               string
	RedirectURL          string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ExpiresAt            *time.Time
	RawJSON              []byte
}

type ConnectionSession struct {
	Connection  Connection
	RedirectURL string
}

type Account struct {
	ID                 string
	Provider           string
	ProviderAccountID  string
	ProviderResourceID string
	InstitutionID      string
	ConnectionID       string
	IBAN               string
	Name               string
	Currency           string
	OwnerName          string
	RawJSON            []byte
}

type Transaction struct {
	ID                    string
	Provider              string
	ProviderTransactionID string
	AccountID             string
	BookingDate           time.Time
	ValueDate             *time.Time
	Amount                string
	Currency              string
	CounterpartyName      string
	CounterpartyAccount   string
	Description           string
	RemittanceInfo        string
	Reference             string
	RawJSON               []byte
}

type SyncRun struct {
	ID               string
	Provider         string
	ConnectionID     string
	AccountID        string
	StartedAt        time.Time
	FinishedAt       *time.Time
	Status           string
	Error            string
	TransactionsNew  int64
	TransactionsSeen int64
}
