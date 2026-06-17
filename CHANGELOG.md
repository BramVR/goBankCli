# Changelog

## 0.1.1 - Unreleased

- Build future release artifacts with Go 1.26.4 to pick up standard-library security fixes.
- Redact raw provider HTTP error bodies from GoCardless and Enable Banking error strings.
- Reject CSV export destinations that target the active archive database or an existing symlink.
- Reject remote plain-HTTP Enable Banking API overrides before provider requests are signed or sent.
- Reject transaction upserts when the account is blank or missing from the local archive.

## 0.1.0 - 2026-06-14

- Initial public release with local SQLite archive, read-only provider sync, query/export commands, generated docs site, GitHub Release archives, and Homebrew install support.
