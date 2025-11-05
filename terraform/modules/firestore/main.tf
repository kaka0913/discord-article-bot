# Firestore Database モジュール
# 記事の重複排除追跡のためのFirestoreデータベースとインデックスを管理

resource "google_firestore_database" "database" {
  project     = var.project_id
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  # 削除保護を有効化（本番環境での誤削除を防止）
  # ABANDON: terraform destroyでもFirestoreは削除されず、GCPコンソールから手動削除が必要
  deletion_policy = "ABANDON"
}

# notified_articles コレクション用のインデックス
# notified_at フィールドでのソート・クエリ用
resource "google_firestore_index" "notified_articles_timestamp" {
  project    = var.project_id
  database   = google_firestore_database.database.name
  collection = "notified_articles"

  fields {
    field_path = "notified_at"
    order      = "DESCENDING"
  }

  fields {
    field_path = "__name__"
    order      = "DESCENDING"
  }
}

# rejected_articles コレクション用のインデックス
# evaluated_at フィールドでのソート・クエリ用（TTL削除クエリ用）
resource "google_firestore_index" "rejected_articles_timestamp" {
  project    = var.project_id
  database   = google_firestore_database.database.name
  collection = "rejected_articles"

  fields {
    field_path = "evaluated_at"
    order      = "DESCENDING"
  }

  fields {
    field_path = "__name__"
    order      = "DESCENDING"
  }
}
