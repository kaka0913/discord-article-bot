# RSS記事キュレーションBot

技術ブログまとめサイトを毎日監視し、Gemini LLMを使用してユーザー定義の興味に対する記事の関連性を評価し、要約付きの3〜5件のキュレーション記事をDiscordに投稿するサーバーレスRSS記事キュレーター。

## 概要

このプロジェクトは、Google Cloud Functions（Go）で実行され、JST午前8時にCloud Schedulerによってトリガーされ、重複排除追跡にFirestore、認証情報にSecret Managerを使用します。

### 主要機能

- 毎日の自動実行（JST午前8時）
- 複数のRSSフィードからの記事収集
- Gemini API v2.0による記事の関連性評価とスコアリング
- AI生成記事の自動検出と除外
- **記事全体のサマリー生成** - 選択された記事全体の傾向分析
- Discord Webhookによる通知
- Firestoreによる重複排除（30日間TTL）

## 技術スタック

- **言語**: Go 1.21+
- **プラットフォーム**: Google Cloud Functions Gen 2
- **インフラ**: Terraform
- **ストレージ**: Firestore
- **LLM**: Google Gemini Flash API
- **通知**: Discord Webhook

## プロジェクト構造

```
.
├── .github/
│   ├── workflows/
│   │   ├── deploy.yml          # 自動デプロイワークフロー
│   │   └── test.yml            # PRテストワークフロー
│   └── PULL_REQUEST_TEMPLATE.md # PRテンプレート
├── cmd/
│   ├── curator/                 # Cloud Functions本番環境用
│   └── local-test/              # ローカルテスト用
├── internal/                    # 内部パッケージ
│   ├── config/                 # 設定管理
│   ├── secrets/                # Secret Manager統合
│   ├── errors/                 # エラーハンドリング
│   ├── logging/                # 構造化ログ
│   ├── storage/                # Firestore操作
│   ├── rss/                    # RSSフィード処理
│   ├── article/                # 記事コンテンツ抽出
│   ├── llm/                    # Gemini API統合（評価、サマリー生成）
│   └── discord/                # Discord通知
├── tests/                      # テストファイル
│   └── contract/               # 契約テスト（Discord, Firestore, Gemini, RSS）
├── terraform/                  # インフラストラクチャコード
│   ├── environments/
│   │   └── prod/
│   └── modules/
│       ├── firestore/
│       ├── secrets/
│       ├── scheduler/
│       └── cloud-function/
├── specs/                      # 設計ドキュメント
│   └── 001-rss-article-curator/
├── config.json                 # 記事の好み設定（カスタマイズ可能）
└── cloudbuild.yaml             # Cloud Build設定
```

## 記事の好みをカスタマイズする

このBotは`config.json`を編集することで、通知される記事の内容を自由にカスタマイズできます。

### 設定ファイルの構造

```json
{
  "rss_sources": [/* RSSフィードのリスト */],
  "interests": [/* 興味のあるトピック */],
  "notification_settings": {/* 通知設定 */}
}
```

### カスタマイズ可能な項目

#### 1. 興味のトピック (`interests`)

```json
{
  "topic": "Go言語",
  "aliases": ["Golang", "Go"],
  "priority": "high"  // high, medium, low
}
```

- **topic**: メインのトピック名
- **aliases**: 記事内で検索する別名のリスト
- **priority**: 優先度（high/medium/low）

#### 2. 通知設定 (`notification_settings`)

```json
{
  "max_articles": 5,           // 1日の最大記事数（1-10）
  "min_articles": 1,           // 最小記事数
  "min_relevance_score": 70    // 最小関連性スコア（0-100）
}
```

- **min_relevance_score**: この値を上げると厳選された記事のみ、下げると幅広い記事が通知されます

#### 3. RSSソース (`rss_sources`)

```json
{
  "name": "dev.to",
  "url": "https://dev.to/feed",
  "enabled": true
}
```

- **enabled**: `false`にすることで一時的にソースを無効化できます

### 設定変更の手順

#### GitHubを使う場合（推奨）

1. **リポジトリをフォーク**（初回のみ）
   ```bash
   # GitHubでフォークボタンをクリック
   git clone https://github.com/YOUR_USERNAME/discord-article-bot.git
   cd discord-article-bot
   ```

2. **ブランチを作成**
   ```bash
   git checkout -b config/update-interests
   ```

3. **config.jsonを編集**
   - 興味のトピックを追加・削除
   - スコア閾値を調整
   - RSSソースを追加・無効化

4. **変更をコミット**
   ```bash
   git add config.json
   git commit -m "config: 機械学習の優先度をhighに変更"
   git push origin config/update-interests
   ```

5. **プルリクエストを作成**
   - GitHubでプルリクエストを作成
   - 自動テストが実行され、設定の妥当性が検証されます

6. **マージ**
   - レビュー後、mainブランチにマージ
   - 自動デプロイが実行されます
   - **翌朝9:00 JSTから新しい設定で記事が通知されます**

#### 直接編集する場合

mainブランチを直接編集する権限がある場合：

```bash
git checkout main
git pull origin main
# config.jsonを編集
git add config.json
git commit -m "config: 興味のトピックを更新"
git push origin main
# 自動デプロイが実行されます
```

### 設定例

#### 機械学習に特化したい場合

