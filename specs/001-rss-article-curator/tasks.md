# タスク: RSS記事キュレーションBot

**入力**: `/specs/001-rss-article-curator/`からの設計ドキュメント
**前提条件**: plan.md、spec.md、data-model.md、contracts/

**整理**: タスクはユーザーストーリーごとにグループ化され、各ストーリーの独立した実装とテストを可能にします。

## Vibe Kanban での実行方法

このタスクリストはVibe Kanbanに移して、複数のClaude Codeインスタンスで並列実行することを想定しています。

### タスク表記の説明

- **[P]**: 並列実行可能（異なるファイル、依存関係なし）→ Kanbanで同時に複数の`In Progress`に移動可能
- **[BLOCK]**: ブロッキングタスク（これが完了するまで他のタスクを開始できない）→ 必ず順番に実行
- **[Story]**: このタスクが属するユーザーストーリー（例：US1、US4）
- **依存**: `→ T###` で依存するタスクを明記

### Kanbanボードの推奨構成

```
┌─────────────┬──────────────┬──────────────┬──────────────┐
│   Backlog   │ Ready to Do  │ In Progress  │     Done     │
├─────────────┼──────────────┼──────────────┼──────────────┤
│ 全タスク    │ 依存関係が   │ 現在実行中   │ 完了済み     │
│             │ 解決済みで   │ （最大3-5個  │              │
│             │ 実行可能     │  推奨）      │              │
└─────────────┴──────────────┴──────────────┴──────────────┘
```

### 並列実行の戦略

1. **フェーズごとに進める**: フェーズ1→2→3→4の順で進め、各フェーズ内では[P]タスクを並列実行
2. **[BLOCK]タスクを優先**: ブロッキングタスクが完了しないと次に進めないため最優先
3. **同じファイルを編集するタスクは避ける**: 競合を防ぐため、異なるファイルのタスクを選ぶ
4. **[P]マークを活用**: 同じフェーズ内の[P]タスクは同時に複数のClaude Codeで実行可能

### 実行順序の絶対ルール

```
フェーズ1（T001-T005）
  ↓ 【必須】T001完了後に他のタスク開始可能
フェーズ2（T006-T020）
  ↓ 【必須】フェーズ2全タスク完了後に次へ
フェーズ3（T021-T025）US4
  ↓ 【必須】US4完了後にUS1開始可能
フェーズ4（T026-T039）US1
  ↓ 【必須】US1完了でMVP達成
フェーズ5（T040-T042）US2 ┐
                         ├ 【並列可能】US1完了後に並列実行
フェーズ6（T043-T045）US3 ┘
  ↓
フェーズ7（T046-T052）デプロイ
  ↓
フェーズ8（T053-T055）監視
  ↓
フェーズ9（T056-T061）磨き
```

---

## フェーズ1：セットアップ（共有インフラ）

**目的**: プロジェクトの初期化と基本構造

**並列実行**: T001完了後、T002-T005は並列実行可能

- [ ] T001 [BLOCK] plan.mdに従ってプロジェクト構造を作成（cmd/、internal/、tests/、terraform/）
- [ ] T002 Goモジュールを初期化（go mod init、go.mod、go.sum）→ T001
- [ ] T003 [P] .gitignoreファイルを作成（terraform.tfvars、.env、シークレットを除外）→ T001
- [ ] T004 [P] config.jsonテンプレートを作成（RSSソース、興味トピック、通知設定）→ T001
- [ ] T005 [P] README.mdを作成（プロジェクト概要、セットアップ手順）→ T001

**並列実行例**:
- Claude Code #1: T002実行中
- Claude Code #2: T003実行中（並列OK）
- Claude Code #3: T004実行中（並列OK）

---

## フェーズ2：基盤（ブロッキング前提条件）

**目的**: いずれかのユーザーストーリーを実装する前に完了していなければならないコアインフラ

**⚠️ 重要**: このフェーズが完了するまでユーザーストーリー作業は開始できません

**並列実行**: T006-T020は最大4-5個並列実行可能（異なるディレクトリ）

### Terraformインフラ（IaC）

