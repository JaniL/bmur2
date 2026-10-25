output "function_domain_name" {
  description = "Public function domain name"
  value       = scaleway_function.telegram_webhook.domain_name
}

output "webhook_url" {
  description = "Telegram webhook URL"
  value       = "https://${scaleway_function.telegram_webhook.domain_name}"
}
