---
summary: "gobankcli sql command reference."
read_when:
  - "Using or changing gobankcli sql."
  - "Updating command flags, usage, or output behavior."
---

# gobankcli sql

Alias for query.

## Usage

```bash
gobankcli sql <sql> [flags]
```

## Arguments

- `<sql>`

## Flags

- `--config` (`string`, default none): Config file path.
- `--db` (`string`, default none): SQLite archive path.
- `--json` (`bool`, default `false`): Emit stable JSON.
- `--plain` (`bool`, default `false`): Emit simple parseable plain text.
- `--no-input` (`bool`, default `false`): Never prompt or wait for input.
- `--version` (`bool`, default `false`): Print version and exit.
