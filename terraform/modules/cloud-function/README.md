# Cloud Functions Gen 2 モジュール

## 概要

このモジュールは、RSS記事キュレーションBotのメイン実行環境であるCloud Functions Gen 2リソースを作成します。

## リソース

- `google_service_account`: Cloud Functions用サービスアカウント
- `google_project_iam_member`: Firestoreアクセス権限
- `google_storage_bucket`: ソースコードアーカイブ保存用バケット
- `google_cloudfunctions2_function`: Cloud Functions Gen 2インスタンス

## 実行環境

- **ランタイム**: Go 1.22
- **メモリ**: 512Mi
- **タイムアウト**: 300秒（5分）
- **最大インスタンス数**: 1（同時実行を1つに制限）
- **最小インスタンス数**: 0（コールドスタート許容）

## トリガー

Pub/Subトピック（`rss-curator-trigger`）からのメッセージによってトリガーされます。

## 環境変数

| 名前 | 説明 |
|------|------|
| CONFIG_URL | config.jsonファイルのURL |
| GCP_PROJECT_ID | GCPプロジェクトID |
| GEMINI_API_KEY_SECRET | Gemini API KeyのSecret Manager ID |
| DISCORD_WEBHOOK_SECRET | Discord Webhook URLのSecret Manager ID |

## 入力変数

| 名前 | 説明 | 型 | 必須 | デフォルト |
|------|------|-----|------|---------|
| project_id | GCPプロジェクトID | string | はい | - |
| region | デプロイリージョン | string | はい | - |
| config_url | config.json URL | string | はい | - |
| pubsub_topic_id | Pub/SubトピックID | string | はい | - |
| source_archive_object | ソースアーカイブ名 | string | いいえ | "source.zip" |
| gemini_api_key_secret | Gemini API Key Secret ID | string | はい | - |
| discord_webhook_secret | Discord Webhook Secret ID | string | はい | - |

## 出力

| 名前 | 説明 |
|------|------|
| function_name | Cloud Function名 |
| function_uri | Cloud FunctionのURI |
| service_account_email | サービスアカウントメール |
| storage_bucket_name | ソースコード保存バケット名 |

## 使用例

```hcl
module "cloud_function" {
  source                   = "../../modules/cloud-function"
  project_id               = var.project_id
  region                   = var.region
  config_url               = var.config_url
  pubsub_topic_id          = module.scheduler.pubsub_topic_id
  gemini_api_key_secret    = module.secrets.gemini_api_key_secret_id
  discord_webhook_secret   = module.secrets.discord_webhook_url_secret_id
}
```

## デプロイ

ソースコードをデプロイするには、事前にCloud Storageバケットにzipファイルをアップロードします：

```bash
# ソースコードをzipに圧縮
zip -r source.zip . -x "*.git*" -x "terraform/*"

# Cloud Storageにアップロード
gsutil cp source.zip gs://${PROJECT_ID}-curator-function-source/source.zip
```
