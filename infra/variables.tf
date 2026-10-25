variable "project_id" {
  description = "Scaleway Project ID"
  type        = string
}

variable "region" {
  description = "Scaleway region"
  type        = string
  default     = "fr-par"
}

variable "namespace_name" {
  description = "Serverless Functions namespace name"
  type        = string
  default     = "bmur2"
}

variable "function_name" {
  description = "Serverless Function name"
  type        = string
  default     = "telegram-webhook"
}

variable "runtime" {
  description = "Serverless Function runtime"
  type        = string
  default     = "go126"
}

variable "handler" {
  description = "Serverless Function handler (exported Go function name)"
  type        = string
  default     = "Handle"
}

variable "zip_file" {
  description = "Path to deployment zip file"
  type        = string
  default     = "../artifacts/function.zip"
}

variable "memory_limit" {
  description = "Memory limit in MB"
  type        = number
  default     = 128
}

variable "timeout" {
  description = "Function timeout in seconds"
  type        = number
  default     = 15
}

variable "bmur_token" {
  description = "Telegram bot token"
  type        = string
  sensitive   = true
}

variable "bmur_lastfm_apikey" {
  description = "Last.fm API key"
  type        = string
  sensitive   = true
}

variable "bmur_lastfm_username" {
  description = "Last.fm username"
  type        = string
  sensitive   = true
}

variable "telegram_webhook_secret" {
  description = "Telegram webhook secret header value"
  type        = string
  sensitive   = true
}
