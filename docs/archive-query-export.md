---
summary: "Archive, sync, query, and CSV export workflow for local bank data."
read_when:
  - "Updating archive, query, sync, or export user docs."
  - "Checking public usage examples."
title: "Archive, Query, Export"
description: "Use gobankcli to sync booked transactions, inspect the local SQLite archive, and export stable CSV."
---
# Archive, Query, Export

`gobankcli` keeps provider data in a private local SQLite archive. Live provider commands need credentials and consent; local archive commands such as `status`, `query`, and `export` do not.

## Archive Shape

The archive stores:

- institutions returned by providers
- consent connections
- accounts
- booked transactions
- sync runs
- raw provider JSON where useful for future normalization fixes

Money amounts stay decimal strings. Provider payloads are normalized before storage, and raw JSON is omitted from normal command output unless you explicitly select it with local read-only SQL.

## Sync

```bash
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01
gobankcli sync --provider gocardless --connection REQUISITION_ID --from 2026-01-01
```

Use `--from` and `--to` to restrict booking dates:

```bash
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01 --to 2026-01-31
```

Machine-readable sync output reports both the provider connection as `provider_connection_id` and the local archive connection as `connection_id`.

Only booked transactions are archived. Pending transactions are ignored for now.

## Status

```bash
gobankcli status
gobankcli --json status
gobankcli --plain status
```

`status` opens or creates the local SQLite archive, applies migrations, and prints row counts for institutions, connections, accounts, transactions, and sync runs.

## Export CSV

Export all transactions:

```bash
gobankcli export
```

Export a date range:

```bash
gobankcli export --from 2026-01-01 --to 2026-01-31 --out january.csv
```

Export one local account:

```bash
gobankcli export --account ACCOUNT_ID --out account.csv
```

Stream CSV to stdout:

```bash
gobankcli export --out -
```

Without `--out`, the file is written to `normalized.csv` inside the configured exports directory.

## Query Read-Only SQL

```bash
gobankcli query "select count(*) as transactions from transactions"
gobankcli query "select booking_date, amount, description from transactions order by booking_date desc limit 20"
gobankcli --json query "select provider, count(*) as rows from transactions group by provider"
```

`query` and `sql` run one read-only `SELECT` or `WITH` statement. Mutating SQL and multiple statements are rejected. JSON output contains `columns` plus positional `rows`; default and `--plain` output are tab-separated values.

## Automation

Use `--no-input` for cron and agent runs that must not prompt or block:

```bash
GOBANKCLI_ENABLEBANKING_APP_ID="<enablebanking-application-id>" \
GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH=~/.config/gobankcli/enablebanking.pem \
gobankcli --no-input sync --provider enablebanking --connection SESSION_ID --from 2026-01-01

gobankcli --no-input export --out ~/Finance/gobankcli/exports/normalized.csv
```

Stdout is requested data only. Stderr is for human hints, warnings, progress, and errors.

## Local-Only Boundary

The archive contains private bank metadata and transactions. Keep the database, WAL, SHM files, exports, screenshots, logs, and copied query output out of commits and public artifacts.
