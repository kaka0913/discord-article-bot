# Secret Manager モジュール
# Gemini API KeyとDiscord Webhook URLのシークレットを管理

# Gemini API Key用のシークレット
resource "google_secret_manager_secret" "gemini_api_key" {
  project   = var.project_id
  secret_id = "gemini-api-key"

  replication {
    auto {}
  }

  labels = {
    app         = "rss-article-curator"
    environment = "prod"
    managed_by  = "terraform"
  }
}

# Discord Webhook URL用のシークレット
resource "google_secret_manager_secret" "discord_webhook_url" {
  project   = var.project_id
  secret_id = "discord-webhook-url"

  replication {
    auto {}
  }

  labels = {
    app         = "rss-article-curator"
    environment = "prod"
    managed_by  = "terraform"
  }
}

# Cloud Functionsサービスアカウントにシークレットへのアクセス権限を付与
resource "google_secret_manager_secret_iam_member" "gemini_api_key_accessor" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.gemini_api_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.cloud_function_service_account}"
}

resource "google_secret_manager_secret_iam_member" "discord_webhook_url_accessor" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.discord_webhook_url.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.cloud_function_service_account}"
}
