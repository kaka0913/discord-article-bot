# タスク: RSS記事キュレーションBot

**入力**: `/specs/001-rss-article-curator/`からの設計ドキュメント
**前提条件**: plan.md、spec.md、data-model.md、contracts/

**整理**: タスクはユーザーストーリーごとにグループ化され、各ストーリーの独立した実装とテストを可能にします。

## 形式: `[ID] [P?] [Story] 説明`

- **[P]**: 並列実行可能（異なるファイル、依存関係なし）
- **[Story]**: このタスクが属するユーザーストーリー（例：US1、US4）
- 説明に正確なファイルパスを含める

---

## フェーズ1：セットアップ（共有インフラ）

**目的**: プロジェクトの初期化と基本構造

- [ ] T001 plan.mdに従ってプロジェクト構造を作成（cmd/、internal/、tests/、terraform/）
- [ ] T002 Goモジュールを初期化（go mod init、go.mod、go.sum）
- [ ] T003 [P] .gitignoreファイルを作成（terraform.tfvars、.env、シークレットを除外）
- [ ] T004 [P] config.jsonテンプレートを作成（RSSソース、興味トピック、通知設定）
- [ ] T005 [P] README.mdを作成（プロジェクト概要、セットアップ手順）

---

## フェーズ2：基盤（ブロッキング前提条件）

**目的**: いずれかのユーザーストーリーを実装する前に完了していなければならないコアインフラ

**⚠️ 重要**: このフェーズが完了するまでユーザーストーリー作業は開始できません

### Terraformインフラ（IaC）

- [ ] T006 terraform/modules/firestore/を作成（Firestore Database + インデックス定義）
- [ ] T007 [P] terraform/modules/secrets/を作成（Secret Manager定義）
- [ ] T008 [P] terraform/modules/scheduler/を作成（Cloud Scheduler + Pub/Sub Topic定義）
- [ ] T009 [P] terraform/modules/cloud-function/を作成（Cloud Function Gen 2定義）
- [ ] T010 terraform/environments/prod/main.tfを作成（モジュールを組み合わせ）
- [ ] T011 [P] terraform/environments/prod/variables.tf、outputs.tf、backend.tfを作成

### 設定管理

- [ ] T012 internal/config/schema.goを実装（RSSSource、InterestTopic、NotificationSettings構造体）
- [ ] T013 internal/config/loader.goを実装（GitHubまたはローカルファイルからconfig.jsonをロード）
- [ ] T014 [P] internal/config/validator.goを実装（設定検証ロジック）

### Secret Manager統合

- [ ] T015 internal/secrets/manager.goを実装（Discord Webhook URL、Gemini APIキーの取得）

### Firestoreクライアント

- [ ] T016 internal/storage/firestore.goを実装（Firestoreクライアントラッパー）
- [ ] T017 internal/storage/notified.goを実装（notified_articlesコレクション操作）
- [ ] T018 internal/storage/rejected.goを実装（rejected_articlesコレクション操作）

### エラー処理とログ

- [ ] T019 internal/errors/errors.goを実装（カスタムエラー型とラッピング）
- [ ] T020 internal/logging/logger.goを実装（構造化ログ設定）

**チェックポイント**: 基盤準備完了 - ユーザーストーリー実装を並列で開始可能

---

## フェーズ3：ユーザーストーリー 4 - 記事の重複排除（優先度: P1）🎯 MVP前提

**目標**: 通知済み記事と却下済み記事をFirestoreで追跡し、重複評価を防止

**独立テスト**: 記事を一度保存し、次回チェック時に「既存」として検出されることを確認

### US4 契約テスト

- [ ] T021 [P] [US4] tests/contract/firestore_test.goを作成（Firestoreエミュレータでの契約テスト）

### US4 実装

- [ ] T022 [US4] internal/storage/notified.goに SaveNotifiedArticle() 実装
- [ ] T023 [US4] internal/storage/notified.goに IsArticleNotified() 実装
- [ ] T024 [US4] internal/storage/rejected.goに SaveRejectedArticle() 実装
- [ ] T025 [US4] internal/storage/rejected.goに IsArticleRejected() 実装

**チェックポイント**: Firestore重複排除が機能し、独立してテスト可能

---

## フェーズ4：ユーザーストーリー 1 - 毎日のキュレーション記事通知（優先度: P1）🎯 MVP

**目標**: RSSフィードを取得し、Geminiで評価し、Discordに通知

