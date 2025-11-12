variable "project_id" {
  description = "GCPプロジェクトID"
  type        = string
}

variable "region" {
  description = "Cloud Functionsのデプロイリージョン"
  type        = string
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
