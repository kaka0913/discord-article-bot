# タスク: RSS記事キュレーションBot

**入力**: `/specs/001-rss-article-curator/`からの設計ドキュメント
**前提条件**: plan.md、spec.md、data-model.md、contracts/

**整理**: タスクはMVP達成に必要な機能単位でグループ化され、並列実行可能なまとまりになっています。

## 開発ワークフロー

### ブランチ戦略とコミットメッセージ

**重要**: タスクを開始する前に[claude.md](../../claude.md)のブランチ戦略とコミットメッセージ規則を確認してください。

**ブランチ命名規則**: `{タイプ}/t{番号}-{簡潔な説明}`
- 例: `terraform/t001-infrastructure`
- 例: `feat/t002-config-utils`
- 例: `feat/t005-gemini-integration`

**コミットメッセージ形式**:
```
{タイプ}: {簡潔な説明}
```

**注意**: コミットメッセージは1〜2行でわかる簡潔でまとまった文章にすること。箇条書きは使用しない。

詳細は[claude.md](../../claude.md)を参照してください。

### プルリクエスト作成とレビュー

**PRテンプレート**:
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

**PRレビューフロー**:
1. タスク完了後、`gh pr create`コマンドでPRを作成（上記テンプレートを使用）
2. 別のClaude Codeインスタンスを立ち上げてレビューを実施
3. レビューコメントは1〜2行程度、日本語で記載
4. **重要**: ユーザーの承認を得るまで絶対にマージしないこと

**レビュー実施時の指示**:

現在のブランチに対応するPull Requestをレビューしてください。

辛口にレビューをしてください。厳しめに、本来あるべき姿や、マサカリ大歓迎です。
あとはそもそものベースの実装が改善の余地がある場合も指摘して欲しいです。

レビュー時には基本的な観点に加えて、以下の観点も入れてください。

- 変更の結果冗長ではないか
- デバッグなどの無駄なコードが残留していないか
- docs配下に今回のコードの変更点が盛り込まれているか（ドキュメントとコードに差分がないか、追加で文章で説明すべき設計がないか）

観点ごとに結果をつけ終わったらユーザーにまず通知して、変更は少し待ってください。

**その他**:
- 想定外の作業が発生したり、エラーが見つかった場合はユーザに相談してください
- PRの内容にClaude Codeによるものであることがわかる署名は不要です

## Vibe Kanban での実行方法

### タスク作成時の重要な指示

**Vibe Kanbanでタスクを作成する際は、必ず以下の手順に従ってください:**

1. **AGENT.mdの内容を完全に読み取る**
   - [AGENT.md](../../AGENT.md)の内容を完全に読み取る
   - 内容を参照リンクではなく、タスク説明の中に完全にコピーして埋め込む

2. **タスク説明に埋め込む内容**
   - プロジェクト概要（目的、技術スタック）の全文
   - 開発ワークフロー（ブランチ戦略、コミットメッセージ規則、PRワークフロー）の全文
   - コーディング規約の全文
   - 現在のプロジェクト構造と実装状況
   - 計画中の構造と各タスクの実装予定
   - タスク間の依存関係

