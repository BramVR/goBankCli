---
summary: "gobankcli connect command reference."
read_when:
  - "Using or changing gobankcli connect."
  - "Updating command flags, usage, or output behavior."
---

# gobankcli connect

Start a read-only bank consent flow.

## Usage

```bash
gobankcli connect --institution=STRING [flags]
```

## Flags

- `--config` (`string`, default none): Config file path.
- `--db` (`string`, default none): SQLite archive path.
- `--json` (`bool`, default `false`): Emit stable JSON.
- `--plain` (`bool`, default `false`): Emit simple parseable plain text.
- `--no-input` (`bool`, default `false`): Never prompt or wait for input.
- `--version` (`bool`, default `false`): Print version and exit.
- `--provider` (`string`, default none): Provider name.
- `--institution` (`string`, required): Provider institution ID.
- `--redirect` (`string`, default none): Redirect URL registered with the provider.
- `--listen` (`string`, default none): Listen on a loopback address for one provider callback, e.g. 127.0.0.1:8787.
- `--listen-https` (`bool`, default `false`): Serve the local callback listener over HTTPS.
- `--listen-cert` (`string`, default none): TLS certificate path for --listen-https.
- `--listen-key` (`string`, default none): TLS private key path for --listen-https.
- `--callback-timeout` (`duration`, default `5m`): How long --listen waits for the callback.
