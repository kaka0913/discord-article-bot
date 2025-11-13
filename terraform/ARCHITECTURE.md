# インフラ構成

RSS記事キュレーションBotのインフラ構成について説明します。

## アーキテクチャの流れ

1. **Cloud Scheduler** が毎日9:00 JSTにHTTP POSTリクエストで **Cloud Functions** をトリガー
2. **Cloud Functions** が以下の処理を順次実行：
   - **RSS Feeds** から技術ブログの記事を取得
   - **Firestore** で重複チェック（過去に通知済み・却下済みの記事を除外）
   - **Gemini API** で記事の関連性を評価（スコアリング＋トピックマッチング）
   - 評価結果を **Firestore** に保存
   - 関連性の高い記事を **Discord** に通知
3. **Secret Manager** から API Key と Webhook URL を安全に取得
4. すべての処理は **Service Account** の権限で実行

## 主要コンポーネント

### Cloud Scheduler
- **役割**: 定期実行トリガー
- **スケジュール**: 毎日9:00 JST（cron: `0 9 * * *`）
- **認証**: OIDC（Service Account経由）
- **タイムアウト**: 30分（1800秒）
- **リトライ**: 失敗時に自動リトライ

### Cloud Functions Gen2
- **役割**: RSS記事取得・評価・通知の実行
- **ランタイム**: Go 1.22
- **メモリ**: 512Mi
- **タイムアウト**: 60分（3600秒）
- **リージョン**: asia-northeast1
- **トリガー**: HTTP（認証必須）
- **環境変数**:
  - `CONFIG_URL`: GitHub上のconfig.json（記事の好み設定）
  - `GCP_PROJECT_ID`: GCPプロジェクトID
  - `GEMINI_API_KEY_SECRET`: Gemini APIキーのSecret名
  - `DISCORD_WEBHOOK_SECRET`: Discord WebhookのSecret名

### Firestore
- **役割**: 記事の重複管理
- **モード**: ネイティブモード
- **リージョン**: asia-northeast1
- **コレクション**:
  - `notified_articles`: 通知済み記事（TTL: 30日で自動削除）
  - `rejected_articles`: 却下済み記事（TTL: 7日で自動削除）
- **インデックス**: TTL削除用のインデックス設定済み

### Secret Manager
- **役割**: 機密情報の安全な保管
- **シークレット**:
  - `gemini-api-key`: Gemini API認証キー
  - `discord-webhook-url`: Discord通知先URL
- **アクセス制御**: Service Accountのみ読み取り可能

### Service Account
- **名前**: `rss-curator-function@rss-article-curator-prod.iam.gserviceaccount.com`
- **役割**: 最小権限の原則に基づいた実行アカウント
- **権限**:
  - `Secret Manager Secret Accessor`: シークレット読み取り
  - `Cloud Datastore User`: Firestoreの読み書き
  - `Cloud Functions Invoker`: Cloud Schedulerからの関数呼び出し

## 外部連携

### RSS Feeds
- **役割**: 技術ブログのRSSフィードから記事を取得
- **形式**: RSS 2.0 / Atom対応
- **設定**: config.jsonのrss_sourcesで管理

### Gemini API
- **役割**: Google製LLMで記事の関連性を評価
- **モデル**: Gemini Pro
- **評価内容**:
  - 関連性スコア（0-100）
  - マッチングトピック抽出
  - 記事要約生成

### Discord
- **役割**: Webhook経由で記事を通知
- **形式**: Discord Embed形式
- **内容**: タイトル、要約、URL、スコア、トピック

## データフロー

```
┌─────────────────┐
│ Cloud Scheduler │ 毎日9:00 JST
└────────┬────────┘
         │ HTTP POST (OIDC)
         ▼
┌─────────────────┐
│ Cloud Functions │
└────────┬────────┘
         │
         ├─► RSS Feeds ────────► 記事取得
         │
         ├─► Firestore ─────────► 重複チェック
         │                        ↓
         │                     結果保存
         │
         ├─► Gemini API ────────► 記事評価
         │
         ├─► Discord ───────────► 通知送信
         │
         └─► Secret Manager ────► 認証情報取得
```

## セキュリティ

### 認証・認可
- Cloud Scheduler → Cloud Functions: OIDC認証
- Cloud Functions → Secret Manager: Service Account権限
- Cloud Functions → Firestore: Service Account権限

### 機密情報管理
- API Key、Webhook URLはSecret Managerで暗号化保存
- コード内にハードコードしない
- 環境変数でSecret名のみ指定

### ネットワーク
- Cloud Functionsは認証なしアクセス不可
- 外部からの直接呼び出しを禁止
- Cloud Schedulerからのみアクセス可能

## コスト最適化

### Cloud Functions
- 最小インスタンス: 0（コールドスタート許容）
- 最大インスタンス: 1（並列実行不要）
- メモリ: 512Mi（必要最小限）

### Firestore
- TTL自動削除で不要データを削除
- インデックスは必要最小限

### Cloud Scheduler
- 1日1回のみ実行
- リトライは最小限

## 監視・運用

### ログ
- Cloud Functions: 自動的にCloud Loggingに出力
- Cloud Scheduler: ジョブ実行ログを記録
- エラーレベル: ERROR、WARN、INFO、DEBUG

### メトリクス
- 実行時間
- エラー率
- 通知記事数
- API呼び出し回数

### アラート
- 連続失敗時にCloud Schedulerが自動リトライ
- 重大なエラーはログに記録

## スケーラビリティ

### 現在の設計
- 1日1回、順次処理
- 最大数十記事を処理
- 単一インスタンスで十分

### 拡張性
- RSS源を追加してもスケール可能
- 複数インスタンス並列実行可能（max_instancesを増やす）
- Firestoreは自動スケール
