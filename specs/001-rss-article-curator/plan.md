# 実装計画: RSS記事キュレーションBot

**ブランチ**: `001-rss-article-curator` |   
**日付**: 2025-10-27 |   
 **仕様**: [spec.md](./spec.md)  
**入力**: `/specs/001-rss-article-curator/spec.md`からの機能仕様

**注意**: このテンプレートは`/speckit.plan`コマンドによって埋められます。実行ワークフローについては`.specify/templates/commands/plan.md`を参照してください。

## 概要

技術ブログまとめサイトを毎日監視し、Gemini LLMを使用してユーザー定義の興味に対する記事の関連性を評価し、要約付きの3〜5件のキュレーション記事をDiscordに投稿するサーバーレスRSS記事キュレーターを構築します。システムはGoogle Cloud Functions（Go）で実行され、JST午前9時にCloud Schedulerによってトリガーされ、重複排除追跡にFirestore、認証情報にSecret Managerを使用します。

**技術的アプローチ**: 非同期処理のためにCloud Functions + Pub/Subを使用したイベント駆動アーキテクチャ、コンテンツ抽出のためのgo-readability、golang.org/x/time/rateによるレート制限付きのLLM評価のためのGemini Flash API、リッチ通知のためのDiscord Webhook Embeds。設定はGitHubリポジトリのconfig.jsonファイル経由で管理。

## 技術コンテキスト

**言語/バージョン**: Go 1.21+（Cloud Functions Gen 2互換）
**主要依存関係**:
  - `goquery`（RSS/記事コンテンツのHTMLスクレイピング）
  - `go-readability`（記事テキスト抽出）
  - `golang.org/x/time/rate`（LLM APIレート制限）
  - `cloud.google.com/go/firestore`（Firestoreクライアント）
  - `cloud.google.com/go/secretmanager`（Secret Managerクライアント）
  - `google.golang.org/genai`（Gemini APIクライアント）

**ストレージ**: Firestore（記事重複排除のためのNoSQLドキュメントストア: notified_articles、rejected_articlesコレクション）

**テスト**:
  - `go test`（ネイティブGoテスト）
  - `testing/httptest`（契約テスト用のHTTPモックサーバー）
  - Discord Webhook API、Gemini API、RSSフィードパーサーの契約テスト
  - **注記**: Firestoreの統合テストは省略

**インフラストラクチャ**: Terraform（IaC - Infrastructure as Code）
  - GCPリソースの宣言的管理（Cloud Functions、Cloud Scheduler、Pub/Sub、Firestore、Secret Manager）
  - バージョン管理されたインフラ設定（`terraform/`ディレクトリ）
  - 本番環境（prod）のみの構成
  - モジュール化による再利用性の向上（modules/配下で共通リソースを定義）
  - 最小限の必要変数のみを使用（project_id、region）
  - インフラ変更の再現性と監査可能性

**ターゲットプラットフォーム**: Google Cloud Functions Gen 2（Linux、マネージドランタイム）

**プロジェクトタイプ**: 単一のサーバーレス関数プロジェクト（フロントエンドなし、スケジュールされたバックエンドのみ）

**パフォーマンス目標**:
  - 1時間のタイムアウト内で100〜200件の記事を処理（Cloud Functions最大実行時間: 60分）
  - Gemini API: 無料枠で15 RPM（分あたりのリクエスト数）と1500 RPD（日あたりのリクエスト数）を尊重
  - Discord Webhook: 分あたり30リクエストのレート制限を尊重
  - Firestoreの読み取り: 1日あたり<1000回（重複排除ルックアップ）、書き込み: 1日あたり<200回（追跡される新しい記事）

**制約**:
  - コスト: 1日あたり<5米ドル（Gemini API 1500 RPD未満のFlash無料枠で$0.00、Firestore無料枠: 1日あたり5万読み取り、2万書き込み、Cloud Functions: 月200万呼び出し無料）
  - タイムアウト: 最大1時間の実行時間（FR-016の解決: Cloud Functions Gen 2最大タイムアウト）
  - レート制限: Gemini API 15 RPM、Discord Webhook 30 RPM
  - 設定更新: GitHub rawファイルURLから読み取り（設定変更のための再ビルド不要）

**スケール/スコープ**:
  - RSSソース: 2〜3のまとめサイト（例: dev.to、Hacker News、Zenn.dev）
  - 毎日の記事: 100〜200件の新しい記事を取得
  - 毎日の通知: 3〜5件の記事をDiscordに投稿
  - 重複排除履歴: Firestoreに10,000件以上の記事URLを保存
  - 予想実行時間: 毎日の実行あたり30〜60分

## 憲章チェック

*ゲート: フェーズ0の調査前に合格する必要があります。フェーズ1の設計後に再チェック。*

