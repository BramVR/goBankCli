---
summary: "Config file, default paths, and provider credential environment variables."
read_when:
  - "Changing config loading, defaults, paths, or credential discovery."
  - "Updating setup instructions."
---
# Configuration

Configuration controls local paths and provider defaults. Provider secrets stay in environment variables, not in `config.toml`.

Default paths:

- config: `~/.config/gobankcli/config.toml`
- database: `~/.local/share/gobankcli/gobankcli.db`
- exports: `~/Finance/gobankcli/exports`

Example:

```toml
default_provider = "enablebanking"
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

Enable Banking credentials use environment variables:

```bash
GOBANKCLI_ENABLEBANKING_APP_ID
GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH
GOBANKCLI_ENABLEBANKING_API # optional, defaults to https://api.enablebanking.com
```

Common local env file:

```bash
export GOBANKCLI_ENABLEBANKING_APP_ID="<enablebanking-application-id>"
export GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH="$HOME/.config/gobankcli/enablebanking.pem"
```

Secrets are never written into config by `gobankcli init`.

Run `gobankcli doctor` after editing config. It reports only provider credential presence, never secret values.

Connection entries are optional operator hints for docs and institution
archiving. Example Enable Banking entry:

```toml
[[connections]]
name = "Belfius personal via Enable Banking"
provider = "enablebanking"
institution_id = "BE:Belfius"
country = "BE"
```

Provider-specific setup is covered in [Provider Setup](provider-setup.md).
