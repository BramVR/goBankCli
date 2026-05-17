---
summary: "Copy-paste setup, GoCardless consent, sync, export, cron, and output examples."
read_when:
  - "Updating quickstart flows, examples, or automation snippets."
  - "Checking user-facing behavior for JSON, plain, sync, or export commands."
---
# Examples

## First-Time Setup

```bash
make build
./bin/gobankcli init
./bin/gobankcli doctor
```

Set GoCardless credentials before live provider commands:

```bash
export GOBANKCLI_GOCARDLESS_SECRET_ID="..."
export GOBANKCLI_GOCARDLESS_SECRET_KEY="..."
```

Or set Enable Banking credentials:

```bash
export GOBANKCLI_ENABLEBANKING_APP_ID="..."
export GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH="$HOME/.config/gobankcli/enablebanking.pem"
```

## Find Belfius

```bash
gobankcli institutions --country BE --query belfius
gobankcli --json institutions --country BE --query belfius
gobankcli institutions --provider enablebanking --country BE --query belfius
```

Without credentials, live provider commands fail with a provider-specific
missing credentials error.

## Connect A Bank

```bash
gobankcli connect \
  --institution BELFIUS_GKCCBEBB \
  --redirect https://example.test/callback
```

Open the returned redirect URL, finish consent with the provider, then use the
returned provider connection ID for account and sync commands.

For Enable Banking, prefer the local callback listener:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --listen 127.0.0.1:8787
```

Register `http://127.0.0.1:8787/enablebanking/callback` as the app redirect
URL. Use the manual callback flow only when you intentionally want to handle the
callback somewhere else, for example on a hosted callback URL you control:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --redirect https://your-domain.example/callback

gobankcli authorize \
  --provider enablebanking \
  --url "https://your-domain.example/callback?code=CODE&state=STATE" \
  --institution BE:Belfius
```

Use the returned session ID for account and sync commands.

## Sync

```bash
gobankcli accounts --connection REQUISITION_ID
gobankcli sync --connection REQUISITION_ID --from 2026-01-01
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01
gobankcli status
```

Use `--from` and `--to` to restrict booking dates.

## Export

Export all transactions:

```bash
gobankcli export
```

Export transactions for one local account:

```bash
gobankcli export --account ACCOUNT_ID --out account.csv
```

Stream CSV to stdout:

```bash
gobankcli export --out -
```

## Automation

Cron-style sync and export:

```bash
GOBANKCLI_GOCARDLESS_SECRET_ID=... \
GOBANKCLI_GOCARDLESS_SECRET_KEY=... \
gobankcli --no-input sync --connection REQUISITION_ID --from 2026-01-01

gobankcli --no-input export --out ~/Finance/gobankcli/exports/normalized.csv
```

## Scriptable Output

JSON:

```bash
gobankcli --json status
gobankcli --json query "select count(*) as transactions from transactions"
```

Plain:

```bash
gobankcli --plain doctor
gobankcli --plain status
```

Read-only SQL:

```bash
gobankcli query "select booking_date, amount, description from transactions order by booking_date desc limit 20"
```
