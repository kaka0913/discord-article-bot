# 本番環境のTerraform設定
# RSS記事キュレーションBotの全GCPリソースを定義

terraform {
  required_version = ">= 1.0, < 2.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# 必要なAPIを有効化
resource "google_project_service" "firestore" {
  project = var.project_id
  service = "firestore.googleapis.com"

  disable_on_destroy = false
}

resource "google_project_service" "secretmanager" {
  project = var.project_id
  service = "secretmanager.googleapis.com"

  disable_on_destroy = false
}

resource "google_project_service" "scheduler" {
  project = var.project_id
  service = "cloudscheduler.googleapis.com"

  disable_on_destroy = false
}

resource "google_project_service" "cloudfunctions" {
  project = var.project_id
  service = "cloudfunctions.googleapis.com"

  disable_on_destroy = false
}

resource "google_project_service" "cloudbuild" {
  project = var.project_id
  service = "cloudbuild.googleapis.com"

  disable_on_destroy = false
}

resource "google_project_service" "pubsub" {
  project = var.project_id
  service = "pubsub.googleapis.com"

  disable_on_destroy = false
}

# Firestoreモジュール
module "firestore" {
  source = "../../modules/firestore"

  project_id = var.project_id
  region     = var.region

  depends_on = [google_project_service.firestore]
}

# Cloud Functionモジュール（先にデプロイしてservice_account_emailを取得）
module "cloud_function" {
  source = "../../modules/cloud-function"

  project_id             = var.project_id
  region                 = var.region
  config_url             = var.config_url
  pubsub_topic_id        = module.scheduler.pubsub_topic_id
  source_archive_object  = var.source_archive_object
  gemini_api_key_secret  = "gemini-api-key"
  discord_webhook_secret = "discord-webhook-url"

  depends_on = [
    google_project_service.cloudfunctions,
    google_project_service.cloudbuild,
    google_project_service.pubsub
  ]
}

# Secret Managerモジュール
module "secrets" {
  source = "../../modules/secrets"

  project_id                     = var.project_id
  cloud_function_service_account = module.cloud_function.service_account_email

  depends_on = [google_project_service.secretmanager]
}

# Cloud Schedulerモジュール
module "scheduler" {
  source = "../../modules/scheduler"

  project_id = var.project_id
  region     = var.region

  depends_on = [google_project_service.scheduler, google_project_service.pubsub]
}
