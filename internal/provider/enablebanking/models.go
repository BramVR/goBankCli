package enablebanking

import (
	"encoding/json"
	"fmt"
)

type aspspsPayload struct {
	ASPSPs []aspspPayload `json:"aspsps"`
}

type aspspPayload struct {
	Name                   string          `json:"name"`
	Country                string          `json:"country"`
	PSUTypes               []string        `json:"psu_types"`
	MaximumConsentValidity int64           `json:"maximum_consent_validity"`
	Raw                    json.RawMessage `json:"-"`
}

func (p *aspspPayload) UnmarshalJSON(b []byte) error {
	type alias aspspPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = aspspPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type authPayload struct {
	URL        string          `json:"url"`
	State      string          `json:"state"`
	ValidUntil string          `json:"valid_until"`
	Raw        json.RawMessage `json:"-"`
}

func (p *authPayload) UnmarshalJSON(b []byte) error {
	type alias authPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = authPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type sessionPayload struct {
	SessionID string           `json:"session_id"`
	Status    string           `json:"status"`
	Accounts  []accountPayload `json:"accounts"`
	Access    accessPayload    `json:"access"`
	ASPSP     aspspPayload     `json:"aspsp"`
	Raw       json.RawMessage  `json:"-"`
}

func (p *sessionPayload) UnmarshalJSON(b []byte) error {
	type alias sessionPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = sessionPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type accessPayload struct {
	ValidUntil string `json:"valid_until"`
}

type accountPayload struct {
	UID                string                `json:"uid"`
	IdentificationHash string                `json:"identification_hash"`
	AccountID          accountIdentification `json:"account_id"`
	Name               string                `json:"name"`
	Details            string                `json:"details"`
	Currency           string                `json:"currency"`
	Raw                json.RawMessage       `json:"-"`
}

func (p *accountPayload) UnmarshalJSON(b []byte) error {
	var uid string
	if err := json.Unmarshal(b, &uid); err == nil {
		p.UID = uid
		p.Raw = append(p.Raw[:0], b...)
		return nil
	}
	type alias accountPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = accountPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type accountIdentification struct {
	IBAN  string                     `json:"iban"`
	Other otherAccountIdentification `json:"other"`
}

type otherAccountIdentification struct {
	Identification string `json:"identification"`
	SchemeName     string `json:"scheme_name"`
	Issuer         string `json:"issuer"`
}

func (a accountIdentification) reference() string {
	if a.IBAN != "" {
		return a.IBAN
	}
	if a.Other.Identification != "" {
		if a.Other.SchemeName != "" {
			return fmt.Sprintf("%s:%s", a.Other.SchemeName, a.Other.Identification)
		}
		return a.Other.Identification
	}
	return ""
}

type transactionsPayload struct {
	Transactions    []transactionPayload `json:"transactions"`
	ContinuationKey string               `json:"continuation_key"`
}

type transactionPayload struct {
	TransactionID         string                `json:"transaction_id"`
	EntryReference        string                `json:"entry_reference"`
	TransactionAmount     amountPayload         `json:"transaction_amount"`
	Creditor              partyPayload          `json:"creditor"`
	CreditorAccount       accountIdentification `json:"creditor_account"`
	Debtor                partyPayload          `json:"debtor"`
	DebtorAccount         accountIdentification `json:"debtor_account"`
	CreditDebitIndicator  string                `json:"credit_debit_indicator"`
	Status                string                `json:"status"`
	BookingDate           string                `json:"booking_date"`
	ValueDate             string                `json:"value_date"`
	TransactionDate       string                `json:"transaction_date"`
	RemittanceInformation []string              `json:"remittance_information"`
	Note                  string                `json:"note"`
	ReferenceNumber       string                `json:"reference_number"`
	Raw                   json.RawMessage       `json:"-"`
}

func (p *transactionPayload) UnmarshalJSON(b []byte) error {
	type alias transactionPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = transactionPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type amountPayload struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type partyPayload struct {
	Name string `json:"name"`
}
