# 調査と設計決定: RSS記事キュレーションBot

**機能**: RSS記事キュレーションBot  
**日付**: 2025-10-27  
**フェーズ**: フェーズ0 - 調査と技術検証

## 概要

このドキュメントは、計画フェーズで評価されたすべての技術選択、設計決定、および代替案を記録します。各決定には理由と却下された代替案が含まれています。

---

## 1. クラウドプラットフォームの選択

### 決定：Google Cloud Platform（GCP）

**理由**:
- 無料枠が予想される使用量をカバー（200万Cloud Functions呼び出し/月、5万Firestore読み取り/日）
- Gemini APIのネイティブ統合（同じエコシステム、クロスクラウド認証不要）
- Cloud Scheduler + Pub/Sub + Functionsはスケジュールされたサーバーレスタスクの標準パターン
- Secret Managerがインフラ管理なしで安全な資格情報ストレージを提供
- Firestore NoSQLが重複排除クエリを簡素化（インデックス検索 vs ファイル解析）

**検討された代替案**:

1. **AWS Lambda + DynamoDB + EventBridge**
   - 却下：ネイティブGemini API統合なし（個別認証で外部HTTP呼び出しが必要）
   - 却下：GCPに比べてGo関数のコールドスタートレイテンシーが高い
   - コスト：価格は類似だがGemini API無料枠の優位性なし

2. **Azure Functions + Cosmos DB + Timer Trigger**
   - 却下：Gemini API統合なし（外部API呼び出しが必要）
   - 却下：低ボリュームユースケースではCosmos DBがFirestoreより高価
   - コスト：無料枠がより制限的（GCPの200万に対して100万リクエスト/月）

3. **VPSでのセルフホスティング（DigitalOcean、Linode）**
   - 却下：手動インフラ管理（OSパッチ、監視、バックアップ）
   - 却下：自動スケーリングなし（サーバーレスは可変実行時間をより適切に処理）
   - コスト：トラフィックがなくても月5〜10ドルの基本コスト（GCP無料枠の0ドルに対して）

**結論**：GCPはGemini API統合、無料枠カバレッジ、サーバーレスの簡素性に最適。

---

## 2. スケジューリングメカニズム

### 決定：Cloud Scheduler → Pub/Sub → Cloud Functions

**理由**:
- Cloud Schedulerが毎日午前9時JSTにPub/Subトピックにメッセージを公開
- Pub/SubトピックにサブスクライブされたCloud Functionsが実行をトリガー
- スケジューリングと実行を分離（再試行はSchedulerではなくPub/Subで処理）
- Pub/Subは最低1回の配信を保証（べき等性はFirestore重複排除で処理）

**検討された代替案**:

1. **Cloud Scheduler → Cloud Functions（直接HTTPトリガー）**
   - 却下：認証のためのOAuth2トークン管理が必要
   - 却下：関数デプロイがダウンした場合の組み込み再試行メカニズムなし
   - 複雑性：2行のPub/Subトリガーに対して認証設定が10行以上の設定を追加

2. **VM上のCronジョブ**
   - 却下：VM管理が必要（パッチ、監視、コスト）
   - 却下：失敗時の自動再試行なし
   - コスト：Cloud Scheduler（無料枠：3ジョブ）の0ドルに対してVM月5〜10ドル

3. **GitHub Actionsスケジュールワークフロー**
   - 却下：シークレット管理がより安全でない（Secret ManagerとGitHub Secrets）
   - 却下：FirestoreアクセスにはサービスアカウントキーがGitHubで必要（セキュリティリスク）
   - 却下：実行ごとにコールドスタート（Cloud Functionsのようなウォームインスタンスなし）

**結論**：Pub/SubパターンはスケジュールされたサーバーレスタスクのGCPベストプラクティス。

---

## 3. 記事コンテンツ抽出

### 決定：go-readabilityライブラリ

**理由**:
- 1.5k以上のGitHubスターを持つ実績のあるライブラリ、アクティブに保守
- HTMLボイラープレート除去（広告、ナビゲーション、サイドバー）を自動処理
- 複数のブログプラットフォームをサポート（Medium、Dev.to、WordPress、カスタムサイト）
- Gemini APIトークン使用量を削減（完全なHTMLではなくクリーンテキストを送信）
- サイト固有の解析ルール保守不要

