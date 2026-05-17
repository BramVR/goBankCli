<img src="docs/assets/gobankcli-header.png" alt="gobankcli header" />

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
- Enable Banking AIS provider
- institutions, consent connections, accounts, booked transactions, sync runs
- local SQLite archive with raw provider JSON preserved
- normalized transaction CSV export
- read-only `query`/`sql` against the archive

Only booked transactions are archived. Pending transactions from provider
payloads are ignored for now.

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

Find Belfius in a provider institution list:

```bash
gobankcli institutions --country BE --query belfius
gobankcli institutions --provider enablebanking --country BE --query belfius
```

Sync after the GoCardless setup below returns a requisition ID:

```bash
gobankcli accounts --connection REQUISITION_ID
gobankcli sync --connection REQUISITION_ID --from 2026-01-01
gobankcli status
gobankcli export
```

For Enable Banking, use a short-lived local callback listener:

```bash
gobankcli connect --provider enablebanking --institution BE:Belfius --listen 127.0.0.1:8787
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01
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

4. Create a consent/requisition. Set `GOCARDLESS_REDIRECT_URL` to the browser
   landing URL you want GoCardless to use after bank authentication:

```bash
gobankcli connect \
  --institution BELFIUS_GKCCBEBB \
  --redirect "$GOCARDLESS_REDIRECT_URL"
```

GoCardless requires this `redirect` URL when creating a requisition.
`gobankcli` does not read the GoCardless callback; it uses the requisition ID
from `connect` for later `accounts` and `sync` calls. For one-person manual use,
the URL only needs to be a valid URL you recognize after the bank flow. Use your
real landing page if you have one; otherwise use another valid URL you can
identify as the post-consent landing page.

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

## Enable Banking / Belfius Setup

This assumes an Enable Banking application with read-only account information
access, application ID, and downloaded RSA private key. Enable Banking requires
a redirect URL in the authorization request and sends the browser back there
with `code` and `state`. For the CLI listener flow, register this exact redirect
URL in your Enable Banking application:

```text
http://127.0.0.1:8787/enablebanking/callback
```

1. Set credentials:

```bash
export GOBANKCLI_ENABLEBANKING_APP_ID="..."
export GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH="$HOME/.config/gobankcli/enablebanking.pem"
```

Optional API override:

```bash
export GOBANKCLI_ENABLEBANKING_API="https://api.enablebanking.com"
```

2. Find Belfius:

```bash
gobankcli institutions --provider enablebanking --country BE --query belfius
```

Enable Banking institution IDs use `COUNTRY:Name`, for example `BE:Belfius`.

3. Start authorization with a local callback listener:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --listen 127.0.0.1:8787
```

The command prints the browser URL on stderr, waits for one callback on the
loopback listener, validates `state`, exchanges the callback `code`, archives
the session/accounts, then exits. The `provider_connection_id` in the output is
the Enable Banking session ID to use with `sync`. Because this waits for browser
consent, it is not available with `--no-input`.

Use the manual fallback only when you intentionally want to handle the callback
somewhere else, for example on a hosted callback URL you control. In that case,
the `--redirect` value must be a real URL registered in your Enable Banking
application. Set `ENABLEBANKING_REDIRECT_URL` to that exact registered URL:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --redirect "$ENABLEBANKING_REDIRECT_URL"
```

Open `redirect_url`, complete the bank flow, then copy the full callback URL
from the browser address bar into `ENABLEBANKING_CALLBACK_URL`.

4. Exchange a manual callback:

```bash
gobankcli authorize \
  --provider enablebanking \
  --url "$ENABLEBANKING_CALLBACK_URL" \
  --institution BE:Belfius
```

`state` must match the pending connection created by `connect`. The
`--institution` flag is needed when the provider session response omits ASPSP
metadata.

5. Sync:

```bash
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01
```

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

GOBANKCLI_ENABLEBANKING_APP_ID=... \
GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH=~/.config/gobankcli/enablebanking.pem \
gobankcli --no-input sync --provider enablebanking --connection SESSION_ID --from 2026-01-01

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

[[connections]]
name = "Belfius personal via Enable Banking"
provider = "enablebanking"
institution_id = "BE:Belfius"
country = "BE"
```

## Commands

```bash
gobankcli doctor
gobankcli init
gobankcli institutions --country BE --query belfius
gobankcli connect --institution BELFIUS_GKCCBEBB --redirect "$GOCARDLESS_REDIRECT_URL"
gobankcli connect --provider enablebanking --institution BE:Belfius --listen 127.0.0.1:8787
gobankcli authorize --provider enablebanking --url "$ENABLEBANKING_CALLBACK_URL" --institution BE:Belfius
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
- Runs a short-lived loopback callback listener only when `--listen` is used.
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

- Live GoCardless or Enable Banking flows require real credentials and consent.
- Pending transactions are not archived yet.
- `category` is present in the CSV schema but currently empty.
- Future providers are not implemented yet: Ponto, CODA/Codabox/Isabel,
  manual CSV import.
- No Homebrew/release packaging yet.
