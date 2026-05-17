package csvexport

import (
	"encoding/csv"
	"io"
)

var Header = []string{
	"date",
	"value_date",
	"account_id",
	"iban",
	"institution",
	"counterparty_name",
	"counterparty_account",
	"description",
	"amount",
	"currency",
	"transaction_id",
	"provider",
	"category",
}

type Row struct {
	Date                string
	ValueDate           string
	AccountID           string
	IBAN                string
	Institution         string
	CounterpartyName    string
	CounterpartyAccount string
	Description         string
	Amount              string
	Currency            string
	TransactionID       string
	Provider            string
	Category            string
}

func Write(w io.Writer, rows []Row) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(Header); err != nil {
		return err
	}
	for _, row := range rows {
		if err := cw.Write([]string{
			row.Date,
			row.ValueDate,
			row.AccountID,
			row.IBAN,
			row.Institution,
			row.CounterpartyName,
			row.CounterpartyAccount,
			row.Description,
			row.Amount,
			row.Currency,
			row.TransactionID,
			row.Provider,
			row.Category,
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
