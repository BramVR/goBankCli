package store

import (
	"context"
	"time"

	"gobankcli/internal/csvexport"
)

type TransactionFilter struct {
	From      *time.Time
	To        *time.Time
	AccountID string
}

func (s *Store) ListTransactions(ctx context.Context, filter TransactionFilter) ([]csvexport.Row, error) {
	from := dateFilter(filter.From)
	to := dateFilter(filter.To)
	rows, err := s.db.QueryContext(ctx, `
select
	t.booking_date,
	coalesce(t.value_date, ''),
	t.account_id,
	coalesce(a.iban, ''),
	coalesce(i.name, ''),
	coalesce(t.counterparty_name, ''),
	coalesce(t.counterparty_account, ''),
	coalesce(t.description, ''),
	t.amount,
	t.currency,
	t.id,
	t.provider
from transactions t
left join accounts a on a.id = t.account_id
left join institutions i on i.id = a.institution_id
where (? = '' or t.booking_date >= ?)
  and (? = '' or t.booking_date <= ?)
  and (? = '' or t.account_id = ?)
order by t.booking_date asc, t.id asc`,
		from, from, to, to, filter.AccountID, filter.AccountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []csvexport.Row
	for rows.Next() {
		var row csvexport.Row
		if err := rows.Scan(
			&row.Date,
			&row.ValueDate,
			&row.AccountID,
			&row.IBAN,
			&row.Institution,
			&row.CounterpartyName,
			&row.CounterpartyAccount,
			&row.Description,
			&row.Amount,
			&row.Currency,
			&row.TransactionID,
			&row.Provider,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func dateFilter(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