3. **タスク説明の形式**
   ```
   ## タスク: T{番号} - {タスク名}

   ---
   ## プロジェクト情報（AGENT.mdより）

   ### プロジェクト名
   RSS記事キュレーションBot（discord-article-bot）

   ### 目的
   {AGENT.mdの「目的」セクションの全文をコピー}

   ### 技術スタック
   {AGENT.mdの「技術スタック」セクションの全文をコピー}

   ### プロジェクト構造
   {AGENT.mdの「プロジェクト構造」セクションの全文をコピー（現在の構造と計画中の構造の両方）}

   ---
   ## 開発ワークフロー（AGENT.mdより）

   ### 言語設定
   {AGENT.mdの「言語設定」セクションの全文をコピー}

   ### ブランチ戦略
   {AGENT.mdの「ブランチ戦略」セクションの全文をコピー}

   ### コミットメッセージ規則
   {AGENT.mdの「コミットメッセージ規則」セクションの全文をコピー}

   ### プルリクエストワークフロー
   {AGENT.mdの「プルリクエストワークフロー」セクションの全文をコピー（PR作成、レビュー、マージの全て）}

   ---
   ## コーディング規約（AGENT.mdより）

   ### Go言語
   {AGENT.mdの「Go言語」セクションの全文をコピー}

   ### テスト
   {AGENT.mdの「テスト」セクションの全文をコピー}

   ### Terraform
   {AGENT.mdの「Terraform」セクションの全文をコピー}

   ---
   ## このタスクの詳細

   ### 目的
   {このタスクで実現すること}

   ### 成果物
   {作成するファイルやパッケージ}

   ### タスク内容
   {具体的な実装ステップ}

   ### ブランチ
   {タイプ}/t{番号}-{説明}

   ### 完了条件
   - [ ] 機能要件の実装完了
   - [ ] テストの実装と成功
   - [ ] ghコマンドでPRを作成（上記PRテンプレートを使用）
   - [ ] 別のClaude Codeインスタンスでレビュー実施（上記レビュー方針に従う）
   - [ ] ユーザーの承認を得てからマージ

   ### 依存関係
   {依存するタスクの番号と名前}

   ### 参照ドキュメント
   - [spec.md](./spec.md)
   - [plan.md](./plan.md)
   - [data-model.md](./data-model.md)
   - その他関連ドキュメント
   ```

4. **重要な注意事項**
   - **AGENT.mdの内容は参照リンクではなく、必ず全文をタスク説明にコピーして埋め込むこと**
   - これにより、新しいClaude Codeインスタンスがタスクを受け取った時点で、プロジェクト全体のコンテキストを完全に理解できる
   - ブランチ命名規則、コミットメッセージ規則、PRテンプレート、レビュー方針の全てを含める
   - 現在の実装状況（完了済みタスク）と計画中の構造を明記
   - 依存関係を明確にする

### タスク表記の説明

- **[並列OK]**: 他のタスクと並列実行可能
- **[依存あり]**: 特定のタスク完了後に実行可能
- **依存**: `→ T##` で依存するタスクを明記

### 実行順序

```
T001: Terraformインフラ構築
  ↓
T002: 設定管理とユーティリティ（並列実行可能）
T003: Firestore重複排除機能（並列実行可能）
  ↓
T004: RSS記事取得機能（並列実行可能）
T005: Gemini評価機能（並列実行可能）
T006: Discord通知機能（並列実行可能）
  ↓
T007: メインオーケストレーション
  ↓
T008: ローカルテストとデバッグ
  ↓
T009: GCPデプロイとCI/CD
```

---

## T001: Terraformインフラ構築 🎯

**目的**: GCPリソースをTerraformで定義し、デプロイ可能な状態にする

**成果物**:
- Terraformモジュール（Firestore、Secret Manager、Cloud Scheduler、Cloud Functions）
- 本番環境設定（terraform/environments/prod/）
- 変数定義とoutputs定義

**タスク内容**:
1. terraform/modules/firestore/を作成（Firestore Database + インデックス定義）
2. terraform/modules/secrets/を作成（Secret Manager定義）
3. terraform/modules/scheduler/を作成（Cloud Scheduler + Pub/Sub定義）
4. terraform/modules/cloud-function/を作成（Cloud Functions Gen 2定義）
5. terraform/environments/prod/を作成（main.tf、variables.tf、outputs.tf、backend.tf）
6. terraform.tfvars.exampleを作成

**ブランチ**: `terraform/t001-infrastructure`

**コミットメッセージ例**:
```
terraform: Firestore、Secret Manager、Cloud Scheduler、Cloud FunctionsのTerraformモジュールと本番環境設定を追加
```

**完了条件**:
- [ ] terraform validateが成功
- [ ] 全モジュールにREADME.mdが含まれている
- [ ] terraform.tfvars.exampleが用意されている
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: なし（最初に実行）

