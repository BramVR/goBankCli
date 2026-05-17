package gocardless

import (
	"encoding/json"
	"time"
)

type tokenPayload struct {
	Access         string `json:"access"`
	AccessExpires  int64  `json:"access_expires"`
	Refresh        string `json:"refresh"`
	RefreshExpires int64  `json:"refresh_expires"`
}

func (p tokenPayload) accessExpiry() time.Time {
	if p.AccessExpires <= 0 {
		return time.Now().Add(5 * time.Minute)
	}
	return time.Now().Add(time.Duration(p.AccessExpires) * time.Second).Add(-1 * time.Minute)
}

func (p tokenPayload) refreshExpiry() time.Time {
	if p.RefreshExpires <= 0 {
		return time.Time{}
	}
	return time.Now().Add(time.Duration(p.RefreshExpires) * time.Second).Add(-1 * time.Minute)
}

type institutionPayload struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	BIC       string          `json:"bic"`
	Countries []string        `json:"countries"`
	Raw       json.RawMessage `json:"-"`
}

func (p *institutionPayload) UnmarshalJSON(b []byte) error {
	type alias institutionPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = institutionPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type requisitionPayload struct {
	ID            string          `json:"id"`
	Status        string          `json:"status"`
	InstitutionID string          `json:"institution_id"`
	Redirect      string          `json:"redirect"`
	Link          string          `json:"link"`
	Accounts      []string        `json:"accounts"`
	Created       string          `json:"created"`
	Raw           json.RawMessage `json:"-"`
}

func (p *requisitionPayload) UnmarshalJSON(b []byte) error {
	type alias requisitionPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = requisitionPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type accountDetailsPayload struct {
	Account accountPayload  `json:"account"`
	Raw     json.RawMessage `json:"-"`
}

func (p *accountDetailsPayload) UnmarshalJSON(b []byte) error {
	type alias accountDetailsPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = accountDetailsPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type accountPayload struct {
	ResourceID  string `json:"resourceId"`
	IBAN        string `json:"iban"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Currency    string `json:"currency"`
	OwnerName   string `json:"ownerName"`
}

type transactionsPayload struct {
	Transactions struct {
		Booked  []transactionPayload `json:"booked"`
		Pending []transactionPayload `json:"pending"`
	} `json:"transactions"`
	Raw json.RawMessage `json:"-"`
}

func (p *transactionsPayload) UnmarshalJSON(b []byte) error {
	type alias transactionsPayload
	var v alias
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = transactionsPayload(v)
	p.Raw = append(p.Raw[:0], b...)
	return nil
}

type transactionPayload struct {
	TransactionID                          string          `json:"transactionId"`
	EntryReference                         string          `json:"entryReference"`
	EndToEndID                             string          `json:"endToEndId"`
	BookingDate                            string          `json:"bookingDate"`
	ValueDate                              string          `json:"valueDate"`
	DebtorName                             string          `json:"debtorName"`
	CreditorName                           string          `json:"creditorName"`
	DebtorAccount                          accountRef      `json:"debtorAccount"`
	CreditorAccount                        accountRef      `json:"creditorAccount"`
	AdditionalInformation                  string          `json:"additionalInformation"`
	RemittanceInformationUnstructured      string          `json:"remittanceInformationUnstructured"`
	RemittanceInformationUnstructuredArray []string        `json:"remittanceInformationUnstructuredArray"`
	RemittanceInformationStructured        string          `json:"remittanceInformationStructured"`
	RemittanceInformationStructuredArray   []string        `json:"remittanceInformationStructuredArray"`
	TransactionAmount                      amountPayload   `json:"transactionAmount"`
	Raw                                    json.RawMessage `json:"-"`
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

type accountRef struct {
	IBAN string `json:"iban"`
	BBAN string `json:"bban"`
}

type amountPayload struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}
