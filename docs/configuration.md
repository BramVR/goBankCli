---
summary: "Config file, default paths, and GoCardless credential environment variables."
read_when:
  - "Changing config loading, defaults, paths, or credential discovery."
  - "Updating setup instructions."
---
# Configuration

Default paths:

- config: `~/.config/gobankcli/config.toml`
- database: `~/.local/share/gobankcli/gobankcli.db`
- exports: `~/Finance/gobankcli/exports`

Example:

```toml
default_provider = "gocardless"
default_country = "BE"

[paths]
db = "~/.local/share/gobankcli/gobankcli.db"
exports = "~/Finance/gobankcli/exports"
```

GoCardless credentials use environment variables:

```bash
GOBANKCLI_GOCARDLESS_SECRET_ID
GOBANKCLI_GOCARDLESS_SECRET_KEY
```

Secrets are never written into config by `gobankcli init`.
