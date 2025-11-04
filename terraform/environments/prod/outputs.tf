output "firestore_database_name" {
  description = "Firestoreデータベース名"
  value       = module.firestore.database_name
}

output "cloud_function_name" {
  description = "Cloud Function名"
  value       = module.cloud_function.function_name
}

output "cloud_function_uri" {
  description = "Cloud FunctionのURI"
  value       = module.cloud_function.function_uri
}

output "service_account_email" {
  description = "Cloud Functionsサービスアカウントのメールアドレス"
  value       = module.cloud_function.service_account_email
}

output "pubsub_topic_name" {
  description = "Pub/Subトピック名"
  value       = module.scheduler.pubsub_topic_name
}

output "scheduler_job_name" {
  description = "Cloud Schedulerジョブ名"
  value       = module.scheduler.scheduler_job_name
}

output "storage_bucket_name" {
  description = "ソースコードアーカイブ保存用バケット名"
  value       = module.cloud_function.storage_bucket_name
}

output "gemini_api_key_secret_id" {
  description = "Gemini API KeyのシークレットID"
  value       = module.secrets.gemini_api_key_secret_id
}

output "discord_webhook_url_secret_id" {
  description = "Discord Webhook URLのシークレットID"
  value       = module.secrets.discord_webhook_url_secret_id
}
