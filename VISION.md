# Vision

gobankcli is a local-first, read-only bank transaction archive for terminals, scripts, and agents. It should keep provider access safe, archive data locally in SQLite, expose predictable query/export surfaces, and never become a scraper, payment tool, or cloud finance service.

## Merge by Default

- Bug fixes with clear provider, archive, export, config, or CLI cause.
- Small provider, command, JSON/plain output, docs, and examples improvements that keep existing contracts stable.
- Local archive, read-only query, normalized CSV, and migration fixes with tests.
- Security and output-hygiene improvements for credentials, logs, fixtures, and docs.
- Public docs and Homebrew release metadata fixes that match the binary behavior.

## Needs Sign-Off

- New providers, auth flows, scopes, or banking API surfaces.
- Payment initiation, scraping, bank-password storage, cloud upload, or write behavior.
- Changes to money representation, transaction identity, archive schema, or export contracts.
- Live-provider behavior that cannot be tested with the required provider access.
- Real bank data in tests, docs, examples, logs, screenshots, or commits.