**独立テスト**: 手動でCloud Functionをトリガーし、Discordに3〜5件の記事が通知されることを確認

### US1 契約テスト

- [ ] T026 [P] [US1] tests/contract/rss_parser_test.goを作成（RSS 2.0/Atomフィードフィクスチャでテスト）
- [ ] T027 [P] [US1] tests/contract/gemini_api_test.goを作成（Gemini APIリクエスト/レスポンス契約）
- [ ] T028 [P] [US1] tests/contract/discord_webhook_test.goを作成（Discord Embeds APIペイロード契約）

### US1 RSSフェッチ・パース

- [ ] T029 [P] [US1] internal/rss/fetcher.goを実装（RSSフィードURLからHTMLを取得）
- [ ] T030 [P] [US1] internal/rss/parser.goを実装（RSS 2.0/Atom XMLをArticle構造体にパース）

### US1 記事コンテンツ抽出

- [ ] T031 [P] [US1] internal/article/fetcher.goを実装（記事URLからHTMLを取得）
- [ ] T032 [P] [US1] internal/article/extractor.goを実装（go-readabilityで記事テキストを抽出）

### US1 LLM評価

- [ ] T033 [US1] internal/llm/client.goを実装（レート制限付きGemini APIクライアント、golang.org/x/time/rate使用）
- [ ] T034 [US1] internal/llm/evaluator.goを実装（記事の関連性を評価、AI生成記事判定、スコア0〜100）
- [ ] T035 [US1] internal/llm/summarizer.goを実装（記事要約を生成、50〜200文字）

### US1 Discord通知

- [ ] T036 [P] [US1] internal/discord/client.goを実装（Discord Webhook APIクライアント）
- [ ] T037 [P] [US1] internal/discord/formatter.goを実装（記事をEmbedsペイロードとしてフォーマット）

### US1 Cloud Functionエントリーポイント

- [ ] T038 [US1] cmd/curator/main.goを実装（Pub/SubトリガーCloud Functionエントリーポイント）
- [ ] T039 [US1] メインオーケストレーションロジックを実装：
  - config.jsonをロード
  - RSSフィードを取得・パース
  - 重複チェック（Firestore）
  - 記事コンテンツ抽出
  - Geminiで評価
  - 上位3〜5件を選択
  - Discordに通知
  - Firestoreに保存

**チェックポイント**: この時点でUS1は完全に機能し、手動トリガーで独立してテスト可能

---

## フェーズ5：ユーザーストーリー 2 - 設定可能な興味管理（優先度: P2）

**目標**: 興味トピックを外部設定（GitHub）から動的にロード

**独立テスト**: config.jsonの興味を更新し、次回実行で新しい興味が反映されることを確認

### US2 実装

- [ ] T040 [US2] internal/config/loader.goにGitHub Raw URLからの設定ロード機能を追加
- [ ] T041 [US2] 優先度重み付けロジックを実装（high: 2.0x, medium: 1.0x, low: 0.5x）
- [ ] T042 [US2] エイリアス対応をinternal/llm/evaluator.goに追加

**チェックポイント**: 設定の動的ロードが機能し、再デプロイ不要で興味を変更可能

---

## フェーズ6：ユーザーストーリー 3 - 設定可能なRSSソース管理（優先度: P2）

**目標**: RSSソースを外部設定から動的にロード、1つのソース失敗で他を継続処理

**独立テスト**: 1つのRSSソースをダウンさせ、他のソースから記事が取得されることを確認

### US3 実装

- [ ] T043 [US3] internal/rss/fetcher.goにエラー処理を追加（1つ失敗しても処理継続）
- [ ] T044 [US3] internal/config/schema.goにRSSSource.Enabledフィールドのサポートを追加
- [ ] T045 [US3] 無効なソースをスキップするロジックをcmd/curator/main.goに追加

**チェックポイント**: RSSソースの動的管理が機能し、障害耐性を確認

---

## フェーズ7：デプロイとスケジューリング

**目的**: インフラをデプロイし、毎日のスケジュール実行を設定

- [ ] T046 GCS bucketを作成し、terraform backendを初期化（terraform init）
- [ ] T047 Secret Managerにシークレットを作成（discord-webhook-url、gemini-api-key）
- [ ] T048 terraform applyで全インフラをデプロイ（terraform/environments/prod/）
- [ ] T049 Cloud Functions用にGoコードをビルドし、zipアーカイブを作成
- [ ] T050 Cloud Functionソースコードをデプロイ（gcloud functions deploy）
- [ ] T051 Cloud Schedulerジョブをテストトリガー（gcloud scheduler jobs run）
- [ ] T052 Discordで通知を確認し、エンドツーエンドフローを検証

