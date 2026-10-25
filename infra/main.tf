resource "scaleway_function_namespace" "bmur2" {
  name        = var.namespace_name
  description = "bmur2 Telegram bot namespace"
}

resource "scaleway_function" "telegram_webhook" {
  namespace_id = scaleway_function_namespace.bmur2.id
  name         = var.function_name
  description  = "Go Telegram webhook handler for bmur2"
  runtime      = var.runtime
  handler      = var.handler
  privacy      = "public"

  memory_limit = var.memory_limit
  timeout      = var.timeout
  min_scale    = 0
  max_scale    = 2

  zip_file = var.zip_file
  zip_hash = filesha256(var.zip_file)
  deploy   = true

  secret_environment_variables = {
    BMUR_TOKEN              = var.bmur_token
    BMUR_LASTFM_APIKEY      = var.bmur_lastfm_apikey
    BMUR_LASTFM_USERNAME    = var.bmur_lastfm_username
    TELEGRAM_WEBHOOK_SECRET = var.telegram_webhook_secret
  }
}
