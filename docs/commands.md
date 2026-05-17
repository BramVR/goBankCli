---
summary: "CLI commands, global flags, and output behavior."
read_when:
  - "Adding or changing commands, flags, or scriptable output."
  - "Updating user-facing command docs."
---
# Commands

Global flags:

- `--config PATH`
- `--db PATH`
- `--json`
- `--plain`
- `--no-input`
- `--version`

## doctor

```bash
gobankcli doctor
gobankcli --json doctor
```

Checks config paths and whether GoCardless credentials are present. It reports
only `set` or `missing`, never secret values.

## init

```bash
gobankcli init
gobankcli init --force
```

Creates config, database, and export directories. Writes a starter config when
none exists.

## institutions

```bash
gobankcli institutions --country BE
gobankcli institutions --country BE --query belfius
gobankcli --json institutions --country BE
```

Lists provider institutions for an ISO country code and archives the returned
institution metadata locally. `--query` filters by name, BIC, or provider
institution ID after the provider response is normalized.

## connect

```bash
gobankcli connect --institution BELFIUS_GKCCBEBB --redirect https://example.test/callback
```

Starts a read-only provider consent flow and stores the returned connection in
the archive. The output includes the provider connection ID and redirect URL.

## accounts

```bash
gobankcli accounts --connection REQUISITION_ID
gobankcli --json accounts --connection REQUISITION_ID
```

Fetches accounts for a provider connection, upserts them into SQLite, and emits
the normalized account records plus a count.

## sync

```bash
gobankcli sync --connection REQUISITION_ID
gobankcli sync --connection REQUISITION_ID --from 2026-01-01 --to 2026-01-31
```

Fetches accounts and transactions for a provider connection, archives normalized
transactions, and records one sync run per account. Dates are booking-date
filters in `YYYY-MM-DD` format. Machine-readable output reports the provider
connection as `provider_connection_id` and the archived local ID as
`connection_id`.

## status

```bash
gobankcli status
```

Opens or creates the local SQLite archive, applies migrations, and prints row
counts for institutions, connections, accounts, transactions, and sync runs.

## export

```bash
gobankcli export
gobankcli export --from 2026-01-01 --to 2026-01-31 --out january.csv
gobankcli export --account ACCOUNT_ID --out -
```

Exports normalized transaction CSV with a stable header. Without `--out`, the
file is written to `normalized.csv` inside the configured exports directory.

## query / sql

```bash
gobankcli query "select count(*) as transactions from transactions"
gobankcli sql "select booking_date, amount, description from transactions limit 20"
gobankcli --json query "select provider, count(*) as rows from transactions group by provider"
```

Runs one read-only `SELECT` or `WITH` statement against the local SQLite archive.
JSON output contains `columns` plus positional `rows`; default and `--plain`
output are tab-separated values. Mutating SQL and multiple statements are
rejected.