- [ ] T006 [P] terraform/modules/firestore/を作成（Firestore Database + インデックス定義）→ T001
- [ ] T007 [P] terraform/modules/secrets/を作成（Secret Manager定義）→ T001
- [ ] T008 [P] terraform/modules/scheduler/を作成（Cloud Scheduler + Pub/Sub Topic定義）→ T001
- [ ] T009 [P] terraform/modules/cloud-function/を作成（Cloud Function Gen 2定義）→ T001
- [ ] T010 [BLOCK] terraform/environments/prod/main.tfを作成（モジュールを組み合わせ）→ T006, T007, T008, T009
- [ ] T011 [P] terraform/environments/prod/variables.tf、outputs.tf、backend.tfを作成 → T010

### 設定管理

- [ ] T012 [P] internal/config/schema.goを実装（RSSSource、InterestTopic、NotificationSettings構造体）→ T001
- [ ] T013 internal/config/loader.goを実装（GitHubまたはローカルファイルからconfig.jsonをロード）→ T012
- [ ] T014 [P] internal/config/validator.goを実装（設定検証ロジック）→ T012

### Secret Manager統合

- [ ] T015 [P] internal/secrets/manager.goを実装（Discord Webhook URL、Gemini APIキーの取得）→ T001

### Firestoreクライアント

- [ ] T016 [P] internal/storage/firestore.goを実装（Firestoreクライアントラッパー）→ T001
- [ ] T017 [P] internal/storage/notified.goを実装（notified_articlesコレクション操作）→ T016
- [ ] T018 [P] internal/storage/rejected.goを実装（rejected_articlesコレクション操作）→ T016

### エラー処理とログ

- [ ] T019 [P] internal/errors/errors.goを実装（カスタムエラー型とラッピング）→ T001
- [ ] T020 [P] internal/logging/logger.goを実装（構造化ログ設定）→ T001

**並列実行例（フェーズ2）**:
- Claude Code #1: T006（Firestore module）実行中
- Claude Code #2: T007（Secrets module）実行中（並列OK）
- Claude Code #3: T012（config schema）実行中（並列OK）
- Claude Code #4: T015（secrets manager）実行中（並列OK）
- Claude Code #5: T019（errors）実行中（並列OK）

**チェックポイント**: 基盤準備完了（T006-T020すべて完了）- ユーザーストーリー実装を開始可能

---

## フェーズ3：ユーザーストーリー 4 - 記事の重複排除（優先度: P1）🎯 MVP前提

**目標**: 通知済み記事と却下済み記事をFirestoreで追跡し、重複評価を防止

**独立テスト**: 記事を一度保存し、次回チェック時に「既存」として検出されることを確認

**並列実行**: T021-T025は並列実行可能（異なる関数を実装）

### US4 契約テスト

- [ ] T021 [P] [US4] tests/contract/firestore_test.goを作成（Firestoreエミュレータでの契約テスト）→ T016-T018

### US4 実装

- [ ] T022 [P] [US4] internal/storage/notified.goに SaveNotifiedArticle() 実装 → T017
- [ ] T023 [P] [US4] internal/storage/notified.goに IsArticleNotified() 実装 → T017
- [ ] T024 [P] [US4] internal/storage/rejected.goに SaveRejectedArticle() 実装 → T018
- [ ] T025 [P] [US4] internal/storage/rejected.goに IsArticleRejected() 実装 → T018

**並列実行例（フェーズ3）**:
- Claude Code #1: T021（契約テスト）実行中
- Claude Code #2: T022-T023（notified.go）実行中（並列OK）
- Claude Code #3: T024-T025（rejected.go）実行中（並列OK）

**チェックポイント**: Firestore重複排除が機能し、独立してテスト可能（T021-T025すべて完了）

---

## フェーズ4：ユーザーストーリー 1 - 毎日のキュレーション記事通知（優先度: P1）🎯 MVP

**目標**: RSSフィードを取得し、Geminiで評価し、Discordに通知

**独立テスト**: 手動でCloud Functionをトリガーし、Discordに3〜5件の記事が通知されることを確認

**並列実行**: 契約テスト（T026-T028）、各実装グループ（RSS、記事、LLM、Discord）は並列実行可能

### US1 契約テスト

