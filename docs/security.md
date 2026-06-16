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

The short version: your archive is a local file, provider secrets come from environment variables, and public docs/examples must stay synthetic.

## Supported

- GoCardless Bank Account Data read-only access.
- Enable Banking AIS read-only access.
- Short-lived HTTPS loopback callback listener for Enable Banking production
  authorization.
- HTTPS manual callback exchange as a fallback.
- Short-lived HTTP loopback callback listener when the provider accepts HTTP
  loopback redirects.
- Local SQLite archive with private file permissions.
- Normalized CSV export to configured paths or explicit `--out` paths.
- Environment-variable credentials for provider API access.

## Not Supported

- Bank-login scraping.
- Browser automation or session-cookie capture.
- Reverse-engineered private bank endpoints.
- Payment initiation.
- Bank password storage.
- Cloud upload of the archive.
- Real bank data in tests, docs, examples, logs, screenshots, public artifacts, or commits.
- Public dashboard or long-running web server.

## Credentials

GoCardless credentials come from:

```bash
GOBANKCLI_GOCARDLESS_SECRET_ID
GOBANKCLI_GOCARDLESS_SECRET_KEY
```

Enable Banking credentials come from:

```bash
GOBANKCLI_ENABLEBANKING_APP_ID
GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH
GOBANKCLI_ENABLEBANKING_API
```

`doctor` reports only `set`, `missing`, or `default`. Config files never
contain provider secrets, and docs/tests must not include real-looking
credentials.

Enable Banking API overrides must use HTTPS for remote hosts. HTTP overrides are
accepted only for loopback test hosts such as `127.0.0.1` or `localhost`.

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

## Public Artifact Rule

Before publishing docs or PR artifacts, check generated site output, README snippets, fixtures, logs, and screenshots for:

- real names, IBANs, account IDs, card numbers, balances, and transaction descriptions
- provider credentials, PEM keys, callback URLs with `code` or `state`, and session IDs copied from real flows
- private bank endpoints, browser session data, and credential-bearing logs

Safe examples use placeholders such as `SESSION_ID`, `REQUISITION_ID`, `<secret-id>`, and synthetic dates.

## Consent Renewal

Provider consents can expire or be revoked. Renewal should use the provider
consent flow again through `connect`, then provider-specific completion
commands such as Enable Banking `authorize`, then `accounts` or `sync`. The CLI
must not fake successful live syncs when credentials or valid consent are
missing.
