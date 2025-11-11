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

# 単一フィールドインデックスは自動的に作成されるため、
# 明示的なインデックス定義は不要です
# 複合インデックスが必要になった場合のみ、ここに追加してください
