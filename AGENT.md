# エージェント指示書

このドキュメントは、このプロジェクトで作業するAIエージェント（Claude Code）向けの指示を提供します。

## プロジェクト概要

### プロジェクト名
RSS記事キュレーションBot（discord-article-bot）

### 目的

**中核的な価値提案**: 自動化されたパーソナライズされた記事発見

特定の技術に興味がある開発者として、複数のRSSまとめサイトを手動で閲覧することなく、最新情報を把握できるように、毎日3〜5件の関連技術記事のダイジェストをDiscordで受け取ることができます。

技術ブログまとめサイトを毎日監視し、Gemini LLMを使用してユーザー定義の興味に対する記事の関連性を評価し、要約付きの3〜5件のキュレーション記事をDiscordに投稿するサーバーレスRSS記事キュレーター。

### 主要機能

1. **毎日のキュレーション記事通知** (P1)
   - 毎日午前9時（JST）に自動実行
   - 興味に一致する記事を3〜5件選択
   - 各記事の要約を生成
   - 最も価値のある記事を強調する全体的な推薦を提供

2. **設定可能な興味管理** (P2)
   - config.jsonで興味トピックを管理
   - ボットを再デプロイせずに興味を更新可能
   - GitHub上の設定ファイルから動的に読み込み

3. **設定可能なRSSソース管理** (P2)
   - 複数のRSSフィードまとめサイトをサポート
   - コード変更なしでソースを追加・削除可能
   - 1つのソース失敗時も他のソースの処理を継続

4. **記事の重複排除** (P1)
   - 通知済み記事をFirestoreで追跡（30日間TTL）
   - 興味がないと評価された記事も追跡
   - 重複通知を防止し、新鮮なコンテンツのみを提供

### 技術スタック

- **言語**: Go 1.21+
- **プラットフォーム**: Google Cloud Functions Gen 2
- **インフラ**: Terraform
- **ストレージ**: Firestore
- **LLM**: Google Gemini Flash API
- **通知**: Discord Webhook

### アーキテクチャ設計思想

**イベント駆動アーキテクチャ**:
- Cloud Scheduler → Pub/Sub → Cloud Functions
- 非同期処理によるスケーラビリティ
- サーバーレスによるインフラ管理の簡素化

**セキュリティ**:
- Secret Managerでシークレット管理（Discord Webhook URL、Gemini APIキー）
- Firestoreセキュリティルール（バックエンドのみアクセス）
- コードや環境変数にシークレットを含めない

**信頼性**:
- RSSフェッチ失敗時は他のソース処理を継続
- Gemini APIレート制限対応（15 RPM、1500 RPD）
- Discord Webhook失敗時は指数バックオフで再試行
- エラーハンドリングとロギングの徹底

**コスト効率**:
- Gemini Flash API無料枠を活用（1日1500リクエスト未満）
- Firestore無料枠内で運用（1日5万読み取り、2万書き込み）
- Cloud Functions無料枠（月200万呼び出し）
- 目標: 1日あたり5ドル未満

**パフォーマンス目標**:
- 100〜200件の記事を1時間以内に処理
- Firestoreの重複チェックを2秒以内に完了
- レート制限を尊重しつつ効率的に評価

## プロジェクト構造

### 現在の構造

```
.
├── AGENT.md             # エージェント指示書（本ドキュメント）
├── README.md            # プロジェクト概要
├── claude.md            # Claude設定（言語、ブランチ戦略、コミット規則）
├── config.json          # RSS設定ファイル
├── go.mod               # Go依存関係管理
├── go.sum               # Go依存関係チェックサム
├── cloudbuild.yaml      # Cloud Build設定
├── cmd/                 # エントリーポイント（✅ T007, T008完了）
│   ├── curator/         # Cloud Functions本番環境用
│   └── local-test/      # ローカルテスト用
├── internal/            # 内部パッケージ（✅ すべて実装済み）
│   ├── config/          # 設定管理（✅ T002完了）
│   ├── secrets/         # Secret Manager統合（✅ T002完了）
│   ├── errors/          # エラーハンドリング（✅ T002完了）
│   ├── logging/         # ロギング（✅ T002完了）
│   ├── storage/         # Firestore操作（✅ T003完了）
│   ├── rss/             # RSSフィード処理（✅ T004完了）
│   ├── article/         # 記事コンテンツ抽出（✅ T004完了）
│   ├── llm/             # Gemini API統合（✅ T005完了、サマリー機能追加済み）
│   └── discord/         # Discord通知（✅ T006完了）
├── tests/               # テストファイル（契約テスト実装済み）
│   └── contract/        # 契約テスト（Discord, Firestore, Gemini, RSS）
├── terraform/           # インフラストラクチャコード（✅ T001完了）
│   ├── README.md
│   ├── environments/
│   │   └── prod/
│   └── modules/
│       ├── firestore/
│       ├── secrets/
│       ├── scheduler/
│       └── cloud-function/
└── specs/               # 設計ドキュメント
    └── 001-rss-article-curator/
        ├── spec.md          # 機能仕様書
        ├── plan.md          # 実装計画
        ├── tasks.md         # タスクリスト
        ├── data-model.md    # データモデル
        ├── quickstart.md    # クイックスタート
        ├── research.md      # 調査ノート
        ├── contracts/       # 契約定義
        └── checklists/      # チェックリスト
```

