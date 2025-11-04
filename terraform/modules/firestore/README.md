# Firestore モジュール

## 概要

このモジュールは、RSS記事キュレーションBotの重複排除追跡のためのFirestoreデータベースを作成します。

## リソース

- `google_firestore_database`: Firestore Nativeモードデータベース
- `google_firestore_index`: notified_articlesコレクション用インデックス
- `google_firestore_index`: rejected_articlesコレクション用インデックス

## 使用するコレクション

### notified_articles
Discordに投稿された記事を追跡（重複通知を防止）

### rejected_articles
関連性が低いと評価された記事を追跡（再評価を回避）

## 入力変数

| 名前 | 説明 | 型 | 必須 |
|------|------|-----|------|
| project_id | GCPプロジェクトID | string | はい |
| region | Firestoreリージョン | string | はい |

## 出力

| 名前 | 説明 |
|------|------|
| database_name | Firestoreデータベース名 |
| database_id | FirestoreデータベースID |

## 使用例

```hcl
module "firestore" {
  source     = "../../modules/firestore"
  project_id = var.project_id
  region     = var.region
}
```

## 参照

- [Firestoreスキーマ契約](../../../specs/001-rss-article-curator/contracts/firestore-schema.md)
