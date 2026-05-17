---
summary: "Core architecture spine, package boundaries, and local archive design."
read_when:
  - "Changing provider, store, export, or CLI boundaries."
  - "Checking the bank-agnostic design constraints."
---
# Architecture

`gobankcli` uses one spine:

```text
read-only provider -> normalized model -> SQLite archive -> export/query -> CLI
```

Providers are bank-agnostic. GoCardless Bank Account Data is the first concrete
provider, but commands should depend on provider interfaces, not bank-specific
code.

The provider package owns generic institutions, connections, accounts,
transactions, and sync runs. Concrete providers normalize their API payloads
into those models before store/export code sees them.

SQLite is the local archive because it is durable, inspectable, scriptable, and
works well for private single-user data. Raw provider JSON will be preserved
beside normalized rows so normalization can improve without losing source data.

Schema migrations are local and monotonic. The initial archive stores
institutions, consent connections, accounts, transactions, and sync runs.
Transactions have a stable dedupe key based on provider/account transaction ID,
then reference, then a hash of normalized transaction fields.

The product evolves by adding capabilities. Do not split the design into hard
version phases.
