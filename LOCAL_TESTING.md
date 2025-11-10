# ローカルテスト手順

このドキュメントでは、RSS記事キュレーションBotをローカル環境でテストする手順を説明します。

## 前提条件

- Go 1.21以上がインストールされていること
- Google Cloud SDK（gcloud CLI）がインストールされていること
- Discord Webhook URLを取得済みであること
- Google Gemini API Keyを取得済みであること

## 1. Firestore エミュレータのセットアップ

### エミュレータコンポーネントのインストール

```bash
gcloud components install cloud-firestore-emulator
```

### エミュレータの起動

```bash
gcloud emulators firestore start --host-port=localhost:8080
```

エミュレータは別のターミナルウィンドウで起動し、バックグラウンドで実行させておきます。

## 2. 環境変数の設定

プロジェクトルートに `.env` ファイルを作成し、以下の内容を記載します:

```bash
# Google Cloud Project設定
GCP_PROJECT_ID=local-test-project

# ローカルテスト用のシークレット
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_WEBHOOK_URL_HERE
GEMINI_API_KEY=YOUR_GEMINI_API_KEY_HERE

# Firestoreエミュレータ設定
FIRESTORE_EMULATOR_HOST=localhost:8080

# ローカル開発用
USE_LOCAL_CONFIG=true
LOCAL_CONFIG_PATH=./config.json
```

### Discord Webhook URLの取得方法

1. Discordサーバーの設定を開く
2. 「連携サービス」→「ウェブフック」を選択
3. 「新しいウェブフック」をクリック
4. ウェブフックURLをコピーして `.env` に設定

### Gemini API Keyの取得方法

1. [Google AI Studio](https://makersuite.google.com/app/apikey)にアクセス
2. 「Get API Key」をクリック
3. APIキーを生成してコピーし、`.env` に設定

## 3. 依存関係のインストール

```bash
go mod download
```

## 4. ローカルテストの実行

ローカルテスト用のスクリプトを実行します:

```bash
go run cmd/local-test/main.go
```

または、ビルドして実行:

```bash
go build -o local-test cmd/local-test/main.go
./local-test
```

## 5. 実行結果の確認

### ログ出力の確認

実行中のログを確認し、以下の項目が正常に動作しているかチェックします:

- ✅ RSSフィードの取得
- ✅ RSSフィードのパース
- ✅ Firestoreエミュレータとの接続
- ✅ 記事コンテンツの取得
- ✅ 記事本文の抽出
- ✅ LLMによる記事評価
- ✅ Discord通知の送信

### Discordでの確認

指定したDiscordチャンネルに記事の通知が届いているか確認します。

### Firestoreエミュレータでの確認

Firestoreエミュレータには、通知済み記事と却下済み記事が保存されます。
エミュレータのWebインターフェースで確認できます（通常は http://localhost:4000 でアクセス可能）。

## トラブルシューティング

### Firestoreエミュレータが起動しない

- `cloud-firestore-emulator` コンポーネントがインストールされているか確認
- ポート8080が他のプロセスで使用されていないか確認

### Gemini APIエラーが発生する

- APIキーが正しく設定されているか確認
- APIキーの有効期限や利用制限を確認

### Discord通知が送信されない

- Webhook URLが正しく設定されているか確認
- Webhook URLが有効かどうかをテスト（curlコマンドなど）

### RSSフィードの取得に失敗する

- インターネット接続を確認
- タイムアウト設定を `config.json` で調整

## 設定のカスタマイズ

`config.json` を編集することで、以下の設定をカスタマイズできます:

- **RSSソース**: 記事を取得するRSSフィードのURL
- **興味トピック**: 記事を評価する際の興味分野
- **通知設定**: 最大通知記事数、最小関連性スコア
- **タイムアウト設定**: RSSフィード取得、記事取得のタイムアウト

詳細は [quickstart.md](./specs/001-rss-article-curator/quickstart.md) を参照してください。