**検討された代替案**:

1. **goqueryでのカスタムHTML解析**
   - 却下：各ブログプラットフォームの解析ルール保守が必要
   - 却下：ブログサイトのリデザイン時に壊れる（高い保守負担）
   - 労力：go-readabilityの5行に対して100行以上のコード

2. **完全なHTMLをGemini APIに送信**
   - 却下：トークン制限を超える（ほとんどの記事はHTMLで50k以上のトークン）
   - 却下：API コストが高い（リクエストごとにトークンが課金される）
   - コスト：クリーンテキストに対して10倍のトークン

3. **Mercury Parser（execを介したNode.jsライブラリ）**
   - 却下：Dockerコンテナに Node.jsランタイムが必要（複雑性増加）
   - 却下：クロス言語exec呼び出しはネイティブGoより遅い
   - 複雑性：単一の`go get`に対してDockerfile + npm依存関係

**結論**：go-readabilityは最もシンプルで保守可能なソリューション。

---

## 4. LLMプロバイダーの選択

### 決定：Google Gemini Flash API

**理由**:
- 無料枠：1500リクエスト/日（評価される100〜200記事 + 3〜5要約をカバー）
- 低レイテンシー：Flashモデルは速度に最適化（応答時間<1秒）
- ネイティブGCP統合（クロスクラウド認証の複雑性なし）
- スコアリング用の構造化JSON出力をサポート（0〜100の関連性スケール）
- 日本語 + 英語サポート（ユースケースに一致）

**検討された代替案**:

1. **OpenAI GPT-4 Turbo**
   - 却下：10ドル/100万トークン（200記事で推定1日2〜3ドル）
   - 却下：無料枠なし（<5ドル/日コスト制約に違反）
   - レイテンシー：Gemini Flashと類似（応答時間1〜2秒）

2. **Anthropic Claude 3 Haiku**
   - 却下：0.25ドル/100万トークン入力（200記事で依然として1日0.50〜1ドル）
   - 却下：個別のAPIアカウント管理が必要（GCP統合請求に対して）
   - レイテンシー：GPT-4より速いがGemini Flashと類似

3. **GPU上のセルフホストLlama 3**
   - 却下：GPU VM必要（推論可能インスタンスで月50〜100ドル）
   - 却下：モデルダウンロード + ホスティングの複雑性（API呼び出しの簡素性に対して）
   - 保守：継続的なモデル更新、スケーリング、監視 vs ゼロオペAPI

4. **Groq API（高速推論）**
   - 却下：無料枠が14kトークン/分に制限（200記事のボトルネック）
   - 却下：構造化JSON出力保証なし（スコア解析が困難）

**結論**：Gemini Flash APIは適切なパフォーマンスで<5ドル/日制約を満たす唯一のオプション。

---

## 5. レート制限戦略

### 決定：golang.org/x/time/rateトークンバケットリミッター

**理由**:
- 公式Go拡張パッケージ（Goチームによる保守）
- トークンバケットアルゴリズムがAPIレート制限エラーを防止（Geminiで15リクエスト/分）
- 設定可能なバースト（1〜2のバーストリクエストを許可、その後15 RPMの定常状態を強制）
- 待機/再試行を自動処理（次のトークンが利用可能になるまでスリープ）

**検討された代替案**:

1. **リクエスト間の手動スリープ（time.Sleep）**
   - 却下：可変リクエスト時間を考慮しない（時間の無駄）
   - 却下：バースト処理なし（クォータが利用可能でも4秒スリープを強制）
   - 効率：トークンバケットは固定スリープに対してクォータを最適に使用

2. **指数バックオフのみ（事前制限なし）**
   - 却下：最初にレート制限エラーをトリガー、その後再試行（API呼び出しの無駄）
   - 却下：Gemini APIは失敗したレート制限リクエストに課金（コストが高い）
   - ユーザー体験：レート制限エラーがログに表示（失敗のように見える）

