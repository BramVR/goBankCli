---
summary: "Copy-paste setup, GoCardless consent, sync, export, cron, and output examples."
read_when:
  - "Updating quickstart flows, examples, or automation snippets."
  - "Checking user-facing behavior for JSON, plain, sync, or export commands."
---
# Examples

These snippets use placeholders and synthetic dates. Do not paste real bank data, copied bank exports, live credentials, or callback URLs containing authorization parameters into docs, tickets, logs, or commits.

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

Set Enable Banking credentials for the recommended live provider:

```bash
source ~/.config/gobankcli/enablebanking.env
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
  --provider gocardless \
  --institution BELFIUS_GKCCBEBB \
  --redirect https://example.test/callback
```

Open the returned redirect URL, finish consent with the provider, then use the
returned provider connection ID for account and sync commands.

For Enable Banking production, register
`https://127.0.0.1:28787/enablebanking/callback` in the Enable Banking app and
use the HTTPS local callback listener:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --listen 127.0.0.1:28787 \
  --listen-https \
  --listen-cert ~/.config/gobankcli/tls/localhost.crt \
  --listen-key ~/.config/gobankcli/tls/localhost.key
```

Use a locally trusted certificate to avoid browser certificate warnings during
the bank redirect. Without `--listen-cert` and `--listen-key`, the CLI uses an
ephemeral self-signed certificate and the browser may show a warning. If Enable
Banking rejects an IP-literal redirect URL, use
`https://localhost:28787/enablebanking/callback` and run the listener with
`--listen localhost:28787 --listen-https`.

If the provider accepts an HTTP loopback redirect, for example in sandbox, TLS
is optional:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --listen 127.0.0.1:8787
```

Manual fallback:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --redirect https://127.0.0.1:28787/enablebanking/callback

gobankcli authorize \
  --provider enablebanking \
  --url "https://127.0.0.1:28787/enablebanking/callback?code=CODE&state=STATE" \
  --institution BE:Belfius
```

Use the returned session ID for account and sync commands.

## Sync

```bash
gobankcli accounts --provider gocardless --connection REQUISITION_ID
gobankcli sync --provider gocardless --connection REQUISITION_ID --from 2026-01-01
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01
gobankcli status
```

Use `--from` and `--to` to restrict booking dates.

Only booked transactions are archived. Pending transactions are ignored for now.

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
GOBANKCLI_ENABLEBANKING_APP_ID=... \
GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH=~/.config/gobankcli/enablebanking.pem \
gobankcli --no-input sync --provider enablebanking --connection SESSION_ID --from 2026-01-01

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
