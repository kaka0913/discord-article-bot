# Terraform State管理
# GCS backendでtfstateを管理（チーム開発時の状態共有と競合防止）

# 注意: 初回適用時はローカルbackendで実行し、その後GCSバケットを作成してから
#       backend設定を有効化してterraform initを再実行してください

# terraform {
#   backend "gcs" {
#     bucket = "YOUR_PROJECT_ID-terraform-state"
#     prefix = "rss-curator/prod"
#   }
# }

# GCSバケット作成コマンド（初回のみ）:
# gsutil mb -p YOUR_PROJECT_ID -l asia-northeast1 gs://YOUR_PROJECT_ID-terraform-state
# gsutil versioning set on gs://YOUR_PROJECT_ID-terraform-state