- [ ] T026 [P] [US1] tests/contract/rss_parser_test.goを作成（RSS 2.0/Atomフィードフィクスチャでテスト）→ T021-T025
- [ ] T027 [P] [US1] tests/contract/gemini_api_test.goを作成（Gemini APIリクエスト/レスポンス契約）→ T021-T025
- [ ] T028 [P] [US1] tests/contract/discord_webhook_test.goを作成（Discord Embeds APIペイロード契約）→ T021-T025

### US1 RSSフェッチ・パース

- [ ] T029 [P] [US1] internal/rss/fetcher.goを実装（RSSフィードURLからHTMLを取得）→ T026
- [ ] T030 [P] [US1] internal/rss/parser.goを実装（RSS 2.0/Atom XMLをArticle構造体にパース）→ T026

### US1 記事コンテンツ抽出

- [ ] T031 [P] [US1] internal/article/fetcher.goを実装（記事URLからHTMLを取得）→ T026
- [ ] T032 [P] [US1] internal/article/extractor.goを実装（go-readabilityで記事テキストを抽出）→ T031

### US1 LLM評価

- [ ] T033 [P] [US1] internal/llm/client.goを実装（レート制限付きGemini APIクライアント、golang.org/x/time/rate使用）→ T027
- [ ] T034 [US1] internal/llm/evaluator.goを実装（記事の関連性を評価、AI生成記事判定、スコア0〜100）→ T033
- [ ] T035 [P] [US1] internal/llm/summarizer.goを実装（記事要約を生成、50〜200文字）→ T033

### US1 Discord通知

- [ ] T036 [P] [US1] internal/discord/client.goを実装（Discord Webhook APIクライアント）→ T028
- [ ] T037 [P] [US1] internal/discord/formatter.goを実装（記事をEmbedsペイロードとしてフォーマット）→ T028

### US1 Cloud Functionエントリーポイント

- [ ] T038 [BLOCK] [US1] cmd/curator/main.goを実装（Pub/SubトリガーCloud Functionエントリーポイント）→ T029-T037
- [ ] T039 [BLOCK] [US1] メインオーケストレーションロジックを実装 → T038
  - config.jsonをロード
  - RSSフィードを取得・パース
  - 重複チェック（Firestore）
  - 記事コンテンツ抽出
  - Geminiで評価
  - 上位3〜5件を選択
  - Discordに通知
  - Firestoreに保存

**並列実行例（フェーズ4）**:
- Claude Code #1: T029-T030（RSS）実行中
- Claude Code #2: T031-T032（article）実行中（並列OK）
- Claude Code #3: T033-T035（LLM）実行中（並列OK）
- Claude Code #4: T036-T037（Discord）実行中（並列OK）

**チェックポイント**: この時点でUS1は完全に機能し、手動トリガーで独立してテスト可能（T026-T039すべて完了）

---

## フェーズ5：ユーザーストーリー 2 - 設定可能な興味管理（優先度: P2）

**目標**: 興味トピックを外部設定（GitHub）から動的にロード

**独立テスト**: config.jsonの興味を更新し、次回実行で新しい興味が反映されることを確認

**並列実行**: T040-T042は並列実行可能（異なるファイルを編集）

### US2 実装

- [ ] T040 [P] [US2] internal/config/loader.goにGitHub Raw URLからの設定ロード機能を追加 → T039
- [ ] T041 [P] [US2] 優先度重み付けロジックを実装（high: 2.0x, medium: 1.0x, low: 0.5x）→ T039
- [ ] T042 [P] [US2] エイリアス対応をinternal/llm/evaluator.goに追加 → T039

**並列実行例（フェーズ5）**:
- Claude Code #1: T040（config loader）実行中
- Claude Code #2: T041（優先度重み付け）実行中（並列OK）
- Claude Code #3: T042（エイリアス対応）実行中（並列OK）

**チェックポイント**: 設定の動的ロードが機能し、再デプロイ不要で興味を変更可能（T040-T042すべて完了）

---

## フェーズ6：ユーザーストーリー 3 - 設定可能なRSSソース管理（優先度: P2）

**目標**: RSSソースを外部設定から動的にロード、1つのソース失敗で他を継続処理

**独立テスト**: 1つのRSSソースをダウンさせ、他のソースから記事が取得されることを確認