---

## フェーズ8：監視とロギング

**目的**: 本番環境での観測可能性を確保

- [ ] T053 [P] Cloud Loggingで構造化ログを確認（エラー、警告、情報）
- [ ] T054 [P] メトリクスを追跡（Gemini APIリクエスト数、レート制限待機回数、処理時間）
- [ ] T055 [P] アラートポリシーを設定（Gemini 401エラー、429エラー、実行失敗）

---

## フェーズ9：磨きと横断的関心事

**目的**: 複数のユーザーストーリーに影響する改善

- [ ] T056 [P] README.mdを更新（デプロイ手順、Terraformコマンド、トラブルシューティング）
- [ ] T057 [P] quickstart.mdの手順を検証（新規ユーザーが従えるか確認）
- [ ] T058 コードクリーンアップとリファクタリング（重複コード削減、命名改善）
- [ ] T059 [P] internal/llm/evaluator_test.goにユニットテストを追加（モックGeminiクライアント）
- [ ] T060 [P] internal/discord/formatter_test.goにユニットテストを追加
- [ ] T061 エラーメッセージの日本語化（ログとDiscord通知）

---

## 必要な認証情報とリソース

### 事前準備が必要なもの

#### 1. Google Cloud Platform（GCP）アカウント
- **必要なタイミング**: フェーズ2（T006-T011 Terraformインフラ）、フェーズ7（T046-T052 デプロイ）
- **必要な操作**:
  - Googleアカウントでログイン
  - GCPプロジェクトを作成（Console UIまたは`gcloud projects create`）
  - 請求先アカウントを設定（無料枠でも必要）
  - `gcloud auth login`でローカルPCから認証
  - `gcloud auth application-default login`でTerraformの認証設定
- **必要な権限**:
  - Project Owner または Editor
  - 以下のAPIを有効化する必要あり:
    - Cloud Functions API
    - Cloud Scheduler API
    - Pub/Sub API
    - Firestore API
    - Secret Manager API
    - Cloud Build API