### 原則I: ボット優先アーキテクチャ ✅ 合格

**評価**: Cloud Functionsは設計により明確なモジュール境界を強制します:
- RSSフェッチ、記事コンテンツ抽出、LLM評価、Discord投稿は別々のGoパッケージ
- テスト可能性のためにリポジトリインターフェースの背後に抽象化されたFirestoreクライアント
- ライブGCPなしでテストするために抽象化されたSecret Managerクライアント
- インターフェース経由で読み込まれる設定（ファイル対HTTPローダーが交換可能）
- 隠れたグローバル状態なし（依存性注入経由で渡されるFirestoreクライアント）

**コンプライアンス**: サーバーレスの制約により、アーキテクチャは自然に原則に従います。

### 原則II: 契約駆動統合テスト ✅ 合格

**評価**: すべての外部サービスには契約テストが必要です:
- Discord Webhook API: Embedsペイロード構造をテスト（httptestモックによる契約テスト）
- Gemini API: リクエスト/レスポンススキーマとレート制限処理をテスト（記録された応答による契約テスト）
- RSSフィード解析: 実際のRSS 2.0/Atomフィードサンプルに対してテスト（ローカルフィクスチャによる統合テスト）
- Firestore操作: Firestoreエミュレータを使用してクエリと書き込みをテスト（統合テスト）
- go-readability抽出: サンプルHTMLページに対してテスト（統合テスト）

**コンプライアンス**: テスト計画にはすべての外部依存関係の契約/統合テストが含まれます。

### 原則III: 優雅なデプロイ & ゼロダウンタイム更新 ⚠️ レビュー必要

**評価**: サーバーレス関数には組み込みの優雅なシャットダウンがありますが、状態の永続化には注意が必要です:
- Cloud Functionsはシャットダウン前に進行中のリクエストを自動的に完了 ✅
- Firestoreの永続化は再起動後も存続 ✅
- バージョン移行不要（ステートレス関数） ✅
- ヘルスチェック: Cloud Functions組み込みの準備プローブ ✅
- ロールバック: GitHub Actionsが`gcloud functions deploy --source=<git-sha>`経由で以前のバージョンを再デプロイ可能 ✅

**潜在的問題**: 関数がrun中にクラッシュした場合、部分的に評価された記事が再評価される可能性（無駄なAPIコール）。緩和策: Firestoreアトミック書き込みで「進行中」状態を追跡。

**コンプライアンス**: ほぼ準拠。クラッシュ時の再評価を避けるために「進行中」追跡の追加を検討。

### 原則IV: 障害時の信頼性 ✅ 合格

**評価**: 回復力のために設計されたエラー処理:
- RSSフェッチ失敗: 他のソースの処理を続行、エラーをログ（FR-014）
- 記事コンテンツフェッチ失敗: 記事をスキップ、エラーをログ（FR-018）
- Gemini APIレート制限: golang.org/x/time/rateを使用したジッター付き指数バックオフ（FR-013）
- Discord Webhook失敗: バックオフで3回再試行、すべて失敗した場合はログ
- Firestore失敗: 高速失敗（重複排除に重要）、エラーをログ
- タイムアウト: HTTPリクエストあたり5秒、関数全体で1時間のタイムアウト

**コンプライアンス**: すべての外部依存関係には再試行ロジックがあり、レートリミッター経由のサーキットブレーカーパターン。

### 原則V: セキュリティ & レート制限 ✅ 合格

**評価**: セキュリティ対策が実施されています:
- Google Secret Managerのシークレット（Discord Webhook URL、Gemini APIキー） - コードや環境変数に決して含めない ✅
- ユーザー入力なし（スケジュールされた関数、使用前に検証されるconfig.json） ✅
- レート制限: golang.org/x/time/rateがGemini APIの15 RPMを強制 ✅
- 監査ログ: Cloud FunctionsがすべてのエラーとAPI呼び出しをCloud Loggingに記録 ✅
- Firestoreセキュリティルール: すべて拒否（サービスアカウント経由のバックエンドのみのアクセス） ✅
- 設定検証: 処理前のJSONスキーマチェック ✅

**コンプライアンス**: セキュリティ違反なし。シークレット管理はGCPベストプラクティスに従います。

### 原則VI: シンプルさ & 保守性 ⚠️ レビュー必要

**評価**: 技術スタックの複雑さの評価:
- Go標準ライブラリ優先: `net/http`、`encoding/json`、`time`（ネイティブパッケージ） ✅
- 正当化された外部依存関係:
  - `goquery` - GoエコシステムでのHTML/RSS解析の標準 ✅
  - `go-readability` - カスタム記事抽出の記述を回避（実証済みライブラリ） ✅
  - `golang.org/x/time/rate` - 公式Goレートリミッター（サードパーティではない） ✅
  - `cloud.google.com/go/*` - 公式GCP SDK（プラットフォームに必要） ✅