3. **サードパーティレート制限（github.com/uber-go/ratelimit）**
   - 却下：Go標準拡張に既にある機能の依存関係を追加
   - 憲章違反：サードパーティよりもGo stdlib/公式拡張を優先

**結論**：golang.org/x/time/rateはAPIレート制限の標準Goソリューション。

---

## 6. 重複排除ストレージ

### 決定：Firestore（2コレクション：notified_articles、rejected_articles）

**理由**:
- インデックス検索に最適化されたNoSQLドキュメントストア（URLによるO(1)クエリ）
- 無料枠が5万読み取り/日をカバー（1日100〜200のURLチェックに十分）
- URLフィールドの自動インデックス（手動インデックス管理不要）
- アトミック書き込みが競合状態を防止（実行中に関数がクラッシュした場合）
- ネイティブGCP統合（サービスアカウント認証、接続文字列不要）

**スキーマ**:
```
notified_articles/
  {article_url}: {
    notified_at: timestamp,
    discord_message_id: string,
    article_title: string
  }

rejected_articles/
  {article_url}: {
    evaluated_at: timestamp,
    reason: string (例："low_relevance", "no_topic_match"),
    relevance_score: number
  }
```

**検討された代替案**:

1. **Cloud Storage JSONファイル（notified_urls.json）**
   - 却下：すべてのルックアップでファイル全体を読み取る必要がある（O(1)に対してO(n)）
   - 却下：アトミック書き込みなし（並行実行でファイルが破損する可能性）
   - パフォーマンス：10,000 URL = 毎回チェックで約500KBのファイル読み取り（遅い + 高コスト）

2. **"status"フィールドを持つFirestore単一コレクション**
   - 却下：ストレージの無駄（却下された記事にはtitle/discord_message_idフィールドは不要）
   - 却下：クエリの複雑性（statusフィールドでフィルタリング vs 直接コレクション読み取り）
   - コスト：最小限の節約だが複雑性増加（簡素性原則に違反）

3. **Redis / Memorystore**
   - 却下：永続性管理が必要（Redisはインメモリ、バックアップが必要）
   - 却下：無料枠なし（Memorystoreインスタンスで最低月35ドル）
   - コスト：<5ドル/日制約に違反

4. **PostgreSQL / Cloud SQL**
   - 却下：キーバリュールックアップには過剰（リレーショナルDB不要）
   - 却下：無料枠なし（Cloud SQLインスタンスで最低月10ドル）
   - コスト：<5ドル/日制約に違反

**結論**：FirestoreはGCP無料枠でのインデックスURL検索に最適。

---

## 7. 設定管理

### 決定：GitHubリポジトリ内のconfig.json（raw.githubusercontent.com経由で取得）

**理由**:
- バージョン管理された設定（Git履歴が変更を追跡）
- 再デプロイ不要（Cloud Functionsが各実行時に最新を取得）
- 編集が簡単（GitHub Web UIまたはgit commit）
- パブリックリポジトリが許容可能（設定にシークレットなし、RSSURLと興味のみ）

**スキーマ**:
```json
{
  "rss_sources": [
    {"url": "https://dev.to/feed", "name": "Dev.to", "enabled": true},
    {"url": "https://zenn.dev/feed", "name": "Zenn", "enabled": true}
  ],
  "interests": [
    {"topic": "Go", "aliases": ["Golang", "Go言語"], "priority": "high"},
    {"topic": "Kubernetes", "aliases": ["k8s"], "priority": "medium"}
  ],
  "notification_settings": {
    "max_articles": 5,
    "min_articles": 3,
    "min_relevance_score": 70
  }
}
```

**検討された代替案**:

1. **Cloud Functions内の環境変数**
   - 却下：設定変更に再デプロイが必要（FR-010、FR-011に違反）
   - 却下：文字数制限（32KB）が多数のRSSソースで超過する可能性
   - ユーザー体験：管理者が関数を再デプロイせずに興味を更新できない

2. **Firestoreコレクション（config_settingsドキュメント）**
   - 却下：バージョン履歴なし（悪い設定変更をロールバックできない）
   - 却下：編集に管理UIが必要（GitHub Web UIに対して）
   - 却下：Firestore書き込みがクォータを消費（無料GitHub取得に対して）

