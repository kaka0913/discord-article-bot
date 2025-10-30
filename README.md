# RSS記事キュレーションBot

技術ブログまとめサイトを毎日監視し、Gemini LLMを使用してユーザー定義の興味に対する記事の関連性を評価し、要約付きの3〜5件のキュレーション記事をDiscordに投稿するサーバーレスRSS記事キュレーター。

## 概要

このプロジェクトは、Google Cloud Functions（Go）で実行され、JST午前9時にCloud Schedulerによってトリガーされ、重複排除追跡にFirestore、認証情報にSecret Managerを使用します。

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
├── cmd/
│   └── curator/          # Cloud Functionsエントリポイント
├── internal/             # 内部パッケージ
│   ├── config/          # 設定管理
│   ├── rss/             # RSSフィード処理
│   ├── article/         # 記事コンテンツ抽出
│   ├── llm/             # Gemini API統合
│   ├── storage/         # Firestore操作
│   ├── discord/         # Discord通知
│   └── secrets/         # Secret Manager統合
├── tests/               # テストファイル
│   ├── contract/        # 契約テスト
│   ├── integration/     # 統合テスト
│   └── unit/           # ユニットテスト
└── terraform/           # インフラストラクチャコード
    ├── environments/
    │   └── prod/
    └── modules/
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

## テスト

```bash
# すべてのテストを実行
go test ./...

# 契約テストのみ
go test ./tests/contract/...

# 統合テストのみ
go test ./tests/integration/...

# ユニットテストのみ
go test ./tests/unit/...
```

## ドキュメント

- [仕様書](specs/001-rss-article-curator/spec.md)
- [実装計画](specs/001-rss-article-curator/plan.md)
- [クイックスタート](specs/001-rss-article-curator/quickstart.md)
- [データモデル](specs/001-rss-article-curator/data-model.md)

## ライセンス

MIT
