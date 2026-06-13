---
summary: "CLI command reference generated from the binary."
read_when:
  - "Adding or changing commands, flags, or scriptable output."
  - "Updating user-facing command docs."
---

# Commands

Every user-facing subcommand exposed by `gobankcli`. Regenerate with `make docs-commands`; committed docs must match the CLI surface.

## Subcommands

- [`gobankcli accounts`](./commands/accounts.md): Fetch and archive accounts for a connection.
- [`gobankcli authorize`](./commands/authorize.md): Exchange a provider callback code for a usable connection.
- [`gobankcli connect`](./commands/connect.md): Start a read-only bank consent flow.
- [`gobankcli doctor`](./commands/doctor.md): Check local config, archive, and provider credentials.
- [`gobankcli export`](./commands/export.md): Export normalized transactions as CSV.
- [`gobankcli institutions`](./commands/institutions.md): List provider institutions by country.
- [`gobankcli init`](./commands/init.md): Write a starter config and create local directories.
- [`gobankcli query`](./commands/query.md): Run a read-only SQL query against the local archive.
- [`gobankcli sql`](./commands/sql.md): Alias for query.
- [`gobankcli status`](./commands/status.md): Show local archive status.
- [`gobankcli sync`](./commands/sync.md): Fetch and archive transactions for a connection.

## Discoverability

- [`gobankcli --help`](./commands/help.md): built-in help and per-command help.
- [`gobankcli --version`](./commands/version.md): build version output.