**参照**:
- [plan.md](plan.md) (line 221-260)
- [contracts/firestore-schema.md](contracts/firestore-schema.md)

---

## T002: 設定管理とユーティリティ 🎯 [並列OK]

**目的**: 共通設定、エラーハンドリング、ログ、Secret Manager統合を実装

**成果物**:
- internal/config/ パッケージ（schema.go、loader.go、validator.go）
- internal/secrets/ パッケージ（manager.go）
- internal/errors/ パッケージ（errors.go）
- internal/logging/ パッケージ（logger.go）

**タスク内容**:
1. internal/config/schema.goを実装（RSSSource、InterestTopic、NotificationSettings構造体）
2. internal/config/loader.goを実装（config.json読み込み、GitHub URL対応）
3. internal/config/validator.goを実装（設定検証ロジック）
4. internal/secrets/manager.goを実装（Secret Managerクライアント）
5. internal/errors/errors.goを実装（カスタムエラー型）
6. internal/logging/logger.goを実装（構造化ログ）

**ブランチ**: `feat/t002-config-utils`

**コミットメッセージ例**:
```
feat: config、secrets、errors、loggingパッケージを追加し、設定ファイルの読み込みと検証、Secret Manager統合を実装

関連: T002
```

**完了条件**:
- [ ] 全パッケージにテストが含まれている
- [ ] config.jsonの読み込みと検証が動作する
- [ ] Secret Managerからシークレット取得が動作する
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T001完了後（Secret Managerリソースが必要）

**参照**:
- [data-model.md](data-model.md)
- [contracts/config-schema.md](contracts/config-schema.md)

---

## T003: Firestore重複排除機能 🎯 [並列OK]

**目的**: 通知済み・却下済み記事をFirestoreで追跡し、重複評価を防止

**成果物**:
- internal/storage/ パッケージ（firestore.go、notified.go、rejected.go）
- tests/contract/firestore_test.go（契約テスト）

**タスク内容**:
1. internal/storage/firestore.goを実装（Firestoreクライアントラッパー）
2. internal/storage/notified.goを実装（SaveNotifiedArticle、IsArticleNotified）
3. internal/storage/rejected.goを実装（SaveRejectedArticle、IsArticleRejected）
4. tests/contract/firestore_test.goを作成（Firestoreエミュレータテスト）
5. TTL（30日）ロジックを実装

**ブランチ**: `feat/t003-firestore-deduplication`

**コミットメッセージ例**:
```
feat: 通知済み・却下済み記事のFirestore追跡を実装し、30日TTLロジックと契約テストを追加

関連: T003
```

**完了条件**:
- [ ] Firestoreエミュレータでテストが成功
- [ ] 重複チェックが正しく動作
- [ ] TTLロジックが実装されている
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T001, T002完了後

**参照**:
- [contracts/firestore-schema.md](contracts/firestore-schema.md)

---

## T004: RSS記事取得機能 🎯 [並列OK]

**目的**: RSSフィード取得、パース、記事コンテンツ抽出を実装

**成果物**:
- internal/rss/ パッケージ（fetcher.go、parser.go）
- internal/article/ パッケージ（fetcher.go、extractor.go）
- tests/contract/rss_parser_test.go（契約テスト）

**タスク内容**:
1. internal/rss/fetcher.goを実装（RSSフィードURL取得）
2. internal/rss/parser.goを実装（RSS 2.0/Atom XMLパース）
3. internal/article/fetcher.goを実装（記事HTML取得）
4. internal/article/extractor.goを実装（go-readabilityで本文抽出）
5. tests/contract/rss_parser_test.goを作成（RSS 2.0/Atomフィクスチャテスト）

**ブランチ**: `feat/t004-rss-article-fetcher`

**コミットメッセージ例**:
```
feat: RSSフィード取得・パースと記事コンテンツ抽出を実装し、go-readabilityを使用した本文抽出を追加

関連: T004
```

