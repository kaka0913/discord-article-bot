variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "region" {
  description = "デプロイリージョン（例: asia-northeast1）"
  type        = string
  default     = "asia-northeast1"
}

variable "config_url" {
  description = "config.jsonファイルのURL（GitHub raw URLなど）"
  type        = string
}

variable "source_archive_object" {
  description = "Cloud Storageに配置されたソースコードアーカイブのオブジェクト名"
  type        = string
  default     = "source.zip"
}
