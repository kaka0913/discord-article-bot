# Cloud Functions Gen 2 モジュール
# RSS記事キュレーションBotのメイン実行環境

# サービスアカウント（Cloud Functionsが使用）
resource "google_service_account" "curator_function" {
  project      = var.project_id
  account_id   = "rss-curator-function"
  display_name = "RSS Curator Cloud Function Service Account"
  description  = "Cloud FunctionsからFirestore、Secret Manager、Gemini APIにアクセスするためのサービスアカウント"
}

# Firestoreへのアクセス権限
resource "google_project_iam_member" "firestore_user" {
  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.curator_function.email}"
}

# Cloud Functions用のストレージバケット（ソースコードアーカイブ保存用）
resource "google_storage_bucket" "function_source" {
  project       = var.project_id
  name          = "${var.project_id}-curator-function-source"
  location      = var.region
  force_destroy = true

  uniform_bucket_level_access = true

  labels = {
    app = "rss-article-curator"
  }
}

# Cloud Functions Gen 2（Pub/Subトリガー）
resource "google_cloudfunctions2_function" "curator" {
  project     = var.project_id
  name        = "rss-article-curator"
  location    = var.region
  description = "RSS記事を取得し、Geminiで評価し、Discordに通知する"

  build_config {
    runtime     = "go122"
    entry_point = "CuratorHandler"
    source {
      storage_source {
        bucket = google_storage_bucket.function_source.name
        object = var.source_archive_object
      }
    }
  }

  service_config {
    max_instance_count    = 1
    min_instance_count    = 0
    available_memory      = "512Mi"
    timeout_seconds       = 300
    service_account_email = google_service_account.curator_function.email

    environment_variables = {
      CONFIG_URL           = var.config_url
      GCP_PROJECT_ID       = var.project_id
      GEMINI_API_KEY_SECRET = var.gemini_api_key_secret
      DISCORD_WEBHOOK_SECRET = var.discord_webhook_secret
    }
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = var.pubsub_topic_id
    retry_policy   = "RETRY_POLICY_RETRY"
  }

  labels = {
    app = "rss-article-curator"
  }
}
