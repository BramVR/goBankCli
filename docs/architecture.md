# Architecture

`gobankcli` uses one spine:

```text
read-only provider -> normalized model -> SQLite archive -> export/query -> CLI
```

Providers are bank-agnostic. GoCardless Bank Account Data is the first concrete
provider, but commands should depend on provider interfaces, not bank-specific
code.

SQLite is the local archive because it is durable, inspectable, scriptable, and
works well for private single-user data. Raw provider JSON will be preserved
beside normalized rows so normalization can improve without losing source data.

The product evolves by adding capabilities. Do not split the design into hard
version phases.
