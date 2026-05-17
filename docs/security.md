---
summary: "Security model, local data handling, credential rules, and unsupported risky flows."
read_when:
  - "Changing credential handling, provider auth, file permissions, or archive storage."
  - "Reviewing safety constraints for bank data, scraping, or payment behavior."
---
# Security

`gobankcli` is read-only account information tooling. It uses provider AIS
flows, stores a private local archive, and exports data only when explicitly
requested.

## Supported

- GoCardless Bank Account Data read-only access.
- Local SQLite archive with private file permissions.
- Normalized CSV export to configured paths or explicit `--out` paths.
- Environment-variable credentials for provider API access.

## Not Supported

- Bank-login scraping.
- Browser automation or session-cookie capture.
- Reverse-engineered private bank endpoints.
- Payment initiation.
- Bank password storage.
- Public dashboard or local web server.

## Credentials

GoCardless credentials come from:

```bash
GOBANKCLI_GOCARDLESS_SECRET_ID
GOBANKCLI_GOCARDLESS_SECRET_KEY
```

`doctor` reports only `set` or `missing`. Config files never contain provider
secrets, and docs/tests must not include real-looking credentials.

## Local Data

The archive contains private bank metadata and transactions. Default paths are:

- database: `~/.local/share/gobankcli/gobankcli.db`
- exports: `~/Finance/gobankcli/exports`

SQLite database, WAL, and SHM files are chmodded to `0600`. Config and export
directories are created with `0700`; config and export files are written with
`0600`.

## Raw JSON

Raw provider JSON is preserved in SQLite so normalization can improve later.
It may include private account or transaction metadata. Command reports omit raw
payloads unless a local read-only SQL query explicitly selects them.

## Consent Renewal

GoCardless consents can expire or be revoked. Renewal should use the provider
consent flow again through `connect`, then `accounts` and `sync`. The CLI must
not fake successful live syncs when credentials or valid consent are missing.
