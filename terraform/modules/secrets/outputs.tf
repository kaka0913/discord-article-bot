output "gemini_api_key_secret_id" {
  description = "Gemini API KeyのシークレットID"
  value       = google_secret_manager_secret.gemini_api_key.secret_id
}

output "discord_webhook_url_secret_id" {
  description = "Discord Webhook URLのシークレットID"
  value       = google_secret_manager_secret.discord_webhook_url.secret_id
}

output "gemini_api_key_secret_name" {
  description = "Gemini API Keyのシークレット完全名"
  value       = google_secret_manager_secret.gemini_api_key.id
}

output "discord_webhook_url_secret_name" {
  description = "Discord Webhook URLのシークレット完全名"
  value       = google_secret_manager_secret.discord_webhook_url.id
}