**完了条件**:
- [ ] RSS 2.0とAtomフィードの両方をパース可能
- [ ] go-readabilityで記事本文抽出が動作
- [ ] 契約テストが成功
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T002完了後（エラーハンドリングとログが必要）

**参照**:
- [contracts/rss-feed.md](contracts/rss-feed.md)

---

## T005: Gemini評価機能 🎯 [並列OK]

**目的**: Gemini APIで記事の関連性評価とAI生成判定を実装

**成果物**:
- internal/llm/ パッケージ（client.go、evaluator.go、summarizer.go）
- tests/contract/gemini_api_test.go（契約テスト）

**タスク内容**:
1. internal/llm/client.goを実装（Gemini APIクライアント、レート制限）
2. internal/llm/evaluator.goを実装（関連性評価、AI判定、スコア計算）
3. internal/llm/summarizer.goを実装（記事要約生成）
4. tests/contract/gemini_api_test.goを作成（APIリクエスト/レスポンステスト）
5. 5カテゴリスコアリングロジック実装（最大100点）

**ブランチ**: `feat/t005-gemini-integration`

**コミットメッセージ例**:
```
feat: 記事の関連性評価、AI生成判定、要約生成を実装し、5カテゴリスコアリング（最大100点）とレート制限を追加
```

**完了条件**:
- [ ] Gemini APIとの通信が動作
- [ ] AI生成記事判定が機能
- [ ] スコアリングロジックが正しく動作
- [ ] レート制限が実装されている
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T002完了後（Secret Managerからのapi-key取得が必要）

**参照**:
- [contracts/gemini-api.md](contracts/gemini-api.md)
- [data-model.md](data-model.md)（スコアリング重み付け）

---

## T006: Discord通知機能 🎯 [並列OK]

**目的**: Discord Webhookで記事を通知する機能を実装

**成果物**:
- internal/discord/ パッケージ（client.go、formatter.go）
- tests/contract/discord_webhook_test.go（契約テスト）

**タスク内容**:
1. internal/discord/client.goを実装（Discord Webhook APIクライアント）
2. internal/discord/formatter.goを実装（Embedsペイロードフォーマット）
3. tests/contract/discord_webhook_test.goを作成（ペイロード検証テスト）
4. リトライロジック実装

**ブランチ**: `feat/t006-discord-notification`

**コミットメッセージ例**:
```
feat: Discord Webhook APIクライアントとEmbedsフォーマッターを実装し、リトライロジックと契約テストを追加
```

**完了条件**:
- [ ] Discord Webhook URLへの送信が動作
- [ ] Embedsフォーマットが正しい
- [ ] リトライロジックが実装されている
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T002完了後（Secret Managerからのwebhook URL取得が必要）

**参照**:
- [contracts/discord-webhook.md](contracts/discord-webhook.md)

---

## T007: メインオーケストレーション 🎯

**目的**: すべての機能を統合し、エンドツーエンドのフローを実装

**成果物**:
- cmd/curator/main.go（Cloud Functionエントリーポイント）
- メインオーケストレーションロジック

**タスク内容**:
1. cmd/curator/main.goを実装（Pub/Subトリガー対応）
2. メインフローの実装:
   - config.jsonロード
   - RSSフィード取得・パース
   - 重複チェック（Firestore）
   - 記事コンテンツ抽出
   - Gemini評価
   - 上位3〜5件選択
   - Discord通知
   - Firestore保存
3. エラーハンドリングとログ出力
4. 環境変数とSecret Manager統合

**ブランチ**: `feat/t007-main-orchestration`

**コミットメッセージ例**:
```
feat: 全機能を統合したCloud Functionエントリーポイントを実装し、RSS取得からDiscord通知までのエンドツーエンドフローを追加

関連: T007
```

**完了条件**:
- [ ] ローカルでgo runが動作
- [ ] エンドツーエンドフローが完走
- [ ] エラーハンドリングが適切
- [ ] ログ出力が構造化されている
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T003, T004, T005, T006完了後（すべての機能が必要）

