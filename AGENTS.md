Read ../agent-scripts/AGENTS.MD before anything.

# gobankcli

Purpose: local-first, read-only bank transaction archive CLI.

## Structure

- `cmd/gobankcli`: binary entrypoint.
- `internal/cmd`: CLI commands and output behavior.
- `internal/config`: config paths and secret presence checks.
- `internal/outfmt`: JSON/plain/human output.
- `docs`: practical operator/developer docs.
- `examples`: safe example config/export fixtures only.

## Commands

- `make build`: build `bin/gobankcli`.
- `make test`: run `go test ./...`.
- `make lint`: run `go vet ./...`.
- `make ci`: format, lint, test.
- `go run ./cmd/gobankcli --help`: quick CLI smoke test.

## CLI Rules

- stdout: requested data only.
- stderr: hints, warnings, progress.
- `--json`: stable machine-readable JSON.
- `--plain`: simple parseable text.
- `--no-input`: no prompts, no blocking reads.
- Non-zero exit for real failures.

## Safety Rules

- No scraping, browser automation, private bank endpoints, or session capture.
- No payment initiation.
- No bank password storage.
- No hard-coded credentials.
- No real bank data in tests, docs, examples, logs, or commits.
- No `float64` for money. Use decimal strings or integer minor units.
- Never print secrets; report only presence/absence.
- Keep bank data inside configured paths unless user explicitly chooses export path.

## Adding Providers

- Add provider-neutral models/interfaces first.
- Add concrete provider under `internal/provider/<name>`.
- Normalize into generic models before storing.
- Preserve raw provider JSON where useful.
- Add testdata with synthetic or provider sample payloads only.

## Adding Commands

- Inspect existing command/output patterns first.
- Keep command structs small; put business logic in provider/store/export packages.
- Add JSON/plain coverage for scriptable commands.
- Update `README.md` and `docs/commands.md` when behavior changes.

## Workflow

- Inspect before editing.
- Keep changes narrow and verified.
- Add or update tests with behavior changes.
- Run `go test ./...`; run `make ci` before handoff/commit.
- Commit with `committer "<conventional message>" <paths...>`.