### 最近追加された機能

以下の機能が最近追加されました:

1. **記事全体サマリー機能** (2025-11-10追加)
   - `internal/llm/summarizer.go` - 複数記事の全体サマリー生成
   - Gemini API v2.0を使用
   - 記事全体の傾向分析と推薦機能

2. **Gemini API v2.0対応** (2025-11-10)
   - 無料版の制限に対応
   - レート制限の調整

3. **ローカルテスト環境** (2025-11-10)
   - `cmd/local-test/main.go` - Firestoreエミュレータ対応
   - API制限考慮（3記事制限）

## 重要なドキュメント

作業を開始する前に、以下のドキュメントを必ず確認してください。

### 必読ドキュメント

1. **[claude.md](./claude.md)** - Claude設定、言語設定、ブランチ戦略、コミットメッセージ規則
2. **[specs/001-rss-article-curator/spec.md](./specs/001-rss-article-curator/spec.md)** - 機能仕様書
3. **[specs/001-rss-article-curator/plan.md](./specs/001-rss-article-curator/plan.md)** - 実装計画
4. **[specs/001-rss-article-curator/tasks.md](./specs/001-rss-article-curator/tasks.md)** - タスクリストと開発ワークフロー
5. **[specs/001-rss-article-curator/data-model.md](./specs/001-rss-article-curator/data-model.md)** - データモデル定義

### 参照ドキュメント

- **[specs/001-rss-article-curator/quickstart.md](./specs/001-rss-article-curator/quickstart.md)** - デプロイ手順
- **[specs/001-rss-article-curator/contracts/](./specs/001-rss-article-curator/contracts/)** - 外部API契約定義
- **[terraform/README.md](./terraform/README.md)** - Terraformインフラ構成

## 開発ワークフロー

### 言語設定

**すべての応答、ドキュメント、コメントは日本語で記述してください。**

### ブランチ戦略

#### ブランチ命名規則
- **フォーマット**: `{タイプ}/t{番号}-{簡潔な説明}`
- **例**:
  - `feat/t002-config-utils`
  - `terraform/t001-infrastructure`
  - `fix/t034-llm-evaluator-bug`

#### タイプ
- `feat`: 新機能の実装
- `fix`: バグ修正
- `terraform`: インフラストラクチャコード
- `test`: テストコードの追加・修正
- `docs`: ドキュメントのみの変更
- `refactor`: リファクタリング

### コミットメッセージ規則

#### フォーマット
```
{タイプ}: {簡潔な説明}
```

#### 重要な注意事項
- **1〜2行でわかる簡潔でまとまった文章にすること**
- **箇条書きは使用しない**
- 現在形で記述（「追加する」ではなく「追加」）
- Vibe KanbanのタスクIDを記載（例: T001, T029）

#### 例
```
feat: config、secrets、errors、loggingパッケージを追加し、設定ファイルの読み込みと検証、Secret Manager統合を実装
```

```
terraform: Firestore、Secret Manager、Cloud Scheduler、Cloud FunctionsのTerraformモジュールと本番環境設定を追加
```

### プルリクエストワークフロー

#### PR作成

1. タスク完了後、`gh pr create`コマンドでPRを作成
2. 以下のテンプレートを使用:

```markdown
## 関連Issue

<!-- 関連IssueのURLまたは番号を記載 -->

Closes #

## やったこと

<!-- このプルリクで何をしたのか？ -->

## やらないこと

<!-- このプルリクでやらないこと（あれば。ない場合は「無し」でOK） -->

## その他

<!-- レビュワーへの補足や懸念点や重点的に見て欲しい箇所 -->
```

