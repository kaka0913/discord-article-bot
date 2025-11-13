# Terraformインフラ構成

このディレクトリには、RSS記事キュレーションBotのGCPインフラをTerraformで管理する設定が含まれています。

## ディレクトリ構造

```
terraform/
├── modules/              # 再利用可能なTerraformモジュール
│   ├── firestore/       # Firestoreデータベースとインデックス
│   ├── secrets/         # Secret Manager（API Key、Webhook URL）
│   ├── scheduler/       # Cloud Scheduler + Pub/Sub
│   └── cloud-function/  # Cloud Functions Gen 2
└── environments/        # 環境別の設定
    └── prod/           # 本番環境
        ├── main.tf
        ├── variables.tf
        ├── outputs.tf
        ├── backend.tf
        └── terraform.tfvars.example
```

## デプロイ手順

### 1. 前提条件

- [Terraform](https://www.terraform.io/downloads) (>= 1.0) がインストールされていること
- [gcloud CLI](https://cloud.google.com/sdk/docs/install) がインストールされ、認証済みであること
- GCPプロジェクトが作成されていること

### 2. GCP認証

```bash
gcloud auth application-default login
gcloud config set project YOUR_PROJECT_ID
```

### 3. 変数ファイルの作成

```bash
cd environments/prod
cp terraform.tfvars.example terraform.tfvars
# terraform.tfvarsを編集して実際の値を設定
```

### 4. Terraform初期化

```bash
terraform init
```

### 5. プランの確認

```bash
terraform plan
```

### 6. インフラのデプロイ

```bash
terraform apply
```

### 7. シークレットの設定

Terraformでシークレットリソースを作成した後、実際の値を設定します：

```bash
# Gemini API Keyの設定
echo -n "your-actual-gemini-api-key" | gcloud secrets versions add gemini-api-key --data-file=-

# Discord Webhook URLの設定
echo -n "https://discord.com/api/webhooks/..." | gcloud secrets versions add discord-webhook-url --data-file=-
```

## 管理対象リソース

- **Firestore**: データベースとインデックス
- **Secret Manager**: Gemini API Key、Discord Webhook URL
- **Cloud Scheduler**: 毎日9:00 JSTの定期実行ジョブ
- **Pub/Sub**: Cloud Schedulerからのトリガートピック
- **Cloud Functions Gen 2**: RSS記事キュレーションの実行環境
- **Cloud Storage**: ソースコードアーカイブ保存用バケット
- **Service Account**: Cloud Functions用サービスアカウント
- **IAM**: 必要な権限設定

## 注意事項

- `terraform.tfvars`は機密情報を含むため、Gitにコミットしないでください（.gitignoreで除外済み）
- 初回デプロイ時は、Cloud Functionsのソースコードが未アップロードのためエラーになる場合があります
- State管理用のGCSバケットは別途手動で作成する必要があります（backend.tf参照）

## 各モジュールの詳細

各モジュールの詳細は、それぞれのREADME.mdを参照してください：

- [Firestoreモジュール](./modules/firestore/README.md)
- [Secret Managerモジュール](./modules/secrets/README.md)
- [Cloud Schedulerモジュール](./modules/scheduler/README.md)
- [Cloud Functionsモジュール](./modules/cloud-function/README.md)
