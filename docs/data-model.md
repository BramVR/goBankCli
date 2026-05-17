---
summary: "SQLite tables, transaction dedupe rules, and normalized CSV schema."
read_when:
  - "Changing archive schema, transaction identity, or CSV exports."
  - "Working on store migrations or export queries."
---
# Data Model

The SQLite archive stores normalized rows plus raw provider JSON where useful.

Core tables:

- `institutions`: provider institution IDs, names, countries, BICs, raw JSON.
- `connections`: read-only consent/requisition state and expiry metadata.
- `accounts`: provider account IDs, IBAN/name/currency/owner metadata.
- `transactions`: normalized transaction fields and raw provider payloads.
- `sync_runs`: sync attempts, status, errors, and transaction counters.

Transaction amounts are text decimal strings. Do not store or compute money with
`float64`.

Transaction deduplication order:

- provider/account transaction ID
- provider/account reference
- stable hash of provider, account, booking date, amount, currency,
  counterparty, and description

CSV exports use this stable header:

```csv
date,value_date,account_id,iban,institution,counterparty_name,counterparty_account,description,amount,currency,transaction_id,provider,category
```

Local `query`/`sql` commands allow read-only `SELECT`/`WITH` inspection of this
schema. Mutating SQL and multiple statements are rejected before execution.
