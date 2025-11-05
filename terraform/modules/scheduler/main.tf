# Cloud Scheduler モジュール
# 毎日8:00 JSTにCloud FunctionsをトリガーするスケジューラとPub/Subトピックを管理

# Pub/Subトピック（Cloud Schedulerからのメッセージ受信用）
resource "google_pubsub_topic" "curator_trigger" {
  project = var.project_id
  name    = "rss-curator-trigger"

  labels = {
    app         = "rss-article-curator"
    environment = "prod"
    managed_by  = "terraform"
  }
}

# Cloud Scheduler ジョブ（毎日8:00 JSTに実行）
resource "google_cloud_scheduler_job" "daily_curator" {
  project          = var.project_id
  name             = "daily-rss-curator"
  description      = "RSS記事キュレーションを毎日8:00 JSTに実行"
  schedule         = "0 8 * * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "320s"

  pubsub_target {
    topic_name = google_pubsub_topic.curator_trigger.id
    data       = base64encode("{\"trigger\":\"scheduled\"}")
  }

  retry_config {
    retry_count          = 3
    min_backoff_duration = "5s"
    max_backoff_duration = "60s"
  }
}
