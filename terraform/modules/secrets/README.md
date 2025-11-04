# Secret Manager モジュール

## 概要

このモジュールは、RSS記事キュレーションBotで使用するシークレット（Gemini API Key、Discord Webhook URL）をGoogle Secret Managerで管理します。

## リソース

- `google_secret_manager_secret`: Gemini API Key用シークレット
- `google_secret_manager_secret`: Discord Webhook URL用シークレット
- `google_secret_manager_secret_iam_member`: シークレットアクセス権限（Cloud Functions用）

## シークレット

### gemini-api-key
Gemini APIとの通信に使用するAPIキー

### discord-webhook-url
Discord通知送信先のWebhook URL

## 入力変数

| 名前 | 説明 | 型 | 必須 |
|------|------|-----|------|
| project_id | GCPプロジェクトID | string | はい |
| cloud_function_service_account | Cloud Functionsサービスアカウント | string | はい |

## 出力

| 名前 | 説明 |
|------|------|
| gemini_api_key_secret_id | Gemini API KeyのシークレットID |
| discord_webhook_url_secret_id | Discord Webhook URLのシークレットID |
| gemini_api_key_secret_name | Gemini API Keyのシークレット完全名 |
| discord_webhook_url_secret_name | Discord Webhook URLのシークレット完全名 |

## 使用例

```hcl
module "secrets" {
  source                          = "../../modules/secrets"
  project_id                      = var.project_id
  cloud_function_service_account  = module.cloud_function.service_account_email
}
```

## シークレット値の設定方法

Terraformでシークレットリソースを作成後、gcloud CLIで実際の値を設定します：

```bash
# Gemini API Keyの設定
echo -n "your-gemini-api-key" | gcloud secrets versions add gemini-api-key --data-file=-

# Discord Webhook URLの設定
echo -n "https://discord.com/api/webhooks/..." | gcloud secrets versions add discord-webhook-url --data-file=-
```
