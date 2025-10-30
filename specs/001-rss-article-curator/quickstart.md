# クイックスタートガイド: RSS記事キュレーションBot

**機能**: RSS記事キュレーションBot
**最終更新**: 2025-10-27

## 前提条件

実装を開始する前に、以下を準備してください：

### ローカル開発ツール
- Go 1.21+がインストールされていること ([ダウンロード](https://go.dev/dl/))
- Gitがインストールされ、GitHubアカウントが設定されていること
- Cursorがインストールされていること
- `gcloud` CLIがインストールされていること ([インストールガイド](https://cloud.google.com/sdk/docs/install))

### GCPアカウントのセットアップ
- 課金が有効化されたGoogle Cloudアカウント（無料枠で十分）
- 新しいGCPプロジェクトの作成（例: `rss-curator-bot`）
- 設定用のプロジェクトIDをメモ

### 必要なAPIの有効化
```bash
gcloud services enable cloudfunctions.googleapis.com
gcloud services enable cloudscheduler.googleapis.com
gcloud services enable pubsub.googleapis.com
gcloud services enable firestore.googleapis.com
gcloud services enable secretmanager.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

### 外部サービス
- **Discord Webhook URL**: Discordサーバー設定でwebhookを作成
- **Gemini APIキー**: [Google AI Studio](https://aistudio.google.com/app/apikey)から無料APIキーを取得

---

## プロジェクトセットアップ (5分)

### 1. リポジトリのクローン

```bash
git clone https://github.com/<your-org>/discord-article-bot.git
cd discord-article-bot
git checkout 001-rss-article-curator
```

### 2. Goモジュールの初期化

```bash
go mod init github.com/<your-org>/discord-article-bot
go mod tidy
```

### 3. 依存関係のインストール

```bash
# コア依存関係
go get github.com/PuerkitoBio/goquery
go get github.com/go-shiori/go-readability
go get golang.org/x/time/rate
go get cloud.google.com/go/firestore
go get cloud.google.com/go/secretmanager/apiv1
go get google.golang.org/genai

# テスト依存関係
go get github.com/stretchr/testify
```

### 4. プロジェクト構造の作成

```bash
mkdir -p cmd/curator
mkdir -p internal/{config,rss,article,llm,storage,discord,secrets}
mkdir -p tests/{contract,integration,unit}
touch cmd/curator/main.go
```

---

## GCP設定 (10分)

### 1. GCPプロジェクトの設定

```bash
export PROJECT_ID="rss-curator-bot"  # あなたのプロジェクトIDに置き換え
gcloud config set project $PROJECT_ID
```

### 2. Firestoreの初期化

```bash
# Firestoreデータベースの作成（ネイティブモード）
gcloud firestore databases create --region=asia-northeast1
```

**注意**: Discordサーバーに最も近いリージョンを選択してください（例: 日本の場合は `asia-northeast1`、米国の場合は `us-central1`）

### 3. Secret Managerへのシークレット保存

```bash
# Discord Webhook URLを保存
echo -n "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN" | \
  gcloud secrets create discord-webhook-url --data-file=-

# Gemini APIキーを保存
echo -n "YOUR_GEMINI_API_KEY" | \
  gcloud secrets create gemini-api-key --data-file=-

# Cloud Functionsにシークレットへのアクセス権を付与
gcloud secrets add-iam-policy-binding discord-webhook-url \
  --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding gemini-api-key \
  --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### 4. Pub/Subトピックの作成

```bash
# Cloud Schedulerが関数をトリガーするためのトピックを作成
gcloud pubsub topics create curator-trigger
```

### 5. Firestoreセキュリティルールのデプロイ

```bash
# firestore.rulesファイルを作成
cat > firestore.rules <<EOF
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /{document=**} {
      allow read, write: if false;  // バックエンドのみのアクセス
    }
  }
}
EOF

# ルールをデプロイ
gcloud firestore rules deploy firestore.rules
```

---

## 設定ファイル (5分)

### 1. config.jsonの作成

```bash
cat > config.json <<EOF
{
  "rss_sources": [
    {
      "url": "https://dev.to/feed",
      "name": "Dev.to",
      "enabled": true
    },
    {
      "url": "https://zenn.dev/feed",
      "name": "Zenn",
      "enabled": true
    }
  ],
  "interests": [
    {
      "topic": "Go",
      "aliases": ["Golang", "Go言語"],
      "priority": "high"
    },
    {
      "topic": "Kubernetes",
      "aliases": ["k8s"],
      "priority": "high"
    },
    {
      "topic": "Rust",
      "aliases": [],
      "priority": "medium"
    }
  ],
  "notification_settings": {
    "max_articles": 5,
    "min_articles": 3,
    "min_relevance_score": 70
  }
}
EOF
```

### 2. config.jsonをGitHubにプッシュ

```bash
git add config.json
git commit -m "Add initial configuration"
git push origin 001-rss-article-curator
```

**注意**: 設定はGitHub raw URLから取得されます: `https://raw.githubusercontent.com/<your-org>/discord-article-bot/001-rss-article-curator/config.json`

---

## ローカル開発 (10分)

### 1. ローカルテスト用の.envの作成

```bash
cat > .env <<EOF
PROJECT_ID=${PROJECT_ID}
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN
GEMINI_API_KEY=YOUR_GEMINI_API_KEY
CONFIG_URL=https://raw.githubusercontent.com/<your-org>/discord-article-bot/001-rss-article-curator/config.json
FIRESTORE_EMULATOR_HOST=localhost:8080  # ローカルテスト用
EOF
```

**セキュリティ**: `.env`を`.gitignore`に追加してください（シークレットは絶対にコミットしないこと！）

### 2. Firestoreエミュレータの実行（テスト用）

```bash
# Firestoreエミュレータのインストール
gcloud components install cloud-firestore-emulator

# バックグラウンドでエミュレータを起動
gcloud emulators firestore start --host-port=localhost:8080 &

# エミュレータはhttp://localhost:8080で実行されます
```

### 3. テストの実行

```bash
# ユニットテスト（高速、外部依存なし）
go test ./internal/... -v

# 統合テスト（Firestoreエミュレータが必要）
go test ./tests/integration/... -v

# 契約テスト（モックHTTPサーバー）
go test ./tests/contract/... -v

# カバレッジ付きの全テスト
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out  # ブラウザで開く
```

### 4. 関数のローカル実行

```bash
# Go用のFunctions Frameworkを使用
go install github.com/GoogleCloudPlatform/functions-framework-go/funcframework@latest

# 環境をセットアップ
export $(cat .env | xargs)

# 関数をローカルで実行
funcframework --target=CuratorHandler --port=8081

# 関数をトリガー
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"data":"e30="}'  # Base64エンコードされた空のJSON
```

---

## デプロイ (10分)

### 1. Cloud Functionのデプロイ

```bash
gcloud functions deploy curator \
  --gen2 \
  --runtime=go121 \
  --region=asia-northeast1 \
  --source=. \
  --entry-point=CuratorHandler \
  --trigger-topic=curator-trigger \
  --timeout=3600s \
  --memory=512MB \
  --set-env-vars PROJECT_ID=${PROJECT_ID},CONFIG_URL=https://raw.githubusercontent.com/<your-org>/discord-article-bot/001-rss-article-curator/config.json \
  --set-secrets DISCORD_WEBHOOK_URL=discord-webhook-url:latest,GEMINI_API_KEY=gemini-api-key:latest
```

**注意**: デプロイには2〜3分かかります。関数はPub/Subメッセージによってトリガーされます。

### 2. Cloud Schedulerジョブの作成

```bash
gcloud scheduler jobs create pubsub curator-daily-trigger \
  --location=asia-northeast1 \
  --schedule="0 9 * * *" \
  --time-zone="Asia/Tokyo" \
  --topic=curator-trigger \
  --message-body='{"trigger":"scheduled"}'
```

**スケジュール**: 毎日午前9:00 JST（cron: `0 9 * * *`）に実行されます

### 3. スケジュールトリガーの手動テスト

```bash
# 関数を即座にトリガー（午前9時を待たない）
gcloud scheduler jobs run curator-daily-trigger --location=asia-northeast1

# 関数ログの表示
gcloud functions logs read curator --gen2 --region=asia-northeast1 --limit=50
```

---

## デプロイの確認 (5分)

### 1. 関数の実行チェック

```bash
# リアルタイムログの表示
gcloud functions logs read curator --gen2 --region=asia-northeast1 --limit=100 --format="table(time,log)"
```

**期待される出力**:
```
TIME                       LOG
2025-10-27T00:15:01.234Z   RSSキュレーターを開始中...
2025-10-27T00:15:02.456Z   設定をロード: 2つのRSSソース、3つの興味
2025-10-27T00:15:15.789Z   Dev.toから120件の記事を取得
2025-10-27T00:15:28.012Z   Zennから85件の記事を取得
2025-10-27T00:20:45.678Z   205件の記事を評価、5件が関連
2025-10-27T00:20:50.123Z   Discordに5件の記事を投稿 (message_id: 1234567890123456789)
2025-10-27T00:20:50.456Z   キュレーター正常完了
```

### 2. Discordチャンネルのチェック

- webhookのあるDiscordチャンネルに移動
- 3〜5件の記事埋め込みのメッセージが投稿されていることを確認
- 埋め込みにタイトル、要約、関連性スコア、トピックが表示されていることを確認

### 3. Firestoreコレクションのチェック

```bash
# Firestore CLIツールのインストール
npm install -g @google-cloud/firestore

# 通知された記事をクエリ
gcloud firestore documents list notified_articles --limit=10

# 却下された記事をクエリ
gcloud firestore documents list rejected_articles --limit=10
```

---

## 監視とデバッグ

### Cloud Functionメトリクスの表示

```bash
# Cloud Consoleメトリクスダッシュボードを開く
echo "https://console.cloud.google.com/functions/details/asia-northeast1/curator?project=${PROJECT_ID}&tab=metrics"
```

**主要メトリクス**:
- **呼び出し回数**: 1日1回であるべき
- **実行時間**: 30〜60分であるべき
- **メモリ使用量**: 300 MB未満であるべき
- **エラー**: 0であるべき

### Cloud Consoleでのログ表示

```bash
# Cloud Loggingを開く
echo "https://console.cloud.google.com/logs/query?project=${PROJECT_ID}&query=resource.labels.function_name%3D%22curator%22"
```

**フィルター例**:
- `severity>=ERROR`: エラーのみ表示
- `jsonPayload.relevance_score>90`: 高関連性の記事を表示
- `textPayload=~"rate limit"`: レート制限の問題を見つける

### よくある問題と解決策

#### 1. 関数がタイムアウト（60分制限）

**症状**: ログにタイムアウトエラーが表示され、関数が実行途中で停止する

**解決策**:
- config.jsonのRSSソース数を減らす
- min_relevance_scoreしきい値を上げる（評価する記事を減らす）
- Gemini APIレート制限を確認（15 RPMは200件以上の記事には遅すぎる可能性）

#### 2. Discord Webhook 404エラー

**症状**: ログに「Unknown Webhook」エラーが表示される

**解決策**:
- Secret ManagerのwebhookURLが正しいことを確認
- Discordサーバー設定でwebhookがまだ存在することを確認
- シークレットを更新: `echo -n "NEW_URL" | gcloud secrets versions add discord-webhook-url --data-file=-`

#### 3. Gemini API 429レート制限

**症状**: ログに「RESOURCE_EXHAUSTED」エラーが表示される

**解決策**:
- レートリミッターがこれを防ぐはず（15 RPM = リクエスト間4秒）
- 日次クォータ（1500 RPD）を超過していないか確認
- 24時間待ってクォータがリセットされるか、有料プランにアップグレード

#### 4. Firestore Permission Denied

**症状**: ログに「PermissionDenied」エラーが表示される

**解決策**:
- サービスアカウントにFirestoreロールがあることを確認:
  ```bash
  gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com" \
    --role="roles/datastore.user"
  ```

---

## 設定の更新 (2分)

### RSSソースまたは興味の変更

1. リポジトリ内の`config.json`を編集:
   ```bash
   # 新しいRSSソースを追加
   nano config.json  # またはGitHub webエディタを使用
   ```

2. 変更をコミットしてプッシュ:
   ```bash
   git add config.json
   git commit -m "Add Hacker News RSS source"
   git push origin 001-rss-article-curator
   ```

3. 次回のスケジュール実行（午前9時JST）を待つか、手動でトリガー:
   ```bash
   gcloud scheduler jobs run curator-daily-trigger --location=asia-northeast1
   ```

**再デプロイ不要！** 関数は実行ごとにGitHubから最新のconfig.jsonを取得します。

---

## CI/CDセットアップ（GitHub Actions）

### 1. GitHub Actionsワークフローの作成

```bash
mkdir -p .github/workflows
cat > .github/workflows/deploy.yml <<EOF
name: Deploy Curator Function

on:
  push:
    branches: [main, 001-rss-article-curator]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: google-github-actions/setup-gcloud@v1
        with:
          service_account_key: \${{ secrets.GCP_SA_KEY }}
          project_id: ${PROJECT_ID}
          export_default_credentials: true

      - name: Deploy Cloud Function
        run: |
          gcloud functions deploy curator \\
            --gen2 \\
            --runtime=go121 \\
            --region=asia-northeast1 \\
            --source=. \\
            --entry-point=CuratorHandler \\
            --trigger-topic=curator-trigger \\
            --timeout=3600s \\
            --memory=512MB \\
            --set-env-vars PROJECT_ID=${PROJECT_ID},CONFIG_URL=https://raw.githubusercontent.com/<your-org>/discord-article-bot/001-rss-article-curator/config.json \\
            --set-secrets DISCORD_WEBHOOK_URL=discord-webhook-url:latest,GEMINI_API_KEY=gemini-api-key:latest
EOF
```

### 2. GitHubシークレットへのサービスアカウントキーの追加

1. サービスアカウントを作成:
   ```bash
   gcloud iam service-accounts create github-actions --display-name="GitHub Actions Deployer"

   gcloud projects add-iam-policy-binding ${PROJECT_ID} \
     --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
     --role="roles/cloudfunctions.developer"

   gcloud iam service-accounts keys create key.json \
     --iam-account=github-actions@${PROJECT_ID}.iam.gserviceaccount.com
   ```

2. `key.json`の内容をGitHubリポジトリシークレットに`GCP_SA_KEY`として追加

3. **セキュリティ**: GitHubにアップロード後、ローカルの`key.json`を削除:
   ```bash
   rm key.json
   ```

---

## コスト監視

### 予想月額コスト（無料枠）

| サービス | 使用量 | コスト |
|---------|-------|------|
| Cloud Functions | 月30回の呼び出し、30時間のコンピューティング | $0（無料枠: 200万回の呼び出し） |
| Cloud Scheduler | 1ジョブ、月30回の実行 | $0（無料枠: 3ジョブ） |
| Pub/Sub | 月30メッセージ | $0（無料枠: 月10 GB） |
| Firestore | 月10,000回の読み取り、500回の書き込み | $0（無料枠: 5万回の読み取り、2万回の書き込み） |
| Secret Manager | 2つのシークレット、月30回のアクセス | $0（無料枠: 6シークレット、1万回のアクセス） |
| Gemini API | 月6,000リクエスト（1日200回） | $0（無料枠: 1500 RPD） |
| **合計** | | **$0/月** |

### 課金アラートの設定

```bash
# $1で予算アラートを作成（セーフティネット）
gcloud billing budgets create \
  --billing-account=YOUR_BILLING_ACCOUNT_ID \
  --display-name="RSS Curator Budget Alert" \
  --budget-amount=1 \
  --threshold-rule=percent=50 \
  --threshold-rule=percent=90 \
  --threshold-rule=percent=100
```

---

## 次のステップ

デプロイが成功した後:

1. **最初の1週間を監視**: 毎日Discordをチェックして記事が関連性があることを確認
2. **min_relevance_scoreを調整**: 記事が多すぎる/少なすぎる場合はconfig.jsonで調整
3. **RSSソースを追加**: 3〜5ソースに拡張（Hacker News、Hashnodeなど）
4. **興味を精緻化**: 記事の品質に基づいてトピックを追加/削除
5. **アラートの設定**: 関数の失敗に対するCloud Monitoringアラートを設定

---

## ロールバック手順

デプロイに問題が発生した場合:

```bash
# 最近のデプロイをリスト
gcloud functions versions list curator --gen2 --region=asia-northeast1

# 前のバージョンにロールバック
gcloud functions rollback curator \
  --gen2 \
  --region=asia-northeast1 \
  --to-version=VERSION_ID

# または特定のgitコミットから再デプロイ
git checkout PREVIOUS_COMMIT_SHA
gcloud functions deploy curator [... デプロイと同じフラグ ...]
```

---

## トラブルシューティングコマンド

```bash
# 関数の詳細を表示
gcloud functions describe curator --gen2 --region=asia-northeast1

# Pub/Subトピックのサブスクリプションを表示
gcloud pubsub topics list-subscriptions curator-trigger

# スケジューラージョブのステータスを表示
gcloud scheduler jobs describe curator-daily-trigger --location=asia-northeast1

# Gemini API接続テスト
curl -X POST "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"parts":[{"text":"Hello"}]}]}'

# Discord webhookテスト
curl -X POST "YOUR_DISCORD_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{"content":"curlからのテストメッセージ"}'
```

---

## サポートリソース

- **GCPドキュメント**: https://cloud.google.com/functions/docs
- **Gemini APIドキュメント**: https://ai.google.dev/docs
- **Discord Webhook API**: https://discord.com/developers/docs/resources/webhook
- **Go Firestore SDK**: https://pkg.go.dev/cloud.google.com/go/firestore
- **プロジェクトIssues**: https://github.com/<your-org>/discord-article-bot/issues
