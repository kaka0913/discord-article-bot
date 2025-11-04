variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "cloud_function_service_account" {
  description = "Cloud Functionsが使用するサービスアカウントのメールアドレス"
  type        = string
}
