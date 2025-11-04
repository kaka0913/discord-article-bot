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

variable "pubsub_topic_id" {
  description = "トリガー用のPub/SubトピックID（完全修飾名）"
  type        = string
}

variable "source_archive_object" {
  description = "Cloud Storageに配置されたソースコードアーカイブのオブジェクト名"
  type        = string
  default     = "source.zip"
}

variable "gemini_api_key_secret" {
  description = "Gemini API KeyのSecret Manager ID"
  type        = string
}

variable "discord_webhook_secret" {
  description = "Discord Webhook URLのSecret Manager ID"
  type        = string
}
