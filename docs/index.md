---
summary: "Project Site home page for gobankcli."
read_when:
  - "Changing the public Project Site landing page."
  - "Checking public positioning, canonical links, or docs-site navigation."
---
# gobankcli

`gobankcli` is a local-first, read-only bank transaction archive CLI. It fetches account information through provider APIs, stores normalized records in SQLite, and exports stable CSV for budgeting and accounting.

Use it when you want a private local archive, scriptable terminal output, and read-only SQL inspection without scraping, payment initiation, bank password storage, cloud upload, or real bank data in examples.

## What It Does

- Archives institutions, consent connections, accounts, booked transactions, and sync runs.
- Keeps a local SQLite archive with restrictive file permissions.
- Preserves raw provider JSON where useful while exposing normalized records for queries and CSV.
- Emits predictable `--json` and `--plain` output for scripts and agents.
- Uses official read-only provider API flows for Enable Banking and GoCardless.

## Start Here

- [Install](install.md): install with Homebrew after release, or build from source.
- [Quickstart](quickstart.md): initialize, connect, sync, export, and query.
- [Provider Setup](provider-setup.md): configure Enable Banking or GoCardless.
- [Archive, Query, Export](archive-query-export.md): work with synced data.
- [Security](security.md): review local data handling and explicit non-goals.

## What It Is Not

`gobankcli` is not a bank dashboard, cloud service, payment tool, scraper, browser automation system, or password manager. It does not initiate payments, capture bank sessions, store bank passwords, upload archives, or ship real bank data as examples.

## Source

The project is open source at [BramVR/goBankCli](https://github.com/BramVR/goBankCli). Development notes and local verification commands live in [Development](development.md), and release notes live in [Release](release.md).
