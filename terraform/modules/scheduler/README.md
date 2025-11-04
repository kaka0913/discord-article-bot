# Cloud Scheduler モジュール

## 概要

このモジュールは、RSS記事キュレーションBotを毎日8:00 JST（日本時間）に自動実行するためのCloud SchedulerジョブとPub/Subトピックを作成します。

## リソース

- `google_pubsub_topic`: Cloud Schedulerからのトリガーメッセージ受信用トピック
- `google_cloud_scheduler_job`: 毎日8:00 JSTに実行されるスケジューラジョブ

## スケジュール

- **頻度**: 毎日
- **時刻**: 8:00 JST（Asia/Tokyo）
- **Cron式**: `0 8 * * *`

## リトライ設定

- **リトライ回数**: 3回
- **最小バックオフ**: 5秒
- **最大バックオフ**: 60秒

## 入力変数

| 名前 | 説明 | 型 | 必須 |
|------|------|-----|------|
| project_id | GCPプロジェクトID | string | はい |
| region | Cloud Schedulerリージョン | string | はい |

## 出力

| 名前 | 説明 |
|------|------|
| pubsub_topic_name | Pub/Subトピック名 |
| pubsub_topic_id | Pub/SubトピックID（完全修飾名） |
| scheduler_job_name | Cloud Schedulerジョブ名 |

## 使用例

```hcl
module "scheduler" {
  source     = "../../modules/scheduler"
  project_id = var.project_id
  region     = var.region
}
```

## 手動トリガー

デプロイ後、以下のコマンドで手動実行できます：

```bash
gcloud scheduler jobs run daily-rss-curator --project=YOUR_PROJECT_ID
```
