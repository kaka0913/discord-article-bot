output "service_account_email" {
  description = "Cloud Functionsサービスアカウントのメールアドレス"
  value       = module.cloud_function.service_account_email
}