#### 2. Gemini API Key（Google AI Studio）
- **必要なタイミング**: フェーズ7（T047 Secret Manager設定）、フェーズ4（T033-T035 LLM評価の実装・テスト時）
- **取得方法**:
  1. [Google AI Studio](https://aistudio.google.com/)にアクセス
  2. Googleアカウントでログイン
  3. 「Get API Key」をクリック
  4. 新しいAPIキーを作成（プロジェクトを選択または新規作成）
  5. APIキーをコピー（後でSecret Managerに保存）
- **注意事項**:
  - 無料枠: 15 RPM（リクエスト/分）、1500 RPD（リクエスト/日）
  - APIキーは絶対にGitにコミットしない（.gitignoreで除外）
  - Secret Managerに保存することで安全に管理

#### 3. Discord Webhook URL
- **必要なタイミング**: フェーズ7（T047 Secret Manager設定）、フェーズ4（T036-T037 Discord通知の実装・テスト時）
- **取得方法**:
  1. Discordサーバーの設定を開く
  2. 「連携サービス」→「ウェブフック」
  3. 「新しいウェブフック」をクリック
  4. 名前を設定（例: RSS記事Bot）
  5. 通知先チャンネルを選択
  6. 「ウェブフックURLをコピー」
- **注意事項**:
  - Webhook URLは絶対にGitにコミットしない
  - Secret Managerに保存することで安全に管理

#### 4. GitHub Personal Access Token（オプション、設定の外部化用）
- **必要なタイミング**: フェーズ5（T040 GitHub Raw URLからの設定ロード）
- **取得方法**:
  1. GitHub Settings → Developer settings → Personal access tokens
  2. 「Generate new token」
  3. `repo`スコープを選択（プライベートリポジトリの場合）
  4. トークンをコピー
- **注意事項**:
  - パブリックリポジトリの場合は不要（Raw URLに直接アクセス可能）
  - プライベートリポジトリの場合のみ必要

### 認証情報の保存場所

| 認証情報 | 開発時（ローカル） | 本番時（Cloud Functions） |
|---------|-----------------|-------------------------|
| GCP認証 | `gcloud auth login`で自動 | サービスアカウントで自動 |
| Gemini API Key | `.env`ファイル（Gitignore対象） | Secret Manager: `gemini-api-key` |
| Discord Webhook URL | `.env`ファイル（Gitignore対象） | Secret Manager: `discord-webhook-url` |
| GitHub Token | `.env`ファイル（Gitignore対象） | Secret Manager: `github-token`（オプション） |

### .envファイルの例

```bash
# .env（Gitignore対象、ローカル開発用）
GCP_PROJECT_ID=your-gcp-project-id
GEMINI_API_KEY=your-gemini-api-key
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
CONFIG_URL=https://raw.githubusercontent.com/your-repo/main/config.json
# GITHUB_TOKEN=your-github-token  # プライベートリポジトリの場合のみ
```

### 認証フロー（タイムライン）

```
フェーズ1（セットアップ）
  ↓ Googleアカウントでログイン
フェーズ2（基盤 - Terraform）
  ↓ gcloud auth login（GCP認証）
  ↓ GCPプロジェクト作成
  ↓ 請求先アカウント設定
  ↓ terraform init（GCS backend用）
フェーズ4（US1実装・テスト）
  ↓ Gemini API Key取得（Google AI Studio）
  ↓ Discord Webhook URL取得
  ↓ .envファイルに保存（ローカルテスト用）
フェーズ7（デプロイ）
  ↓ Secret Managerにシークレット作成
  ↓ Gemini API Key、Discord Webhook URLを保存
  ↓ terraform apply（インフラデプロイ）
  ↓ Cloud Functionデプロイ
```

---

## 依存関係と実行順序

### フェーズ依存関係

- **セットアップ（フェーズ1）**: 依存関係なし - すぐに開始可能
- **基盤（フェーズ2）**: セットアップ完了に依存 - すべてのユーザーストーリーをブロック
- **US4（フェーズ3）**: 基盤完了に依存 - US1の前提条件
- **US1（フェーズ4）**: US4完了に依存 - MVP
- **US2、US3（フェーズ5、6）**: US1完了後に並列実行可能
- **デプロイ（フェーズ7）**: US1完了に依存（MVP）、US2/US3はオプション
- **監視（フェーズ8）**: デプロイ完了後
- **磨き（フェーズ9）**: すべてのユーザーストーリー完了後

### ユーザーストーリー依存関係

- **US4（P1）**: 基盤後に開始可能 - US1の前提条件（重複排除）
- **US1（P1）**: US4完了に依存 - MVP機能
- **US2（P2）**: US1完了後に開始可能 - US1と独立してテスト可能
- **US3（P2）**: US1完了後に開始可能 - US1と独立してテスト可能

### 並列機会

- [P]マークのすべてのセットアップタスクは並列実行可能
- [P]マークのすべての基盤タスクは並列実行可能（フェーズ2内）
- US1内の契約テスト（T026、T027、T028）は並列実行可能
- US1内のRSSフェッチ、記事抽出、Discord実装は並列実行可能
- US2とUS3はUS1完了後に並列実行可能
- 磨きフェーズのタスクは並列実行可能

---

## 実装戦略

### MVP優先（US4 + US1のみ）

1. フェーズ1完了：セットアップ
2. フェーズ2完了：基盤（Terraform、設定、Firestore、シークレット）
3. フェーズ3完了：US4（重複排除）
4. フェーズ4完了：US1（記事キュレーションと通知）
5. **停止して検証**: 手動トリガーでUS1をテスト
6. フェーズ7完了：デプロイとスケジューリング
7. **本番検証**: 毎日午前9時の自動実行を確認

### 段階的デリバリー

1. セットアップ + 基盤完了 → 基盤準備完了
2. US4追加 → 重複排除機能テスト
3. US1追加 → 独立してテスト → デプロイ（MVP！）
4. US2追加 → 設定の柔軟性テスト
5. US3追加 → ソース管理の柔軟性テスト
6. 監視 + 磨き → 本番運用準備完了

---

## 注記

- [P]タスク = 異なるファイル、依存関係なし
- [Story]ラベルはトレーサビリティのためタスクを特定のユーザーストーリーにマップ
- 各ユーザーストーリーは独立して完了およびテスト可能であるべき
- 契約テストは実装前に書き、失敗することを確認（TDD）
- 各タスクまたは論理的グループの後にコミット
- Terraformインフラは本番環境（prod）のみで、environments/ディレクトリ構造により将来の拡張に対応
- Firestoreの統合テストは省略（本番環境のみで検証）
- すべてのドキュメントとコメントは日本語で記述