#### PRレビュー

PRを作成したら、**Vibe Kanbanにレビュー用タスクを自動作成**し、そのタスクを即座に開始します。

**レビュータスクの作成**:

1. **タスクタイトル**: `T{番号}-review: {元のタスク名}のレビュー`
2. **タスク説明**: 元のタスクの説明全体 + レビュー指示を追加
3. **タスク作成後、即座に開始**: `mcp__vibe_kanban__start_task_attempt`を使用
   - executor: `CLAUDE_CODE`
   - base_branch: `{実装したブランチ名}`（例: `feat/t003-firestore-deduplication`）
   - **重要**: レビューは実装したブランチをチェックアウトして行うため、親タスクで作成したブランチ名を指定すること

**レビュー方針**:
- 辛口にレビューをする（厳しめに、本来あるべき姿を指摘）
- マサカリ大歓迎
- ベースの実装に改善の余地がある場合も指摘

**レビュー観点**:
- 変更の結果冗長ではないか
- デバッグなどの無駄なコードが残留していないか
- docs配下に今回のコードの変更点が盛り込まれているか（ドキュメントとコードに差分がないか、追加で文章で説明すべき設計がないか）
- コードがコーディング規約に従っているか
- テストが十分に実装されているか
- エラーハンドリングが適切か

**レビュー実施手順**:
1. PRの差分を確認（`gh pr diff`または`git diff main...{ブランチ名}`）
2. 変更されたファイルを読み取り、実装内容を確認
3. 上記の観点に基づいてレビューコメントを作成
4. 観点ごとに結果をつけ終わったらユーザーにまず通知
5. **変更は少し待つ - ユーザーの指示を待つこと**

**その他**:
- 想定外の作業が発生したり、エラーが見つかった場合はユーザに相談
- PRにClaude Codeの署名は不要

詳細は[tasks.md](./specs/001-rss-article-curator/tasks.md)の「PRレビューフロー」セクションを参照してください。

#### マージ

**重要**: ユーザーの承認を得るまで絶対にマージしないこと

## タスク管理

### タスクリスト

[specs/001-rss-article-curator/tasks.md](./specs/001-rss-article-curator/tasks.md)にすべてのタスクが定義されています。

**現在の進捗**: 2/9タスク完了

- [x] T001: Terraformインフラ構築 ✅
- [x] T002: 設定管理とユーティリティ ✅
- [ ] T003: Firestore重複排除機能 🔄 次のタスク
- [ ] T004: RSS記事取得機能 🔄 次のタスク
- [ ] T005: Gemini評価機能 🔄 次のタスク
- [ ] T006: Discord通知機能 🔄 次のタスク
- [ ] T007: メインオーケストレーション
- [ ] T008: ローカルテストとデバッグ
- [ ] T009: GCPデプロイとCI/CD

### タスク実行順序

```
T001: Terraformインフラ構築 ✅ 完了
  ↓
T002: 設定管理とユーティリティ ✅ 完了
  ↓
T003: Firestore重複排除機能（並列実行可能） 🔄 次のタスク
  ↓
T004: RSS記事取得機能（並列実行可能） 🔄 次のタスク
T005: Gemini評価機能（並列実行可能） 🔄 次のタスク
T006: Discord通知機能（並列実行可能） 🔄 次のタスク
  ↓
T007: メインオーケストレーション
  ↓
T008: ローカルテストとデバッグ
  ↓
T009: GCPデプロイとCI/CD
```

**次のステップ**: T003、T004、T005、T006は並列実行可能です。

### Vibe Kanbanでのタスク管理

このプロジェクトではVibe Kanbanを使用してタスクを管理します。

**Vibe Kanbanアクセス方法**:
- WebUI: http://127.0.0.1:3000
- MCPサーバー: `claude mcp list`で接続状態を確認
- MCPサーバー状態: ✓ Connected

**現在のタスク状態**:
- 完了: T001 (Terraformインフラ構築), T002 (設定管理とユーティリティ)
- 次のタスク: T003, T004, T005, T006 (並列実行可能)

**タスク作成時の重要な指示**:

Vibe Kanbanで新しいタスクを作成する際は、**必ずこのAGENT.mdの内容を完全に読み取り、タスク説明の中に全文をコピーして埋め込んでください**。