3. **Cloud Storageバケット（config.jsonファイル）**
   - 却下：パブリックアクセスのためのIAMセットアップが必要（複雑性増加）
   - 却下：組み込みバージョニングなし（バケットバージョニングを有効にする必要がある）
   - ユーザー体験：GitHub Web UIよりも編集が困難

4. **ソースコードにハードコード**
   - 却下：FR-010、FR-011に違反（外部設定要件）
   - 却下：すべての設定更新にコード変更 + 再デプロイが必要
   - 憲章違反：設定は外部化されなければならない

**結論**：GitHub ホストconfig.jsonはバージョン管理された再デプロイ不要の設定に最もシンプル。

---

## 8. Discord通知形式

### 決定：Embedsペイロードを持つWebhook API

**理由**:
- Embedsがリッチフォーマットを提供（タイトル、説明、URL、色、タイムスタンプ）
- 通知ごとに単一のwebhook POST（1メッセージに3〜5記事をバッチ）
- ボットアカウント不要（webhookはOAuth2ボットトークンよりシンプル）
- レート制限：30リクエスト/分（1日1通知に十分）

**Embedsスキーマ**:
```json
{
  "content": "📰 Daily Tech Article Digest - 2025-10-27",
  "embeds": [
    {
      "title": "Article Title",
      "description": "LLM-generated summary (200 chars max)...",
      "url": "https://article-url.com",
      "color": 5814783,
      "fields": [
        {"name": "Relevance", "value": "95/100", "inline": true},
        {"name": "Topics", "value": "Go, Kubernetes", "inline": true}
      ],
      "footer": {"text": "Source: Dev.to"}
    }
  ]
}
```

**検討された代替案**:

1. **コマンド付きDiscord Bot（OAuth2トークン）**
   - 却下：ボットアプリケーションセットアップ + OAuth2フローが必要（複雑）
   - 却下：ボット権限管理（webhook URLの簡素性に対して）
   - 却下：ユーザーストーリースコープはボットコマンドを除外（FR-007は投稿のみ要求）

2. **プレーンテキストメッセージ（Embedsなし）**
   - 却下：UXが悪い（クリック可能なタイトルなし、視覚階層なし）
   - 却下：スキャンが困難（ユーザーは素早い視覚ダイジェストが欲しい）
   - ユーザー体験：Embedsは可読性を大幅に向上

3. **複数メッセージ（記事ごとに1つ）**
   - 却下：Discordレート制限をトリガー（1バッチに対して5メッセージ = 5リクエスト）
   - 却下：Discordチャネルをスパム（ユーザーは単一ダイジェストメッセージが欲しい）
   - ユーザー体験：単一メッセージは興味がない場合にスクロールが容易

**結論**：Webhook + Embedsはリッチな単一メッセージ通知に最もシンプル。

---

## 9. テスト戦略

### 決定：3層テストピラミッド（ユニット、統合、契約）

**理由**:
- ユニットテスト：高速、外部依存関係なし（Firestore、Gemini APIをモック）
- 統合テスト：Firestoreエミュレータ、サンプルHTML上のgo-readability
- 契約テスト：実際のAPIスキーマ（Discord Embeds、Geminiリクエスト/応答）
- 憲章原則II（契約駆動統合テスト）に整合

**テスト計画**:

| テストタイプ | カバレッジ | ツール | 例 |
|-----------|----------|-------|----------|
| ユニットテスト | ビジネスロジックの70%以上 | `go test`、モック | 設定検証、スコアリングロジック、フォーマッター |
| 統合テスト | 外部ライブラリ | Firestoreエミュレータ、`go-readability` | Firestoreクエリ、記事抽出 |
| 契約テスト | APIスキーマ | `httptest`、記録された応答 | Discord Embedsペイロード、Gemini API JSON |

**検討された代替案**:

1. **モックのみのユニットテスト**
   - 却下：APIスキーマ変更をキャッチしない（Discord、Geminiの破壊的変更）
   - 却下：go-readability動作がテストされない（HTMLエッジケースで壊れる可能性）
   - 憲章違反：原則IIは契約/統合テストを要求

