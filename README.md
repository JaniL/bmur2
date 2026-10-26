# bmur2

Telegram bot that shows your Last.fm now playing and recent tracks, plus upcoming Christina Regina (ilotalo) reservations. Rewritten in Go, deployed as a Scaleway Serverless Function that Telegram calls over a webhook. No polling, no background loops: the function only runs when someone sends the bot a command.

## Commands

- `/np` -> current track
- `/recent` -> recent tracks
- `/reservations` -> upcoming ilotalo reservations for the next 3 months (max 15, links to the calendar)

All also work with a bot mention (`/np@botname`) and trailing arguments.

## Architecture

- Telegram -> HTTPS webhook -> Scaleway Function (`go126` runtime, public endpoint)
- Function validates `X-Telegram-Bot-Api-Secret-Token` (constant-time compare)
- Function fetches Last.fm data on demand (`user.getrecenttracks`, API key only)
- Function scrapes the ilotalo reservation calendar on demand (`ilotalo.matlu.fi`, table parsed with `golang.org/x/net/html`)
- Function shows a typing indicator (`sendChatAction`) in the chat while fetching
- Function replies through the Telegram Bot API (`sendMessage`, HTML parse mode)

Requests without the correct secret header are rejected with 401 before any bot
work happens. Unknown updates are ACKed with 200 so Telegram does not redeliver them.

## Repository layout

```
config.go           # required environment variables
handler.go          # exported Handle(w, r) - the Scaleway entrypoint
bot.go              # bot init, /np, /recent and /reservations handlers, telegram sends
lastfm.go           # Last.fm REST client
reservations.go     # ilotalo calendar scraper (fetch + HTML table parser + upcoming filter)
format.go           # HTML formatting, relative "x hours ago" dates
cmd/devserver/      # local webhook server for manual testing
scripts/package.sh  # builds artifacts/function.zip for deployment
infra/              # Terraform for the Scaleway function
```

## Environment variables

| Variable | Purpose |
|---|---|
| `BMUR_TOKEN` | Telegram bot token from @BotFather |
| `TELEGRAM_WEBHOOK_SECRET` | Shared secret Telegram sends in the `X-Telegram-Bot-Api-Secret-Token` header (1-256 chars, `A-Za-z0-9_-`) |
| `BMUR_LASTFM_APIKEY` | Last.fm API key |
| `BMUR_LASTFM_USERNAME` | Last.fm username to query |

Note: the old `BMUR_LASTFM_SECRET` is no longer needed; `user.getrecenttracks` requires only the API key.

## Local development

```bash
go test ./...
go vet ./...
```

Run the webhook server locally (set the four variables first):

```bash
export BMUR_TOKEN=... TELEGRAM_WEBHOOK_SECRET=... BMUR_LASTFM_APIKEY=... BMUR_LASTFM_USERNAME=...
go run ./cmd/devserver
curl -X POST localhost:8080/telegram/webhook \
  -H 'X-Telegram-Bot-Api-Secret-Token: ...' \
  -d '{"update_id":1,"message":{"message_id":1,"chat":{"id":123,"type":"private"},"date":0,"text":"/np"}}'
```

## Deployment

### 1) Build the artifact

From the repository root:

```bash
scripts/package.sh
```

The script runs tests, `go mod tidy`, `go mod vendor`, and zips the sources plus
`vendor/` into `artifacts/function.zip`. If Go is not installed it zips sources only;
Scaleway then fetches the dependencies itself at build time.

### 2) Configure credentials

Scaleway credentials (Terraform provider picks these up from `scw` CLI config or env):

```bash
scw init
# or: export SCW_ACCESS_KEY=... SCW_SECRET_KEY=...
```

Secrets for the function (never commit these; use env vars):

```bash
export TF_VAR_project_id="<scaleway-project-id>"
export TF_VAR_bmur_token="<telegram-bot-token>"
export TF_VAR_telegram_webhook_secret="<webhook-shared-secret>"
export TF_VAR_bmur_lastfm_apikey="<lastfm-api-key>"
export TF_VAR_bmur_lastfm_username="<lastfm-username>"
```

### 3) Apply

All non-secret values have sensible defaults, and secrets come from the exported
`TF_VAR_*` variables, so no `-var-file` is needed (a `-var-file` would take
precedence over the environment variables):

```bash
cd infra
terraform init
terraform plan
terraform apply
terraform output -raw webhook_url
```

Optional non-secret overrides (`runtime`, `handler`, `memory_limit`, ...) can be set
by copying `terraform.tfvars.example` to `terraform.tfvars` (gitignored) and editing it.

The machine that runs `terraform apply` owns the local `terraform.tfstate` file
(it is gitignored). Keep deploying from that machine, or move the state to a remote
backend before switching machines.

### 4) Register the Telegram webhook

Run from the `infra/` directory:

```bash
curl -X POST "https://api.telegram.org/bot${TF_VAR_bmur_token}/setWebhook" \
  -d "url=$(terraform output -raw webhook_url)" \
  -d "secret_token=${TF_VAR_telegram_webhook_secret}" \
  -d 'allowed_updates=["message"]' \
  -d "drop_pending_updates=true"
```

`drop_pending_updates=true` discards the backlog queued while the bot was offline
(safe to omit afterwards). Run this step again whenever the webhook secret changes.

### 5) Verify

```bash
curl "https://api.telegram.org/bot${TF_VAR_bmur_token}/getWebhookInfo"
```

- Send `/np`, `/recent` and `/reservations` to the bot in Telegram.
- Check function logs in the Scaleway console (Cockpit) if something fails.
- Confirm the endpoint rejects requests without the secret header:

```bash
curl -i -X POST "$(terraform output -raw webhook_url)" -d '{}'
# HTTP/1.1 401 Unauthorized
```

## Notes

- Updating the runtime later only requires changing `runtime` (e.g. `go126` -> newer) and redeploying.
- The previous TypeScript implementation is preserved on the `archive/ts-serverless-refactor` branch.
- If tokens rotate, re-apply Terraform with updated `TF_VAR_*` values and re-register the webhook when the webhook secret changes.
- The ilotalo server serves an expired certificate on some TLS code paths; the reservations client pins
  `MaxVersion: tls.VersionTLS12` with Go's default cipher suites, which currently negotiates the valid
  Let's Encrypt certificate. If `/reservations` starts failing with certificate errors, the ilotalo TLS
  setup has changed and that pin should be revisited.
- The old bmur-reservation-api is dead; the bot scrapes ilotalo directly. The calendar markup is tag-soup
  (rows without `</tr>`, unquoted attributes, raw `&` in hrefs), which the tokenizer-based parser tolerates.
