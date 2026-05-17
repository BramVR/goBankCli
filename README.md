# gobankcli

`gobankcli` is a local-first, read-only bank transaction archive. It fetches
bank data through safe provider APIs, stores normalized records in SQLite, and
exports stable CSV for budgeting and accounting.

Safety stance:

- read-only account information only
- no scraping or browser automation
- no payment initiation
- no bank password storage
- no hard-coded secrets
- no `float64` money values

## Build

```bash
make build
go run ./cmd/gobankcli --help
```

## Quickstart

```bash
gobankcli init
gobankcli doctor
gobankcli --json doctor
gobankcli status
gobankcli export
```

GoCardless credentials are read from environment variables:

```bash
export GOBANKCLI_GOCARDLESS_SECRET_ID="..."
export GOBANKCLI_GOCARDLESS_SECRET_KEY="..."
```

Without those variables, live GoCardless commands fail clearly. Local archive
commands and tests do not need live credentials.

## Defaults

- config: `~/.config/gobankcli/config.toml`
- database: `~/.local/share/gobankcli/gobankcli.db`
- exports: `~/Finance/gobankcli/exports`

## Current Commands

- `doctor`: checks config paths and GoCardless credential presence
- `export`: writes normalized transaction CSV
- `init`: writes a starter config and creates local directories
- `status`: shows local archive status

Use `--json` for stable JSON and `--plain` for simple key-value output.
Human hints and warnings go to stderr; requested data goes to stdout.

## Providers

The core provider interface is bank-agnostic. GoCardless Bank Account Data is
the first concrete provider target; Ponto, CODA-style statement providers, and
manual CSV import can be added behind the same interface later.

## Development

```bash
make fmt
make test
make ci
```