- 早期の抽象化を回避: 3以上のデータソースまでリポジトリパターンなし、現時点では直接Firestore呼び出し ✅
- フレームワークなし: プレーンなCloud Functions HTTPハンドラー ✅

**潜在的懸念**: Cloud Functions + Pub/Sub + Firestore + Secret Manager = 4つのGCPサービス。憲章はシンプルさを好みます。

**正当化**:
- Cloud Scheduler → Pub/Sub → Cloud Functions: スケジュールされたタスクの標準GCPパターン（直接HTTPトリガーには認証の複雑さがある）
- Firestore: 重複排除のための最もシンプルなNoSQL（代替: Cloud StorageのJSONファイル - 遅い、インデックスなし）
- Secret Manager: 認証情報のGCPベストプラクティス（代替: 環境変数 - セキュリティが低い）

**コンプライアンス**: 依存関係は正当化されています。サーバーレス + セキュアなアーキテクチャに複雑さが必要。

### 全体的なゲートステータス: ✅ 合格（2つの小さなレビュー項目あり）

**アクションアイテム**:
1. 関数クラッシュ時の再評価を防ぐためにFirestoreに「進行中」追跡を追加（オプションの拡張）
2. 4つのGCPサービスが代替案よりもシンプルである理由を以下の複雑さ追跡セクションで文書化

**推奨事項**: フェーズ0の調査に進みます。フェーズ1の設計中にアクションアイテムに対処します。

## プロジェクト構造

### ドキュメント（この機能）

```text
specs/001-rss-article-curator/
├── plan.md              # このファイル（/speckit.planコマンド出力）
├── research.md          # フェーズ0出力（/speckit.planコマンド）
├── data-model.md        # フェーズ1出力（/speckit.planコマンド）
├── quickstart.md        # フェーズ1出力（/speckit.planコマンド）
├── contracts/           # フェーズ1出力（/speckit.planコマンド）
│   ├── discord-webhook.md
│   ├── gemini-api.md
│   └── firestore-schema.md
└── tasks.md             # フェーズ2出力（/speckit.tasksコマンド - /speckit.planでは作成されない）
```

### ソースコード（リポジトリルート）

```text
# サーバーレス単一関数プロジェクト構造

cmd/
└── curator/
    └── main.go                 # Cloud Functionsエントリポイント（HTTPハンドラー）

internal/
├── config/
│   ├── loader.go              # GitHubまたはローカルファイルからconfig.jsonをロード
│   └── schema.go              # 設定検証（RSSソース、興味、記事数）
├── rss/
│   ├── fetcher.go             # まとめサイトURLからRSSフィードを取得
│   └── parser.go              # RSS 2.0 / Atomフィードを解析
├── article/
│   ├── fetcher.go             # 記事HTMLコンテンツを取得
│   └── extractor.go           # go-readabilityを使用して記事テキストを抽出
├── llm/
│   ├── client.go              # レート制限付きGemini APIクライアント
│   ├── evaluator.go           # 記事の関連性を評価（スコア0〜100）
│   └── summarizer.go          # 記事要約を生成
├── storage/
│   ├── firestore.go           # Firestoreクライアントラッパー
│   ├── notified.go            # 通知された記事を追跡（URL + タイムスタンプ）
│   └── rejected.go            # 却下された記事を追跡（URL + 理由）
├── discord/
│   ├── client.go              # Discord Webhook APIクライアント
│   └── formatter.go           # 記事をEmbedsペイロードとしてフォーマット
└── secrets/
    └── manager.go             # Google Secret Managerクライアント

tests/
├── contract/
│   ├── discord_webhook_test.go     # Discord Embeds API契約
│   ├── gemini_api_test.go          # Geminiリクエスト/レスポンス契約
│   └── rss_parser_test.go          # フィクスチャによるRSSフィード解析
├── integration/
│   ├── firestore_test.go           # Firestoreエミュレータテスト
│   └── article_extraction_test.go  # go-readability統合
└── unit/
    ├── config_test.go              # 設定検証ユニットテスト
    ├── evaluator_test.go           # LLMスコアリングロジック（モックAPI）
    └── formatter_test.go           # Discord Embedsフォーマット

terraform/
├── environments/
│   └── prod/
│       ├── main.tf                 # 本番環境のリソース定義
│       ├── variables.tf            # 入力変数（project_id、region）
│       ├── outputs.tf              # 出力値（Cloud Functions URL、Pub/Sub topic等）
│       ├── terraform.tfvars        # 変数の実際の値（Gitignore対象）
│       ├── terraform.tfvars.example # 変数値の例
│       └── backend.tf              # GCS backend設定（tfstate管理）
└── modules/
    ├── cloud-function/
    │   ├── main.tf                 # Cloud Function Gen 2リソース定義
    │   ├── variables.tf            # モジュール変数
    │   └── outputs.tf              # モジュール出力
    ├── scheduler/
    │   ├── main.tf                 # Cloud Scheduler + Pub/Sub定義
    │   ├── variables.tf
    │   └── outputs.tf
    ├── firestore/
    │   ├── main.tf                 # Firestore Database + インデックス定義
    │   ├── variables.tf
    │   └── outputs.tf
    └── secrets/
        ├── main.tf                 # Secret Manager定義
        ├── variables.tf
        └── outputs.tf

config.json                         # 設定ファイル（RSSソース、興味）
go.mod                              # Goモジュール依存関係
go.sum                              # 依存関係チェックサム
.env.example                        # 環境変数の例（GCPプロジェクトID）
cloudbuild.yaml                     # GitHub Actions → Cloud Buildデプロイ
.gitignore                          # シークレット、ローカル設定を無視
README.md                           # プロジェクトセットアップとデプロイガイド
```

