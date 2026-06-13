---
summary: "Enable Banking and GoCardless setup paths, credentials, consent flow, and limits."
read_when:
  - "Updating provider setup instructions."
  - "Checking Enable Banking or GoCardless public docs."
title: "Provider Setup"
description: "Configure Enable Banking or GoCardless for read-only bank archive sync with gobankcli."
---
# Provider Setup

Providers supply read-only account information through official API flows. `gobankcli` normalizes provider payloads before writing the local SQLite archive.

## Enable Banking

Enable Banking AIS is the recommended live setup for Belfius.

### Dashboard Application

Create an application in the Enable Banking control panel:

```text
Environment: Production
Private key generation: Generate in browser and export private key
Application name: gobankcli
Allowed Redirect URL: https://127.0.0.1:28787/enablebanking/callback
Description: Personal read-only transaction archive
Data protection email: <your email>
Privacy Policy URL: https://github.com/BramVR/goBankCli
Terms of Service URL: https://github.com/BramVR/goBankCli
```

If Enable Banking rejects the IP-literal URL, use `https://localhost:28787/enablebanking/callback` and run the listener with `--listen localhost:28787 --listen-https`.

### Restricted Production Account Linking

Restricted production only returns accounts linked in the Enable Banking dashboard. Open the application and use `Link accounts` before running the CLI authorization flow. If an account is not linked there, `sync` cannot discover it later.

### Credentials

Install the downloaded PEM key locally:

```bash
ENABLEBANKING_APPLICATION_ID="replace-with-your-application-id"
mkdir -p ~/.config/gobankcli
install -m 600 "$HOME/Downloads/${ENABLEBANKING_APPLICATION_ID}.pem" ~/.config/gobankcli/enablebanking.pem
```

Store the local environment variables outside `config.toml`:

```bash
cat > ~/.config/gobankcli/enablebanking.env <<EOF
export GOBANKCLI_ENABLEBANKING_APP_ID="$ENABLEBANKING_APPLICATION_ID"
export GOBANKCLI_ENABLEBANKING_PRIVATE_KEY_PATH="$HOME/.config/gobankcli/enablebanking.pem"
EOF
chmod 600 ~/.config/gobankcli/enablebanking.env
source ~/.config/gobankcli/enablebanking.env
```

Optional override:

```bash
export GOBANKCLI_ENABLEBANKING_API="https://api.enablebanking.com"
```

### HTTPS Loopback Callback

Enable Banking production redirect URLs must be HTTPS. The recommended CLI flow uses a short-lived HTTPS listener on your own machine:

```bash
mkdir -p ~/.config/gobankcli/tls
chmod 700 ~/.config/gobankcli/tls
openssl req -x509 -nodes -newkey rsa:2048 -days 825 \
  -keyout ~/.config/gobankcli/tls/localhost.key \
  -out ~/.config/gobankcli/tls/localhost.crt \
  -subj '/CN=127.0.0.1' \
  -addext 'subjectAltName = IP:127.0.0.1,DNS:localhost'
chmod 600 ~/.config/gobankcli/tls/localhost.key ~/.config/gobankcli/tls/localhost.crt
```

On macOS, trust the local certificate:

```bash
security add-trusted-cert -d -r trustRoot -p ssl \
  -k ~/Library/Keychains/login.keychain-db \
  ~/.config/gobankcli/tls/localhost.crt
```

On other systems, use your OS trust-store tooling or a tool such as `mkcert`. The certificate and key stay local and are used only by the loopback callback server.

### Authorize

```bash
gobankcli doctor
gobankcli institutions --provider enablebanking --country BE --query belfius
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --listen 127.0.0.1:28787 \
  --listen-https \
  --listen-cert ~/.config/gobankcli/tls/localhost.crt \
  --listen-key ~/.config/gobankcli/tls/localhost.key \
  --callback-timeout 10m
```

`connect --listen` prints an Enable Banking browser URL on stderr, waits for one callback, validates `state`, exchanges `code`, archives the session/accounts, and prints the authorized session report.

The `provider_connection_id` in that report is the Enable Banking session ID. Use it for later `sync`, `accounts`, and automation commands:

```bash
gobankcli sync --provider enablebanking --connection SESSION_ID --from 2026-01-01
```

### Manual Callback Fallback

If the local listener cannot be used, keep the same registered local HTTPS callback and exchange the callback manually:

```bash
gobankcli connect \
  --provider enablebanking \
  --institution BE:Belfius \
  --redirect https://127.0.0.1:28787/enablebanking/callback

gobankcli authorize \
  --provider enablebanking \
  --url "$ENABLEBANKING_CALLBACK_URL" \
  --institution BE:Belfius
```

When no listener is running, the browser may show a local connection error after bank consent. Copy the full `https://127.0.0.1:28787/...` callback URL from the address bar into `ENABLEBANKING_CALLBACK_URL`. Do not use a third-party redirect URL because the callback contains authorization parameters.

### Enable Banking Limits

- Restricted production returns only dashboard-linked accounts.
- Production callback URLs must be HTTPS.
- Missing credentials fail live provider commands clearly.
- If a later session response only returns account UIDs, `accounts` and `sync` reuse the stable account metadata archived during authorization.

## GoCardless

GoCardless Bank Account Data is also supported for read-only institution, consent, account, and booked transaction flows.

### Credentials

Set credentials as environment variables:

```bash
export GOBANKCLI_GOCARDLESS_SECRET_ID="<secret-id>"
export GOBANKCLI_GOCARDLESS_SECRET_KEY="<secret-key>"
```

Do not put these values in `config.toml`, tests, docs, examples, logs, or commits.

### Consent Flow

Find an institution:

```bash
gobankcli institutions --provider gocardless --country BE --query belfius
```

Create a requisition. Use your own browser landing URL for `GOCARDLESS_REDIRECT_URL`:

```bash
gobankcli connect \
  --provider gocardless \
  --institution BELFIUS_GKCCBEBB \
  --redirect "$GOCARDLESS_REDIRECT_URL"
```

Open the returned `redirect_url`, complete bank consent, then use the printed `provider_connection_id` for account and sync commands:

```bash
gobankcli accounts --provider gocardless --connection REQUISITION_ID
gobankcli sync --provider gocardless --connection REQUISITION_ID --from 2026-01-01
```

### GoCardless Limits

- Pending transactions are not archived yet.
- Consent can expire or be revoked; rerun `connect` when renewal is needed.
- Without credentials, live GoCardless commands fail clearly and never fake a successful sync.

## Non-Goals

Both provider paths keep the same project boundaries: no scraping, no payment initiation, no bank password storage, no cloud upload, and no real bank data in examples.
