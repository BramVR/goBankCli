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