2. **ライブAPIに対するエンドツーエンドテスト**
   - 却下：お金がかかる（Gemini APIがテストリクエストに課金）
   - 却下：遅い（ネットワークレイテンシー + レート制限 = 1分以上のテストスイート）
   - 却下：不安定（ネットワーク障害、APIダウンタイムでテストが壊れる）

3. **手動テストのみ**
   - 却下：回帰検出なし（変更が既存機能を壊す可能性）
   - 却下：遅いフィードバックループ（デプロイ → テスト → ロールバック vs 即座のCIフィードバック）
   - 憲章違反：テストカバレッジ要件（最低70%）

**結論**：3層ピラミッドは高速フィードバック、コスト、憲章コンプライアンスのバランスを取る。

---

## 10. デプロイパイプライン

### 決定：GitHub Actions → Cloud Build → Cloud Functions

**理由**:
- GitHub Actionsがmainブランチへのプッシュでトリガー
- Cloud BuildがGoバイナリをコンパイルしてCloud Functionsにアップロード
- シークレットはGoogle Secret Manager経由で注入（GitHub Secretsではない）
- `gcloud functions deploy --source=<git-sha>`コマンドでロールバック

**ワークフロー**:
```yaml
# .github/workflows/deploy.yml
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: google-github-actions/setup-gcloud@v1
      - run: gcloud functions deploy curator --gen2 --runtime=go121 ...
```

**検討された代替案**:

1. **`gcloud` CLI経由の手動デプロイ**
   - 却下：CI/CDなし（人的エラーリスク、デプロイ前のテスト忘れ）
   - 却下：デプロイ履歴なし（すべてのデプロイを追跡するGitコミットに対して）
   - 憲章違反：デプロイは自動化されなければならない

2. **GitHubプッシュでのCloud Buildトリガー（GitHub Actionsなし）**
   - 却下：ワークフローの制御が少ない（例：デプロイ前にテストを実行できない）
   - 却下：GitHub Actions無料枠で十分（2000分/月）

3. **インフラストラクチャ as コードのTerraform**
   - 却下：単一Cloud Functionには過剰（複雑性増加）
   - 却下：Cloud Functions設定はほとんど変更されない（頻繁に変更されるコードに対して）
   - 複雑性：10行のGitHub Actionsワークフローに対して50行以上のTerraform

**結論**：GitHub ActionsはGCP統合でのCI/CDに最もシンプル。

---

## 主要決定のサマリー

| 決定領域 | 選択 | 主要な理由 |
|---------------|--------|-------------------|
| クラウドプラットフォーム | Google Cloud Platform（GCP） | Gemini API統合 + 無料枠 |
| コンピューティング | Cloud Functions Gen 2（Go） | サーバーレス + 1時間タイムアウト + 200万無料呼び出し |
| スケジューリング | Cloud Scheduler → Pub/Sub | 標準パターン + 組み込み再試行 |
| ストレージ | Firestore（notified_articles、rejected_articles） | O(1)インデックス検索 + 無料枠 |
| LLM | Google Gemini Flash API | 0ドルコスト（1500 RPD無料枠） + 低レイテンシー |
| レート制限 | golang.org/x/time/rate | 公式Goパッケージ + トークンバケットアルゴリズム |
| コンテンツ抽出 | go-readabilityライブラリ | 実績あり、保守、HTML複雑性を処理 |
| 設定 | GitHub config.json（生URL） | バージョン管理 + 再デプロイ不要 + 簡単編集 |
| 通知 | Discord Webhook + Embeds | リッチフォーマット + 単一メッセージバッチ |
| テスト | ユニット + 統合 + 契約 | 憲章コンプライアンス + 高速フィードバック |
| デプロイ | GitHub Actions → Cloud Build | 自動化CI/CD + 無料枠 |

---

## 未解決の質問

なし。すべての技術的明確化が解決済み：
- **FR-016タイムアウト**：1時間として解決（Cloud Functions Gen 2最大）
- **設定ストレージ**：GitHubリポジトリconfig.jsonとして解決
- **LLMプロバイダー**：Gemini Flash APIとして解決（コスト + パフォーマンスが最適）

**次フェーズ**：フェーズ1（データモデル + 契約 + クイックスタート）に進む
