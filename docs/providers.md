---
summary: "Provider interface, provider credentials, and future provider notes."
read_when:
  - "Changing provider behavior or adding provider implementations."
  - "Working on provider normalization or credentials."
---
# Providers

This page describes provider internals and implementation expectations. For user setup, start with [Provider Setup](provider-setup.md).

Providers expose read-only bank data through one generic interface:

- list institutions by country
- start and inspect a consent/connection
- list accounts for a connection
- fetch transactions for an account and date range

Concrete providers normalize API payloads into `internal/provider` models before
the store layer writes SQLite rows. Money amounts stay decimal strings; provider
code must not use `float64`.

## Enable Banking

Enable Banking AIS is the recommended concrete provider for live Belfius use.
Credentials come from:

```bash
GOBANKCLI_ENABLEBANKING_APP_ID
GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH
GOBANKCLI_ENABLEBANKING_API # optional
```

Requests are signed with RS256 JWTs using the application ID as `kid` and the
configured RSA private key. Institution IDs are normalized as `COUNTRY:Name`,
for example `BE:Belfius`.

Enable Banking restricted production requires the accounts to be linked in the
Enable Banking dashboard before the CLI authorization flow. Production redirect
URLs must be HTTPS, so the recommended CLI flow is an HTTPS loopback listener:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --listen 127.0.0.1:28787 \
  --listen-https \
  --listen-cert ~/.config/gobankcli/tls/localhost.crt \
  --listen-key ~/.config/gobankcli/tls/localhost.key
```

Use `https://127.0.0.1:28787/enablebanking/callback` as the registered redirect
URL. The command posts `/auth`, stores pending `state`, waits for the callback,
validates `state`, exchanges `code` through `/sessions`, and archives the
returned session/accounts. The session ID is printed as `provider_connection_id`
and is used for later `sync`.

Manual callback fallback:

```bash
gobankcli connect --provider enablebanking --institution BE:Belfius --redirect https://127.0.0.1:28787/enablebanking/callback
gobankcli authorize --provider enablebanking --url "$ENABLEBANKING_CALLBACK_URL" --institution BE:Belfius
```

Account archive identity uses `identification_hash` or a non-UID account
identifier. The provider UID is stored separately as `provider_resource_id` for
live transaction fetches because UIDs can change across reauthorization.

If a later session response only returns account UIDs, `accounts` and `sync`
reuse the accounts archived during authorization instead of inventing a stable
ID. Without archived stable account metadata, live account listing fails
clearly.

## GoCardless

GoCardless Bank Account Data is also supported. Credentials come from:

```bash
GOBANKCLI_GOCARDLESS_SECRET_ID
GOBANKCLI_GOCARDLESS_SECRET_KEY
```

The provider package contains the live API client plus offline normalization
tests for institutions, consent connections, account details, and booked
transactions. The CLI wires the provider into `institutions`, `connect`,
`accounts`, and `sync`. Pending transactions are not archived yet.

Without credentials, live GoCardless commands must fail clearly and never fake a
successful sync.

## Future Providers

Expected future provider shapes:

- Ponto: read-only PSD2/AIS provider.
- CODA/Codabox/Isabel: statement/archive providers.
- Manual CSV importer: local import provider for downloaded bank exports.

Adding a provider requires:

- provider implementation under `internal/provider/<name>`
- registration in the provider registry
- synthetic or public sample testdata
- normalization tests
- docs for credentials, consent renewal, and limitations