**構造決定**:
- **Goコード**: 単一のサーバーレス関数プロジェクト。すべてのコードはCloud Functionsエントリポイントの`cmd/curator/`の下にあり、ビジネスロジックは`internal/`パッケージにあります。これはサーバーレスアプリのGo規約に従い、構造をフラットに保ちます（ネストされたモジュールなし）。テストは`_test.go`サフィックスを介してコードパッケージと同じ場所に配置され、統合/契約テストは、ユニットテスト中にFirestoreエミュレータの起動オーバーヘッドを避けるために別の`tests/`ディレクトリにあります。
- **Terraformインフラ**: 本番環境（prod）のみの構成で、`terraform/environments/prod/`配下にリソース定義を配置。`terraform/modules/`配下に各GCPリソース（cloud-function、scheduler、firestore、secrets）を独立したモジュールとして定義し、`terraform/environments/prod/main.tf`で組み合わせて使用。必須変数は`project_id`と`region`の2つのみ。`environments/`ディレクトリ構造により、将来的な開発環境の追加（`environments/dev/`）に対する拡張性を確保。

**理論的根拠**:
- **Goコード**: Cloud Functions Gen 2は、エントリポイント関数を持つルートに単一の`go.mod`を期待します。`internal/`ディレクトリは外部インポートを防ぎます（プライベートパッケージのGo規約）。`tests/`を分離することで、Firestoreエミュレータの起動オーバーヘッドなしに高速ユニットテストを実行できます。
- **Terraformインフラ**: `terraform/environments/`ディレクトリ構成により、将来的に開発環境が必要になった場合は`environments/dev/`を追加するだけで対応可能。現時点では`environments/prod/`のみだが、環境分離の構造を最初から確保することで、後からのリファクタリングを回避。モジュール化により各リソースの責務が明確になり、複数環境でのモジュール再利用を実現。最小限の変数設計（project_id、regionのみ）により、設定ミスのリスクを低減し、シンプルな運用を実現。tfstateはGCS backendで管理し、チーム開発時の状態共有と競合防止を実現。

## 複雑さの追跡

> **憲章チェックに違反があり正当化が必要な場合のみ記入**

| 違反 | 必要な理由 | より簡単な代替案が却下された理由 |
|-----------|------------|-------------------------------------|
| 複数のGCPサービス（Cloud Scheduler、Pub/Sub、Functions、Firestore、Secret Manager） | サーバーレススケジュールタスクのGCPベストプラクティス | Cloud Schedulerからの直接HTTPトリガーには認証トークンの管理が必要（セキュリティが低く、より複雑）。Firestoreは高速URLルックアップのための最もシンプルなNoSQLオプション（Cloud Storageファイルは重複排除のために完全なファイル読み取りが必要）。Secret Managerは認証情報のGCP推奨（Cloud Functionsの環境変数はセキュリティが低く、ローテーションに再デプロイが必要）。 |
| go-readabilityライブラリ（サードパーティ） | HTMLからの記事テキスト抽出は複雑（ボイラープレート削除、主要コンテンツ検出） | カスタム抽出の記述には、何千ものブログサイトのHTML解析ルールの維持が必要。go-readabilityは実戦でテストされ、エッジケースを処理します。代替: 完全なHTMLをGemini APIに送信（ただしトークン制限を超え、コストが高い）。 |

**結論**: 複雑さはGCPプラットフォームの制約と実証済みライブラリの価値によって正当化されます。セキュリティや信頼性を犠牲にすることなく、よりシンプルな代替案はありません。
