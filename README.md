# gobankcli

`gobankcli` is a local-first, read-only bank transaction archive. It fetches
bank data through safe provider APIs, stores normalized records in SQLite, and
exports stable CSV for budgeting and accounting.

It is built for terminals, shell scripts, cron, and coding agents:

- predictable `--json` and `--plain` output on stdout
- human hints and warnings on stderr
- local SQLite archive
- normalized CSV export
- read-only SQL inspection
- no scraping, payment initiation, bank password storage, or hard-coded secrets
- no `float64` money values

## Current Scope

- generic provider abstraction
- GoCardless Bank Account Data provider
- institutions, consent connections, accounts, booked transactions, sync runs
- local SQLite archive with raw provider JSON preserved
- normalized transaction CSV export
- read-only `query`/`sql` against the archive

Pending transactions are not archived yet. GoCardless transaction payloads can
include both `booked` and `pending`; `gobankcli` currently stores booked
transactions only.

## Install

Build from source:

```bash
git clone https://github.com/BramVR/goBankCli.git
cd goBankCli
make build
./bin/gobankcli --help
```

Run without installing:

```bash
go run ./cmd/gobankcli --help
```

## Quick Start

Create local config and inspect the setup:

```bash
gobankcli init
gobankcli doctor
gobankcli --json doctor
```

Find Belfius in the GoCardless institution list:

```bash
gobankcli institutions --country BE --query belfius
```

Sync after a GoCardless consent exists:

```bash
gobankcli accounts --connection REQUISITION_ID
gobankcli sync --connection REQUISITION_ID --from 2026-01-01
gobankcli status
gobankcli export
```

Inspect the local archive:

```bash
gobankcli query "select count(*) as transactions from transactions"
gobankcli query "select booking_date, amount, description from transactions order by booking_date desc limit 20"
```

## GoCardless / Belfius Setup

This assumes you already have access to GoCardless Bank Account Data and can
create user secrets in the GoCardless Bank Account Data portal. See the
[GoCardless Bank Account Data quickstart](https://developer.gocardless.com/bank-account-data/quick-start-guide/)
for the upstream API flow.

1. Set credentials:

```bash
export GOBANKCLI_GOCARDLESS_SECRET_ID="..."
export GOBANKCLI_GOCARDLESS_SECRET_KEY="..."
```

2. Confirm credential presence:

```bash
gobankcli doctor
```

Expected credential fields: `set`. Secret values are never printed.

3. Find Belfius:

```bash
gobankcli institutions --country BE --query belfius
```

Use the provider institution ID from the output, for example
`BELFIUS_GKCCBEBB`.

4. Create a consent/requisition:

```bash
gobankcli connect \
  --institution BELFIUS_GKCCBEBB \
  --redirect https://example.test/callback
```

The redirect URL is where GoCardless sends the browser after bank
authentication. For local/manual testing it only needs to be a valid URL you can
recognize after returning from the bank flow; `gobankcli` does not run a local
web server.

The command prints:

- `provider_connection_id`: the GoCardless requisition ID
- `connection_id`: the local archive ID
- `redirect_url`: URL to open in a browser

5. Open `redirect_url`, complete the bank consent, then list accounts:

```bash
gobankcli accounts --connection PROVIDER_CONNECTION_ID
```

6. Sync transactions:

```bash
gobankcli sync --connection PROVIDER_CONNECTION_ID --from 2026-01-01
```

7. Verify and export:

```bash
gobankcli status
gobankcli export --out ~/Finance/gobankcli/exports/normalized.csv
```

If credentials are missing, live provider commands fail with
`gocardless credentials missing`. Local archive commands such as `status`,
`export`, and `query` do not need live credentials.

## Output And Automation

Use `--json` for structured output:

```bash
gobankcli --json status
gobankcli --json query "select count(*) as transactions from transactions"
```

Use `--plain` for simple parseable output:

```bash
gobankcli --plain doctor
gobankcli --plain status
```

Use `--no-input` for cron and agent runs:

```bash
GOBANKCLI_GOCARDLESS_SECRET_ID=... \
GOBANKCLI_GOCARDLESS_SECRET_KEY=... \
gobankcli --no-input sync --connection PROVIDER_CONNECTION_ID --from 2026-01-01

gobankcli --no-input export --out ~/Finance/gobankcli/exports/normalized.csv
```

Stdout is for requested data. Stderr is for human hints, warnings, and errors.

## Local Defaults

- config: `~/.config/gobankcli/config.toml`
- database: `~/.local/share/gobankcli/gobankcli.db`
- exports: `~/Finance/gobankcli/exports`

Override paths when needed:

```bash
gobankcli --config /tmp/gobankcli.toml doctor
gobankcli --db /tmp/gobankcli.db status
gobankcli export --out /tmp/transactions.csv
```

Example config:

```toml
default_provider = "gocardless"
default_country = "BE"

[paths]
db = "~/.local/share/gobankcli/gobankcli.db"
exports = "~/Finance/gobankcli/exports"

[[connections]]
name = "Belfius personal"
provider = "gocardless"
institution_id = "BELFIUS_GKCCBEBB"
country = "BE"
```

## Commands

```bash
gobankcli doctor
gobankcli init
gobankcli institutions --country BE --query belfius
gobankcli connect --institution BELFIUS_GKCCBEBB --redirect https://example.test/callback
gobankcli accounts --connection PROVIDER_CONNECTION_ID
gobankcli sync --connection PROVIDER_CONNECTION_ID --from 2026-01-01
gobankcli status
gobankcli export --from 2026-01-01 --to 2026-01-31 --out january.csv
gobankcli query "select count(*) as transactions from transactions"
gobankcli sql "select booking_date, amount, description from transactions limit 20"
```

Use command help for flags:

```bash
gobankcli sync --help
gobankcli export --help
```

## Safety Model

- Uses read-only provider rails.
- Does not scrape bank websites.
- Does not automate browsers or capture sessions.
- Does not store bank passwords.
- Does not initiate payments.
- Does not run a dashboard or public web server.
- Keeps bank data under configured local paths unless `--out` explicitly points
  elsewhere.
- Writes SQLite/config/export files with restrictive permissions.

The archive contains private bank data. Keep it local and out of commits,
shared logs, and broad backups unless that is intentional.

## Development

```bash
make fmt
make test
make lint
make ci
```

Docs:

- [architecture](docs/architecture.md)
- [commands](docs/commands.md)
- [configuration](docs/configuration.md)
- [data model](docs/data-model.md)
- [examples](docs/examples.md)
- [providers](docs/providers.md)
- [security](docs/security.md)

## Known Gaps

- Live GoCardless/Belfius flow requires real credentials and consent.
- Pending transactions are not archived yet.
- `category` is present in the CSV schema but currently empty.
- Future providers are not implemented yet: Ponto, CODA/Codabox/Isabel, manual
  CSV import.
- No Homebrew/release packaging yet.