参照リンクではなく、以下の内容を完全にタスクに埋め込むこと:
- プロジェクト概要（目的、技術スタック、主要機能、アーキテクチャ設計思想）
- プロジェクト構造（現在の構造と計画中の構造）
- 開発ワークフロー（言語設定、ブランチ戦略、コミットメッセージ規則、PRワークフロー）
- コーディング規約（Go、テスト、Terraform）
- 設定ファイルの説明
- タスク間の依存関係

これにより、新しいClaude Codeインスタンスがタスクを受け取った時点で、プロジェクト全体のコンテキストを完全に理解できます。

詳細は[tasks.md](./specs/001-rss-article-curator/tasks.md)の「Vibe Kanban での実行方法」セクションを参照してください。

### タスク完了条件

各タスクには以下の共通完了条件があります:

1. 機能要件の実装完了
2. テストの実装と成功
3. `gh pr create`でPRを作成
4. 別のClaude Codeインスタンスでレビューを実施
5. ユーザーの承認を得てからマージ

## コーディング規約

### Go言語

- Go 1.21+の標準ライブラリとイディオムに従う
- `gofmt`でコードをフォーマット
- コメントは日本語で記述
- エラーハンドリングを適切に実装
- 構造化ログを使用（`internal/logging`パッケージ）

### テスト

- すべての新機能にテストを含める
- 契約テストで外部API統合をテスト
- テスト名は日本語で記述可能
- テストカバレッジを意識する

### Terraform

- モジュール化を推進
- 変数と出力を明確に定義
- `terraform fmt`でフォーマット
- 各モジュールにREADME.mdを含める

## 設定ファイル

### config.json

RSSソース、興味トピック、通知設定を定義します。

```json
{
  "rss_sources": [
    {
      "name": "dev.to",
      "url": "https://dev.to/feed",
      "enabled": true
    }
  ],
  "interests": [
    "Go言語",
    "クラウドアーキテクチャ",
    "サーバーレス"
  ],
  "article_count": 5,
  "score_threshold": 70,
  "schedule": "0 9 * * *"
}
```

### 環境変数

ローカル開発時は`.env`ファイルを使用:

```bash
GCP_PROJECT_ID=your-gcp-project-id
GEMINI_API_KEY=your-gemini-api-key
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
CONFIG_URL=https://raw.githubusercontent.com/your-repo/main/config.json
```

本番環境ではSecret Managerを使用。

## トラブルシューティング

### よくある問題

1. **Terraform実行エラー**
   - GCP認証を確認: `gcloud auth login`
   - プロジェクトIDを確認: `gcloud config get-value project`

2. **ローカルテスト失敗**
   - Firestoreエミュレータが起動しているか確認
   - 環境変数が正しく設定されているか確認

3. **PR作成エラー**
   - `gh`コマンドがインストールされているか確認
   - GitHub認証を確認: `gh auth status`

### サポート

想定外の問題が発生した場合は、必ずユーザーに相談してください。

## 追加リソース

- [Google Cloud Functions ドキュメント](https://cloud.google.com/functions/docs)
- [Terraform GCP プロバイダー](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [Gemini API ドキュメント](https://ai.google.dev/docs)
- [Discord Webhook ガイド](https://discord.com/developers/docs/resources/webhook)

## 現在の進捗状況

**実装タスク**: 8/9タスク完了

- [x] T001: infrastructure (Terraform構築) ✅
- [x] T002: config-utils (設定管理とユーティリティ) ✅
- [x] T003: firestore-deduplication (Firestore重複排除) ✅
- [x] T004: rss-article-fetcher (RSS記事取得) ✅
- [x] T005: gemini-integration (Gemini API統合) ✅
- [x] T006: discord-notification (Discord通知) ✅
- [x] T007: main-orchestration (メインオーケストレーション) ✅
- [x] T008: local-testing (ローカルテスト) ✅
- [ ] T009: gcp-deployment (GCPデプロイとCI/CD) 🔄 **次のタスク**

**次のステップ**: T009 (GCPデプロイとCI/CD) のみ残っています。すべての実装とローカルテストは完了済みです。

詳細は[tasks.md](./specs/001-rss-article-curator/tasks.md)を参照してください。

## 変更履歴

### 2025-11-10
- 記事全体サマリー機能を追加 (`internal/llm/summarizer.go`)
- Gemini API v2.0に更新（無料版制限対応）
- ローカルテスト環境を整備 (`cmd/local-test/`)
- T003〜T008の実装完了
- AGENT.mdとtasks.mdの進捗状況を更新

### 初版
- プロジェクト構造とワークフロー定義
- T001, T002の実装完了

---

**最終更新**: 2025-11-06