```json
{
  "interests": [
    {
      "topic": "機械学習",
      "aliases": ["Machine Learning", "ML", "Deep Learning", "AI"],
      "priority": "high"
    },
    {
      "topic": "Python",
      "aliases": ["Python3", "Py"],
      "priority": "high"
    },
    {
      "topic": "TensorFlow",
      "aliases": ["Keras", "PyTorch"],
      "priority": "medium"
    }
  ],
  "notification_settings": {
    "max_articles": 3,
    "min_relevance_score": 75
  }
}
```

#### 幅広いトピックを受け取りたい場合

```json
{
  "notification_settings": {
    "max_articles": 8,
    "min_relevance_score": 60
  }
}
```

## セットアップ

### 前提条件

- Go 1.21以上
- Google Cloud SDK
- Terraform
- Google Cloud Projectとその権限

### ローカル開発

1. リポジトリをクローン
```bash
git clone <repository-url>
cd rss-article-curator
```

2. 環境変数を設定
```bash
cp .env.example .env
# .envファイルを編集して必要な値を設定
```

3. 依存関係をインストール
```bash
go mod download
```

4. config.jsonを編集してRSSソースと興味を設定

### デプロイ

詳細は `specs/001-rss-article-curator/quickstart.md` を参照してください。

## CI/CD

このプロジェクトはGitHub Actionsを使用した自動テスト・デプロイパイプラインを備えています。

### 自動ワークフロー

#### プルリクエスト（test.yml）

プルリクエストを作成すると、以下が自動実行されます：

1. ✅ **ユニットテスト**: すべてのGoテストを実行
2. ✅ **config.json検証**: JSON構文と必須フィールドをチェック
3. ✅ **スキーマ検証**: 設定値の妥当性を確認
4. 📊 **テスト結果コメント**: PRに結果を自動投稿

テスト結果はPRページで確認できます：

```
## テスト結果 🧪

### ユニットテスト
✅ すべてのテストが合格しました
- カバレッジ: 85.2%

### config.json検証
✅ 設定ファイルは有効です
- RSSソース数: 3
- 興味トピック数: 5
```

#### mainブランチへのマージ（deploy.yml）

mainブランチにマージされると、以下が自動実行されます：

1. ✅ **テスト実行**: 再度すべてのテストを実行
2. 📦 **ビルド**: Cloud Functions用のデプロイパッケージを作成
3. 🚀 **デプロイ**: Google Cloud Functionsに自動デプロイ
4. ✔️ **検証**: デプロイが成功したことを確認

デプロイ状況は[GitHub Actions](../../actions)タブで確認できます。

### デプロイトリガー

以下のファイルが変更されると、自動デプロイが実行されます：

- `cmd/curator/**` - メイン処理コード
- `internal/**` - 内部パッケージ
- `go.mod`, `go.sum` - 依存関係
- **`config.json`** - 設定ファイル（これが最も頻繁に変更されます）

### セットアップ手順（リポジトリ管理者向け）

GitHub Actionsを有効にするには、以下のシークレットを設定してください：

1. GitHubリポジトリの Settings > Secrets and variables > Actions へ移動
2. 以下のシークレットを追加：

| シークレット名 | 説明 | 取得方法 |
|------------|------|---------|
| `GCP_SA_KEY` | サービスアカウントのJSONキー | GCP Console > IAM > Service Accounts |
| `GCP_PROJECT_ID` | GCPプロジェクトID | `rss-article-curator-prod` |

サービスアカウントには以下の権限が必要です：
- Cloud Functions Developer
- Service Account User

### 手動デプロイ

GitHub Actions経由ではなく、手動でデプロイする場合：

```bash
# gcloudコマンドでデプロイ
gcloud functions deploy rss-article-curator \
  --gen2 \
  --region=asia-northeast1 \
  --runtime=go122 \
  --source=/tmp/function-deploy \
  --entry-point=CuratorHandler \
  --trigger-http \
  --no-allow-unauthenticated \
  --service-account=rss-curator-function@rss-article-curator-prod.iam.gserviceaccount.com \
  --memory=512Mi \
  --timeout=3600s \
  --max-instances=1 \
  --min-instances=0 \
  --set-env-vars=CONFIG_URL=https://raw.githubusercontent.com/kaka0913/discord-article-bot/main/config.json,GCP_PROJECT_ID=rss-article-curator-prod,GEMINI_API_KEY_SECRET=gemini-api-key,DISCORD_WEBHOOK_SECRET=discord-webhook-url \
  --project=rss-article-curator-prod
```

### デプロイ後の確認

1. **Cloud Functionsダッシュボード**で状態を確認
   ```bash
   gcloud functions describe rss-article-curator \
     --region=asia-northeast1 \
     --project=rss-article-curator-prod
   ```

2. **ログを確認**
   ```bash
   gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=rss-article-curator" \
     --limit=50 \
     --project=rss-article-curator-prod
   ```

3. **翌朝9:00 JST**に記事が通知されることを確認

## テスト

```bash
# すべてのテストを実行
go test ./...

# 契約テストのみ（実装済み）
go test ./tests/contract/...

# 各パッケージのユニットテスト
go test ./internal/config/...
go test ./internal/errors/...
go test ./internal/logging/...
go test ./internal/secrets/...
```

## ドキュメント

- [AGENT.md](AGENT.md) - エージェント指示書、プロジェクト概要
- [仕様書](specs/001-rss-article-curator/spec.md)
- [実装計画](specs/001-rss-article-curator/plan.md)
- [タスクリスト](specs/001-rss-article-curator/tasks.md)
- [クイックスタート](specs/001-rss-article-curator/quickstart.md)
- [データモデル](specs/001-rss-article-curator/data-model.md)

## ライセンス

MIT License
