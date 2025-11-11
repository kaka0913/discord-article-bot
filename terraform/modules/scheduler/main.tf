# Cloud Scheduler モジュール
# 毎日8:00 JSTにCloud FunctionsをトリガーするスケジューラとPub/Subトピックを管理

# Cloud Scheduler ジョブ（毎日9:00 JSTに実行）
# HTTPトリガーに変更（イベントトリガーは540秒制限があるため）
resource "google_cloud_scheduler_job" "daily_curator" {
  project          = var.project_id
  name             = "daily-rss-curator"
  description      = "RSS記事キュレーションを毎日9:00 JSTに実行"
  schedule         = "0 9 * * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "3600s"

  http_target {
    uri         = var.function_url
    http_method = "POST"

    oidc_token {
      service_account_email = var.scheduler_service_account_email
    }
  }

  retry_config {
    retry_count          = 3
    min_backoff_duration = "5s"
    max_backoff_duration = "60s"
  }
}
