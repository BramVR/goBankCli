---
summary: "Fresh-install path from local config through first sync, status, export, and query."
read_when:
  - "Updating quickstart flows, examples, or first-run docs."
  - "Checking the site-first user journey."
title: "Quickstart"
description: "Walk from a fresh gobankcli build to local config, provider consent, sync, status, CSV export, and read-only SQL."
---
# Quickstart

This path gets from a fresh checkout to a private local archive. Use synthetic dates and placeholders until you are ready to connect a real provider account.

## 1. Build And Initialize

```bash
make build
./bin/gobankcli init
./bin/gobankcli doctor
```

Default local paths:

- config: `~/.config/gobankcli/config.toml`
- database: `~/.local/share/gobankcli/gobankcli.db`
- exports: `~/Finance/gobankcli/exports`

`doctor` prints credential presence as `set`, `missing`, or `default`; secret values are never printed.

## 2. Choose A Provider

Recommended live Belfius path: Enable Banking AIS. It supports a restricted production setup for your own dashboard-linked accounts.

Alternative path: GoCardless Bank Account Data. It is supported for read-only account information and transaction sync, but Enable Banking is the main documented Belfius path.

Before running the live commands below, complete [Provider Setup](provider-setup.md) for your chosen provider. For Enable Banking, that means:

- dashboard application created with the HTTPS loopback redirect URL
- account linked in restricted production
- `GOBANKCLI_ENABLEBANKING_APP_ID` set
- `GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH` points at a readable PEM key
- `~/.config/gobankcli/tls/localhost.crt` and `~/.config/gobankcli/tls/localhost.key` exist for the HTTPS callback listener

```bash
source ~/.config/gobankcli/enablebanking.env
test -r "$GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH"
test -r ~/.config/gobankcli/tls/localhost.crt
test -r ~/.config/gobankcli/tls/localhost.key
```

## 3. Find An Institution

```bash
./bin/gobankcli institutions --provider enablebanking --country BE --query belfius
```

Use the returned provider institution ID in the consent flow. For Enable Banking this is normalized as `COUNTRY:Name`, for example `BE:Belfius`.

## 4. Connect

Enable Banking production uses an HTTPS loopback callback. The provider dashboard must allow the same redirect URL that the CLI listens on:

```text
https://127.0.0.1:28787/enablebanking/callback
```

```bash
./bin/gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --listen 127.0.0.1:28787 \
  --listen-https \
  --listen-cert ~/.config/gobankcli/tls/localhost.crt \
  --listen-key ~/.config/gobankcli/tls/localhost.key \
  --callback-timeout 10m
```

Open the printed browser URL, complete provider and bank authorization, then return to the terminal. The successful report includes:

- `provider_connection_id`: provider session or requisition ID for later live commands.
- `connection_id`: local archive ID.
- `accounts`: archived account count.

## 5. Sync Booked Transactions

```bash
./bin/gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01
./bin/gobankcli status
```

Only booked transactions are archived. Pending transactions from provider payloads are ignored for now.

## 6. Export CSV

```bash
./bin/gobankcli export --out ~/Finance/gobankcli/exports/normalized.csv
```

Use `--from`, `--to`, and `--account` to narrow the export. Use `--out -` to stream CSV to stdout.

## 7. Inspect Locally

```bash
./bin/gobankcli query "select count(*) as transactions from transactions"
./bin/gobankcli query "select booking_date, amount, description from transactions order by booking_date desc limit 20"
```

`query` and `sql` accept one read-only `SELECT` or `WITH` statement. Mutating SQL and multiple statements are rejected.

## Safety Check

`gobankcli` does not scrape bank websites, initiate payments, store bank passwords, perform cloud upload, or include real bank data in examples. The archive stays local unless you explicitly export to a path you choose.
