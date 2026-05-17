# Commands

Global flags:

- `--config PATH`
- `--db PATH`
- `--json`
- `--plain`
- `--no-input`
- `--version`

## doctor

```bash
gobankcli doctor
gobankcli --json doctor
```

Checks config paths and whether GoCardless credentials are present. It reports
only `set` or `missing`, never secret values.

## init

```bash
gobankcli init
gobankcli init --force
```

Creates config, database, and export directories. Writes a starter config when
none exists.

## status

```bash
gobankcli status
```

Opens or creates the local SQLite archive, applies migrations, and prints row
counts for institutions, connections, accounts, transactions, and sync runs.
