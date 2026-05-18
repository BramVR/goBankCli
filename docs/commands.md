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

Checks config paths and whether provider credentials are present. It reports
only `set`, `missing`, or `default`, never secret values.

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
gobankcli connect --provider gocardless --institution BELFIUS_GKCCBEBB --redirect https://example.test/callback
gobankcli connect --provider enablebanking --institution BE:Belfius --listen 127.0.0.1:28787 --listen-https --listen-cert ~/.config/gobankcli/tls/localhost.crt --listen-key ~/.config/gobankcli/tls/localhost.key
gobankcli connect --provider enablebanking --institution BE:Belfius --redirect https://127.0.0.1:28787/enablebanking/callback
gobankcli connect --provider enablebanking --institution BE:Belfius --listen 127.0.0.1:8787
```

Starts a read-only provider consent flow and stores the returned connection in
the archive. The output includes the provider connection ID and redirect URL.

For Enable Banking, the provider connection ID from this command is the pending
callback `state`. Production Enable Banking applications require an HTTPS
redirect URL, so register `https://127.0.0.1:28787/enablebanking/callback` and
use `--listen 127.0.0.1:28787 --listen-https` for an automatic local callback.
Add `--listen-cert` and `--listen-key` for a locally trusted certificate; without
them, the listener uses an ephemeral self-signed localhost certificate and the
browser may show a certificate warning. With `--listen`, `connect` starts a
loopback callback server, prints the browser URL on stderr, waits for one
callback, validates `state`, exchanges `code`, archives the session/accounts,
and outputs the authorized session report. Use `--callback-timeout` to control
how long it waits. Because it waits for browser consent, `--listen` is rejected
with `--no-input`.

## authorize

```bash
gobankcli authorize --provider enablebanking --url "https://127.0.0.1:28787/enablebanking/callback?code=CODE&state=STATE" --institution BE:Belfius
gobankcli authorize --provider enablebanking --code CODE --state STATE --institution BE:Belfius
```

Exchanges an Enable Banking callback code for a usable session connection and
archives returned accounts. `state` must match a pending connection created by
`connect`. `--institution` is required when the provider response omits ASPSP
metadata.

## accounts

```bash
gobankcli accounts --provider enablebanking --connection SESSION_ID
gobankcli --json accounts --provider enablebanking --connection SESSION_ID
gobankcli accounts --provider gocardless --connection REQUISITION_ID
```

Fetches accounts for a provider connection, upserts them into SQLite, and emits
the normalized account records plus a count.

Enable Banking account listing uses archived accounts when the live session
response only contains provider UIDs.

## sync

```bash
gobankcli sync --provider enablebanking --connection SESSION_ID
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01 --to 2026-01-31
gobankcli sync --provider gocardless --connection REQUISITION_ID --from 2026-01-01
```

Fetches accounts and transactions for a provider connection, archives normalized
transactions, and records one sync run per account. Dates are booking-date
filters in `YYYY-MM-DD` format. Machine-readable output reports the provider
connection as `provider_connection_id` and the archived local ID as
`connection_id`.

For Enable Banking, use the session ID printed as `provider_connection_id` as
`--connection`. Manual callback flows get that ID from `authorize`; automatic
local callback flows get it directly from `connect --listen`.

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
