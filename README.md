# RSS記事キュレーションBot

技術ブログまとめサイトを毎日監視し、Gemini LLMを使用してユーザー定義の興味に対する記事の関連性を評価し、要約付きの3〜5件のキュレーション記事をDiscordに投稿するサーバーレスRSS記事キュレーター。

## 概要

このプロジェクトは、Google Cloud Functions（Go）で実行され、JST午前8時にCloud Schedulerによってトリガーされ、重複排除追跡にFirestore、認証情報にSecret Managerを使用します。

**現在のステータス**: 実装とローカルテストが完了し、GCPデプロイ待ちです（8/9タスク完了）。

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
├── cmd/
│   ├── curator/          # Cloud Functions本番環境用（✅ 実装済み）
│   └── local-test/       # ローカルテスト用（✅ 実装済み）
├── internal/             # 内部パッケージ（✅ すべて実装済み）
│   ├── config/          # 設定管理
│   ├── secrets/         # Secret Manager統合
│   ├── errors/          # エラーハンドリング
│   ├── logging/         # 構造化ログ
│   ├── storage/         # Firestore操作
│   ├── rss/             # RSSフィード処理
│   ├── article/         # 記事コンテンツ抽出
│   ├── llm/             # Gemini API統合（評価、サマリー生成）
│   └── discord/         # Discord通知
├── tests/               # テストファイル（✅ 契約テスト実装済み）
│   └── contract/        # 契約テスト（Discord, Firestore, Gemini, RSS）
├── terraform/           # インフラストラクチャコード（✅ 実装済み）
│   ├── environments/
│   │   └── prod/
│   └── modules/
│       ├── firestore/
│       ├── secrets/
│       ├── scheduler/
│       └── cloud-function/
└── specs/               # 設計ドキュメント
    └── 001-rss-article-curator/
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