**並列実行**: T043-T045は並列実行可能（異なるファイルを編集）

### US3 実装

- [ ] T043 [P] [US3] internal/rss/fetcher.goにエラー処理を追加（1つ失敗しても処理継続）→ T039
- [ ] T044 [P] [US3] internal/config/schema.goにRSSSource.Enabledフィールドのサポートを追加 → T039
- [ ] T045 [P] [US3] 無効なソースをスキップするロジックをcmd/curator/main.goに追加 → T039

**並列実行例（フェーズ6）**:
- Claude Code #1: T043（rss fetcher）実行中
- Claude Code #2: T044（config schema）実行中（並列OK）
- Claude Code #3: T045（main.go）実行中（並列OK）

**チェックポイント**: RSSソースの動的管理が機能し、障害耐性を確認（T043-T045すべて完了）

---

## フェーズ7：デプロイとスケジューリング

**目的**: インフラをデプロイし、毎日のスケジュール実行を設定

**⚠️ 重要**: このフェーズは順次実行が必要（Terraformの依存関係）

**並列実行**: なし（順次実行のみ）

- [ ] T046 [BLOCK] GCS bucketを作成し、terraform backendを初期化（terraform init）→ T040-T045
- [ ] T047 [BLOCK] Secret Managerにシークレットを作成（discord-webhook-url、gemini-api-key）→ T046
- [ ] T048 [BLOCK] terraform applyで全インフラをデプロイ（terraform/environments/prod/）→ T047
- [ ] T049 [BLOCK] Cloud Functions用にGoコードをビルドし、zipアーカイブを作成 → T048
- [ ] T050 [BLOCK] Cloud Functionソースコードをデプロイ（gcloud functions deploy）→ T049
- [ ] T051 [BLOCK] Cloud Schedulerジョブをテストトリガー（gcloud scheduler jobs run）→ T050
- [ ] T052 [BLOCK] Discordで通知を確認し、エンドツーエンドフローを検証 → T051

**チェックポイント**: 本番環境デプロイ完了、毎日午前9時の自動実行が設定済み（T046-T052すべて完了）

---

## フェーズ8：監視とロギング

**目的**: 本番環境での観測可能性を確保

**並列実行**: T053-T055は並列実行可能（独立した設定タスク）

- [ ] T053 [P] Cloud Loggingで構造化ログを確認（エラー、警告、情報）→ T052
- [ ] T054 [P] メトリクスを追跡（Gemini APIリクエスト数、レート制限待機回数、処理時間）→ T052
- [ ] T055 [P] アラートポリシーを設定（Gemini 401エラー、429エラー、実行失敗）→ T052

**並列実行例（フェーズ8）**:
- Claude Code #1: T053（ログ確認）実行中
- Claude Code #2: T054（メトリクス）実行中（並列OK）
- Claude Code #3: T055（アラート）実行中（並列OK）

---

## フェーズ9：磨きと横断的関心事

**目的**: 複数のユーザーストーリーに影響する改善

**並列実行**: T056-T061は並列実行可能（異なるファイル、独立したタスク）

- [ ] T056 [P] README.mdを更新（デプロイ手順、Terraformコマンド、トラブルシューティング）→ T053-T055
- [ ] T057 [P] quickstart.mdの手順を検証（新規ユーザーが従えるか確認）→ T053-T055
- [ ] T058 [P] コードクリーンアップとリファクタリング（重複コード削減、命名改善）→ T053-T055
- [ ] T059 [P] internal/llm/evaluator_test.goにユニットテストを追加（モックGeminiクライアント）→ T053-T055
- [ ] T060 [P] internal/discord/formatter_test.goにユニットテストを追加 → T053-T055
- [ ] T061 [P] エラーメッセージの日本語化（ログとDiscord通知）→ T053-T055

**並列実行例（フェーズ9）**:
- Claude Code #1: T056（README更新）実行中
- Claude Code #2: T057（quickstart検証）実行中（並列OK）
- Claude Code #3: T059（ユニットテスト）実行中（並列OK）
- Claude Code #4: T060（ユニットテスト）実行中（並列OK）
- Claude Code #5: T061（日本語化）実行中（並列OK）

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