**参照**:
- [spec.md](spec.md)（フロー全体）
- [plan.md](plan.md)

---

## T008: ローカルテストとデバッグ 🎯

**目的**: ローカル環境で全機能をテストし、バグを修正

**成果物**:
- テスト実行スクリプト
- デバッグログ
- バグ修正

**タスク内容**:
1. .envファイルを設定（Gemini API Key、Discord Webhook URL）
2. Firestoreエミュレータを起動
3. ローカルでgo run cmd/curator/main.goを実行
4. 実際のRSSフィードで動作確認
5. Discordへの通知を確認
6. バグ修正とリファクタリング

**ブランチ**: `test/t008-local-testing`

**コミットメッセージ例**:
```
test: ローカル環境での動作確認とバグ修正を実施
```

**完了条件**:
- [ ] ローカルで正常に動作
- [ ] Discordに記事が通知される
- [ ] エラーが適切にハンドリングされる
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T007完了後

---

## T009: GCPデプロイとCI/CD 🎯

**目的**: GCPにデプロイし、CI/CDパイプラインを構築

**成果物**:
- デプロイ済みCloud Function
- CI/CD設定（Cloud Build）
- 監視設定

**タスク内容**:
1. GCPプロジェクト作成と設定
2. gcloud auth loginとプロジェクト設定
3. Secret ManagerにGemini API Key、Discord Webhook URLを登録
4. Terraformでインフラをデプロイ（terraform apply）
5. Cloud Functionをデプロイ
6. Cloud Schedulerの動作確認
7. cloudbuild.yamlの設定
8. GitHub Actionsまたは手動デプロイフローの確認

**ブランチ**: `deploy/t009-gcp-deployment`

**コミットメッセージ例**:
```
deploy: Cloud FunctionとTerraformインフラをGCPにデプロイし、Cloud BuildによるCI/CDパイプラインを設定

関連: T009
```

**完了条件**:
- [ ] Cloud Functionがデプロイされている
- [ ] Cloud Schedulerが毎日8:00 JSTに実行される
- [ ] Discordに記事が通知される
- [ ] 監視とログが正常に動作
- [ ] ghコマンドでPRを作成（PRテンプレートを使用）
- [ ] 別のClaude Codeインスタンスでレビューを実施
- [ ] ユーザーの承認を得てからマージ

**依存**: T008完了後

**参照**:
- [quickstart.md](quickstart.md)

---

## 認証情報とAPIキー

### 必要な認証情報

1. **Googleアカウント**: GCPプロジェクト作成とTerraform実行
2. **Gemini API Key**: Google AI Studioから取得（T005, T008, T09で必要）
3. **Discord Webhook URL**: Discordサーバー設定から取得（T006, T008, T09で必要）

### .envファイル例

```bash
# .env（ローカル開発用、Gitignore対象）
GCP_PROJECT_ID=your-gcp-project-id
GEMINI_API_KEY=your-gemini-api-key
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
CONFIG_URL=https://raw.githubusercontent.com/your-repo/main/config.json
```

### 認証フロー

```
T001: Terraform開発
  ↓ gcloud auth login（GCP認証）
T002-T007: ローカル開発
  ↓ .envファイルに開発用APIキー設定
T008: ローカルテスト
  ↓ Gemini API Key、Discord Webhook URL取得
T009: GCPデプロイ
  ↓ Secret Managerにシークレット登録
  ↓ terraform apply（本番デプロイ）
```

---

## タスク一覧（チェックリスト）

- [ ] T001: Terraformインフラ構築
- [ ] T002: 設定管理とユーティリティ（並列OK）
- [ ] T003: Firestore重複排除機能（並列OK）
- [ ] T004: RSS記事取得機能（並列OK）
- [ ] T005: Gemini評価機能（並列OK）
- [ ] T006: Discord通知機能（並列OK）
- [ ] T007: メインオーケストレーション
- [ ] T008: ローカルテストとデバッグ
- [ ] T009: GCPデプロイとCI/CD

**MVP完成**: T009完了時点
