---
summary: "gobankcli institutions command reference."
read_when:
  - "Using or changing gobankcli institutions."
  - "Updating command flags, usage, or output behavior."
---

# gobankcli institutions

List provider institutions by country.

## Usage

```bash
gobankcli institutions [flags]
```

## Flags

- `--config` (`string`, default none): Config file path.
- `--db` (`string`, default none): SQLite archive path.
- `--json` (`bool`, default `false`): Emit stable JSON.
- `--plain` (`bool`, default `false`): Emit simple parseable plain text.
- `--no-input` (`bool`, default `false`): Never prompt or wait for input.
- `--version` (`bool`, default `false`): Print version and exit.
- `--provider` (`string`, default none): Provider name.
- `--country` (`string`, default none): ISO country code.
- `--query` (`string`, default none): Case-insensitive name or BIC filter.
