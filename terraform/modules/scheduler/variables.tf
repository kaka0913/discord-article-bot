variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "region" {
  description = "Cloud Schedulerジョブのリージョン"
  type        = string
}

variable "function_url" {
  description = "Cloud FunctionのHTTPエンドポイントURL"
  type        = string
}

variable "scheduler_service_account_email" {
  description = "Cloud Schedulerが使用するサービスアカウント"
  type        = string
}
