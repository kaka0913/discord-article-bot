output "function_name" {
  description = "Cloud Function名"
  value       = google_cloudfunctions2_function.curator.name
}

output "function_uri" {
  description = "Cloud FunctionのURI"
  value       = google_cloudfunctions2_function.curator.service_config[0].uri
}

output "service_account_email" {
  description = "Cloud Functionsが使用するサービスアカウントのメールアドレス"
  value       = google_service_account.curator_function.email
}

output "storage_bucket_name" {
  description = "ソースコードアーカイブ保存用バケット名"
  value       = google_storage_bucket.function_source.name
}
