output "service_account_email" {
  description = "Cloud Functionsが使用するサービスアカウントのメールアドレス - Secretsモジュールのアクセス権限設定に使用"
  value       = google_service_account.curator_function.email
}
