---
summary: "gobankcli status command reference."
read_when:
  - "Using or changing gobankcli status."
  - "Updating command flags, usage, or output behavior."
---

# gobankcli status

Show local archive status.

## Usage

```bash
gobankcli status [flags]
```

## Flags

- `--config` (`string`, default none): Config file path.
- `--db` (`string`, default none): SQLite archive path.
- `--json` (`bool`, default `false`): Emit stable JSON.
- `--plain` (`bool`, default `false`): Emit simple parseable plain text.
- `--no-input` (`bool`, default `false`): Never prompt or wait for input.
- `--version` (`bool`, default `false`): Print version and exit.
