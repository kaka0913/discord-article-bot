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

# Cloud Functionモジュール（gcloudでデプロイしたため、コメントアウト）
# module "cloud_function" {
#   source                = "../../modules/cloud-function"
#   project_id            = var.project_id
#   region                = var.region
#   config_url            = var.config_url
#   source_archive_object = var.source_archive_object
#   depends_on = [
#     google_project_service.cloudfunctions,
#     google_project_service.cloudbuild
#   ]
# }

# Cloud Schedulerモジュール（Cloud Function作成後に設定）
# gcloudでデプロイしたCloud FunctionのURLとサービスアカウントを直接指定
module "scheduler" {
  source = "../../modules/scheduler"

  project_id                        = var.project_id
  region                            = var.region
  function_url                      = "https://asia-northeast1-rss-article-curator-prod.cloudfunctions.net/rss-article-curator"
  scheduler_service_account_email   = "rss-curator-function@rss-article-curator-prod.iam.gserviceaccount.com"

  depends_on = [google_project_service.scheduler]
}

# Secret Managerモジュール
module "secrets" {
  source = "../../modules/secrets"

  project_id                     = var.project_id
  cloud_function_service_account = "rss-curator-function@rss-article-curator-prod.iam.gserviceaccount.com"

  depends_on = [google_project_service.secretmanager]
}
