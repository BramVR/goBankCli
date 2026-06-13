---
summary: "gobankcli authorize command reference."
read_when:
  - "Using or changing gobankcli authorize."
  - "Updating command flags, usage, or output behavior."
---

# gobankcli authorize

Exchange a provider callback code for a usable connection.

## Usage

```bash
gobankcli authorize [flags]
```

## Flags

- `--config` (`string`, default none): Config file path.
- `--db` (`string`, default none): SQLite archive path.
- `--json` (`bool`, default `false`): Emit stable JSON.
- `--plain` (`bool`, default `false`): Emit simple parseable plain text.
- `--no-input` (`bool`, default `false`): Never prompt or wait for input.
- `--version` (`bool`, default `false`): Print version and exit.
- `--provider` (`string`, default none): Provider name.
- `--code` (`string`, default none): Callback authorization code.
- `--state` (`string`, default none): Callback state.
- `--url` (`string`, default none): Full callback URL containing the authorization code.
- `--institution` (`string`, default none): Provider institution ID, needed if the provider session response omits it.
