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

## プロジェクト構造

```
.
├── .github/workflows/      # CI/CDワークフロー
├── cmd/curator/            # Cloud Functions本番用
├── internal/               # 内部パッケージ
├── terraform/              # インフラコード
├── config.json             # 記事の好み設定
└── cloudbuild.yaml         # Cloud Build設定
```

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
