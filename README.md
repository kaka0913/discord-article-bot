# RSS記事キュレーションBot

技術ブログから毎日記事を収集し、Gemini LLMで関連性を評価してDiscordに通知するサーバーレスBot。

## 主要機能

- 毎日自動実行（JST 9:00）
- Gemini APIによる記事の関連性評価
- Discord Webhookで通知
- Firestoreで重複排除
- **config.json**で記事の好みをカスタマイズ可能

## 技術スタック

- **Go 1.21+** / Google Cloud Functions Gen 2
- **Terraform** / Firestore / Secret Manager
- **Gemini Flash API** / Discord Webhook
- **GitHub Actions** (CI/CD)

## 記事の好みをカスタマイズ

### 設定ファイル（config.json）

```json
{
  "interests": [
    {
      "topic": "Go言語",
      "aliases": ["Golang", "Go"],
      "priority": "high"
    }
  ],
  "notification_settings": {
    "max_articles": 5,
    "min_relevance_score": 70
  },
  "rss_sources": [
    {
      "name": "Zenn.dev",
      "url": "https://zenn.dev/feed",
      "enabled": true
    }
  ]
}
```

### 変更手順

1. ブランチ作成: `git checkout -b config/update-interests`
2. config.jsonを編集
3. コミット: `git commit -m "config: トピック更新"`
4. プッシュしてPR作成
5. 自動テストが実行される
6. マージすると自動デプロイ
7. **翌朝9:00から新設定で通知開始**

## CI/CD

### プルリクエスト
- ユニットテスト自動実行
- config.json構文検証
- テスト結果をPRにコメント

### mainマージ時
- テスト → ビルド → 自動デプロイ
- Cloud Functionsに反映

## 自分用にデプロイする

このリポジトリをフォークして、自分のGCPプロジェクトにデプロイできます。

### 1. フォークとクローン

```bash
# GitHubでリポジトリをフォーク後
git clone https://github.com/YOUR_USERNAME/discord-article-bot.git
cd discord-article-bot
```

### 2. GCPプロジェクトの作成

```bash
# プロジェクト作成（プロジェクトIDは自分で決める）
gcloud projects create YOUR-PROJECT-ID
gcloud config set project YOUR-PROJECT-ID

# 必要なAPIを有効化（すべて必須）
gcloud services enable \
  cloudresourcemanager.googleapis.com \
  cloudfunctions.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  run.googleapis.com \
  eventarc.googleapis.com \
  pubsub.googleapis.com \
  firestore.googleapis.com \
  secretmanager.googleapis.com \
  cloudscheduler.googleapis.com
```

### 3. Terraformでインフラ構築

```bash
cd terraform/environments/prod

# 変数ファイルを作成
cp terraform.tfvars.example terraform.tfvars

# terraform.tfvarsを編集してプロジェクトIDを設定
# 例: project_id = "YOUR-PROJECT-ID"

# デプロイ
terraform init
terraform apply
```

### 4. シークレットの設定

```bash
# Gemini API Key（https://aistudio.google.com/app/apikey で取得）
echo -n "YOUR_GEMINI_API_KEY" | \
  gcloud secrets versions add gemini-api-key \
  --project=YOUR-PROJECT-ID \
  --data-file=-

# Discord Webhook URL（Discordサーバー設定から取得）
echo -n "https://discord.com/api/webhooks/YOUR_WEBHOOK" | \
  gcloud secrets versions add discord-webhook-url \
  --project=YOUR-PROJECT-ID \
  --data-file=-
```

### 5. GitHub Secretsの設定

```bash
# デプロイに必要な権限をサービスアカウントに追加
gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/cloudfunctions.developer"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/cloudbuild.builds.editor"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/run.developer"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding YOUR-PROJECT-ID \
  --member="serviceAccount:rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"

# サービスアカウントのJSONキーを作成
gcloud iam service-accounts keys create ~/gcp-sa-key.json \
  --iam-account=rss-curator-function@YOUR-PROJECT-ID.iam.gserviceaccount.com \
  --project=YOUR-PROJECT-ID

# GitHub CLIでシークレット設定（またはGitHub UIから手動設定）
gh secret set GCP_SA_KEY --body "$(cat ~/gcp-sa-key.json)"

# セキュリティのため削除
rm ~/gcp-sa-key.json
```

### 6. config.jsonをカスタマイズ

自分の興味のあるトピックやRSSソースに変更：

```bash
# config.jsonを編集
vim config.json

# コミット＆プッシュ
git add config.json
git commit -m "config: 自分の好みに変更"
git push origin main
```

### 7. デプロイ確認

- GitHub Actions: `https://github.com/YOUR_USERNAME/discord-article-bot/actions`
- GCP Console: Cloud Functions、Cloud Schedulerを確認
- 翌朝9:00 JSTに自動実行

### 費用について

無料枠で運用可能（月間約$0-5）：
- Cloud Functions: 毎日1回実行で無料枠内
- Firestore: 数千記事まで無料
- Cloud Scheduler: 3ジョブまで無料
- Gemini API: 月15RPMまで無料

## ローカル開発

```bash
git clone <repository-url>
cd discord-article-bot
go mod download
# config.jsonを編集
go test ./...
```

## ドキュメント

- [仕様書](specs/001-rss-article-curator/spec.md)
- [実装計画](specs/001-rss-article-curator/plan.md)
- [クイックスタート](specs/001-rss-article-curator/quickstart.md)

## ライセンス

MIT License
