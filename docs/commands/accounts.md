---
summary: "gobankcli accounts command reference."
read_when:
  - "Using or changing gobankcli accounts."
  - "Updating command flags, usage, or output behavior."
---

# gobankcli accounts

Fetch and archive accounts for a connection.

## Usage

```bash
gobankcli accounts --connection=STRING [flags]
```

## Flags

- `--config` (`string`, default none): Config file path.
- `--db` (`string`, default none): SQLite archive path.
- `--json` (`bool`, default `false`): Emit stable JSON.
- `--plain` (`bool`, default `false`): Emit simple parseable plain text.
- `--no-input` (`bool`, default `false`): Never prompt or wait for input.
- `--version` (`bool`, default `false`): Print version and exit.
- `--provider` (`string`, default none): Provider name.
- `--connection` (`string`, required): Provider connection/requisition ID.
