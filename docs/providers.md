# Providers

Providers expose read-only bank data through one generic interface:

- list institutions by country
- start and inspect a consent/connection
- list accounts for a connection
- fetch transactions for an account and date range

Concrete providers normalize API payloads into `internal/provider` models before
the store layer writes SQLite rows. Money amounts stay decimal strings; provider
code must not use `float64`.

## GoCardless

GoCardless Bank Account Data is the first concrete provider target. Credentials
come from:

```bash
GOBANKCLI_GOCARDLESS_SECRET_ID
GOBANKCLI_GOCARDLESS_SECRET_KEY
```

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
