---
summary: "gobankcli export command reference."
read_when:
  - "Using or changing gobankcli export."
  - "Updating command flags, usage, or output behavior."
---

# gobankcli export

Export normalized transactions as CSV.

## Usage

```bash
gobankcli export [flags]
```

## Flags

- `--config` (`string`, default none): Config file path.
- `--db` (`string`, default none): SQLite archive path.
- `--json` (`bool`, default `false`): Emit stable JSON.
- `--plain` (`bool`, default `false`): Emit simple parseable plain text.
- `--no-input` (`bool`, default `false`): Never prompt or wait for input.
- `--version` (`bool`, default `false`): Print version and exit.
- `--from` (`string`, default none): Start booking date, inclusive, as YYYY-MM-DD.
- `--to` (`string`, default none): End booking date, inclusive, as YYYY-MM-DD.
- `--account` (`string`, default none): Restrict export to one local account ID.
- `--out` (`string`, default none): CSV output path. Use - for stdout.
