# データモデル: RSS記事キュレーションBot

**機能**: RSS記事キュレーションBot  
**日付**: 2025-10-27  
**フェーズ**: フェーズ1 - データモデル設計

## 概要

このドキュメントは、すべてのデータエンティティ、その属性、関係、および検証ルールを定義します。システムは永続化にFirestore（NoSQLドキュメントストア）を使用し、処理にはインメモリ構造体を使用します。

---

## 1. 設定エンティティ（config.json）

### 1.1 RSS Source

**説明**: 監視するRSSフィードアグリゲーターサイトを表す

**属性**:
- `url`（string、必須）：RSS/Atomフィードへの完全なURL（例："https://dev.to/feed"）
- `name`（string、必須）：人間が読めるソース名（例："Dev.to"）
- `enabled`（boolean、必須）：このソースがアクティブに監視されているかどうか

**検証ルール**:
- `url`は有効なHTTP/HTTPS URLでなければならない
- `url`はContent-Type: application/rss+xmlまたはapplication/atom+xmlを返さなければならない
- `name`は1〜50文字でなければならない
- `enabled`は明示的にtrue/falseでなければならない（nullは不可）

**例**:
```json
{
  "url": "https://dev.to/feed",
  "name": "Dev.to",
  "enabled": true
}
```

**Go構造体**:
```go
type RSSSource struct {
    URL     string `json:"url" validate:"required,url"`
    Name    string `json:"name" validate:"required,min=1,max=50"`
    Enabled bool   `json:"enabled"`
}
```

---

### 1.2 InterestTopic

**説明**: 記事をフィルタリングするために使用される技術タグまたはキーワード

**属性**:
- `topic`（string、必須）：プライマリトピック名（例："Go"、"Kubernetes"）
- `aliases`（[]string、オプション）：トピックマッチングのための代替名（例：["Golang", "Go言語"]）
- `priority`（string、必須）：マッチング優先度："high"、"medium"、"low"

**検証ルール**:
- `topic`は1〜50文字でなければならない
- `topic`はすべての興味の中で一意でなければならない
- `aliases`の各要素は1〜50文字でなければならない
- `priority`は次のいずれかでなければならない："high"、"medium"、"low"
- "high"優先度のトピックはスコア2倍、"medium"は1倍、"low"は0.5倍

**例**:
```json
{
  "topic": "Go",
  "aliases": ["Golang", "Go言語"],
  "priority": "high"
}
```

**Go構造体**:
```go
type InterestTopic struct {
    Topic    string   `json:"topic" validate:"required,min=1,max=50"`
    Aliases  []string `json:"aliases,omitempty"`
    Priority string   `json:"priority" validate:"required,oneof=high medium low"`
}
```

---

### 1.3 NotificationSettings

**説明**: Discord通知のグローバル設定

**属性**:
- `max_articles`（int、必須）：1日あたりの投稿可能な最大記事数（デフォルト：5）
- `min_articles`（int、必須）：投稿前の最小記事数（デフォルト：3）
- `min_relevance_score`（int、必須）：資格を得るための最小LLMスコア（0〜100）（デフォルト：70）

**検証ルール**:
- `max_articles`は1〜10でなければならない（Discordスパムを防止）
- `min_articles`は1〜10でmax_articles以下でなければならない
- `min_relevance_score`は0〜100でなければならない

**例**:
```json
{
  "max_articles": 5,
  "min_articles": 3,
  "min_relevance_score": 70
}
```

**Go構造体**:
```go
type NotificationSettings struct {
    MaxArticles        int `json:"max_articles" validate:"required,min=1,max=10"`
    MinArticles        int `json:"min_articles" validate:"required,min=1,max=10"`
    MinRelevanceScore  int `json:"min_relevance_score" validate:"required,min=0,max=100"`
}
```

---

### 1.4 Config（ルート）

**説明**: config.jsonから読み込まれるルート設定オブジェクト

**属性**:
- `rss_sources`（[]RSSSource、必須）：監視するRSSフィードのリスト
- `interests`（[]InterestTopic、必須）：ユーザー定義の興味トピックのリスト
- `notification_settings`（NotificationSettings、必須）：通知動作設定

**検証ルール**:
- `rss_sources`は1〜10エントリでなければならない（過負荷を防止）
- 少なくとも1つのソースが`enabled: true`でなければならない
- `interests`は1〜50エントリでなければならない
- interestsに重複する`topic`値があってはならない

**例**:
```json
{
  "rss_sources": [...],
  "interests": [...],
  "notification_settings": {...}
}
```

**Go構造体**:
```go
type Config struct {
    RSSSources           []RSSSource          `json:"rss_sources" validate:"required,min=1,max=10,dive"`
    Interests            []InterestTopic      `json:"interests" validate:"required,min=1,max=50,dive"`
    NotificationSettings NotificationSettings `json:"notification_settings" validate:"required"`
}
```

---

## 2. 処理エンティティ（インメモリ）

### 2.1 Article

**説明**: RSSフィードから発見された技術ブログ投稿

**属性**:
- `title`（string、必須）：記事の見出し
- `url`（string、必須）：記事への正規URL（一意IDとして使用）
- `published_date`（time.Time、オプション）：RSSフィードからの公開タイムスタンプ
- `source_feed`（string、必須）：RSSソースの名前（例："Dev.to"）
- `content_text`（string、オプション）：抽出された記事本文テキスト（go-readabilityから）
- `fetched_at`（time.Time、必須）：記事が発見されたタイムスタンプ

**検証ルール**:
- `url`は有効なHTTP/HTTPS URLでなければならない
- `url`はグローバルに一意でなければならない（Firestore重複排除チェックで強制）
- `content_text`の長さは100〜50,000文字であるべき（短すぎる＝記事ではない、長すぎる＝ペイウォール/エラー）
- `title`は5〜500文字でなければならない

**例**:
```go
Article{
    Title: "Building Microservices with Go and Kubernetes",
    URL: "https://dev.to/example/building-microservices",
    PublishedDate: time.Parse(...),
    SourceFeed: "Dev.to",
    ContentText: "In this article, we explore...",
    FetchedAt: time.Now(),
}
```

**Go構造体**:
```go
type Article struct {
    Title         string    `json:"title"`
    URL           string    `json:"url" validate:"required,url"`
    PublishedDate time.Time `json:"published_date,omitempty"`
    SourceFeed    string    `json:"source_feed"`
    ContentText   string    `json:"content_text,omitempty"`
    FetchedAt     time.Time `json:"fetched_at"`
}
```

---

### 2.2 ArticleEvaluation

**説明**: 記事のLLM分析結果

**属性**:
- `article_url`（string、必須）：Article.URLへの参照
- `relevance_score`（int、必須）：Gemini APIからの0〜100スコア（高いほど関連性が高い）
- `matching_topics`（[]string、必須）：一致したInterestTopic.topic名のリスト
- `summary`（string、必須）：LLM生成の要約（Discord Embed用に最大200文字）
- `evaluated_at`（time.Time、必須）：評価のタイムスタンプ
- `is_relevant`（bool、必須）：スコアがmin_relevance_score以上の場合true

**検証ルール**:
- `relevance_score`は0〜100でなければならない
- `summary`は50〜200文字でなければならない（短すぎる＝不完全、長すぎる＝Discord Embed制限）
- `is_relevant == true`の場合、`matching_topics`は少なくとも1つのトピックを含まなければならない
- `is_relevant` = trueは`relevance_score >= NotificationSettings.min_relevance_score`の場合

**例**:
```go
ArticleEvaluation{
    ArticleURL: "https://dev.to/example/building-microservices",
    RelevanceScore: 85,
    MatchingTopics: []string{"Go", "Kubernetes"},
    Summary: "A comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters.",
    EvaluatedAt: time.Now(),
    IsRelevant: true,
}
```

**Go構造体**:
```go
type ArticleEvaluation struct {
    ArticleURL     string    `json:"article_url" validate:"required,url"`
    RelevanceScore int       `json:"relevance_score" validate:"min=0,max=100"`
    MatchingTopics []string  `json:"matching_topics"`
    Summary        string    `json:"summary" validate:"required,min=50,max=200"`
    EvaluatedAt    time.Time `json:"evaluated_at"`
    IsRelevant     bool      `json:"is_relevant"`
}
```

---

### 2.3 CuratedArticle

**説明**: Discord通知用のArticle + ArticleEvaluationの組み合わせ

**属性**:
- `article`（Article、必須）：完全な記事メタデータ
- `evaluation`（ArticleEvaluation、必須）：LLMスコアリングと要約
- `rank`（int、必須）：選択された記事内のランキング（1＝最高スコア）

**検証ルール**:
- `article.URL`は`evaluation.article_url`と一致しなければならない
- `rank`は1〜5でなければならない（max_articles制限に一致）

**例**:
```go
CuratedArticle{
    Article: Article{...},
    Evaluation: ArticleEvaluation{...},
    Rank: 1,
}
```

**Go構造体**:
```go
type CuratedArticle struct {
    Article    Article           `json:"article"`
    Evaluation ArticleEvaluation `json:"evaluation"`
    Rank       int               `json:"rank" validate:"min=1,max=10"`
}
```

---

## 3. 永続化エンティティ（Firestore）

### 3.1 NotifiedArticle（コレクション：notified_articles）

**説明**: Discordに投稿された記事の記録（重複排除）

**ドキュメントID**: `article_url`（例："https://dev.to/example/post"）

**属性**:
- `notified_at`（timestamp、必須）：Discordに投稿された日時
- `discord_message_id`（string、必須）：投稿されたメッセージのDiscord Snowflake ID
- `article_title`（string、必須）：記事のタイトル（ログ/デバッグ用）
- `relevance_score`（int、必須）：通知時のスコア

**インデックス**:
- プライマリ：ドキュメントID（article_url）- 自動
- セカンダリ：`notified_at`（TTL/クリーンアップクエリ用）

**TTLポリシー**: ストレージクォータを節約するために、オプションで90日以上前のドキュメントを削除

**検証ルール**:
- ドキュメントIDは有効なURLでなければならない
- `discord_message_id`は17〜19桁のSnowflake IDでなければならない
- `article_title`は元のArticle.titleと一致しなければならない

**例**:
```go
// Firestoreドキュメント: notified_articles/https:--dev.to-example-post
{
    "notified_at": "2025-10-27T09:15:00Z",
    "discord_message_id": "1234567890123456789",
    "article_title": "Building Microservices with Go",
    "relevance_score": 85
}
```

**Go構造体**:
```go
type NotifiedArticle struct {
    NotifiedAt       time.Time `firestore:"notified_at"`
    DiscordMessageID string    `firestore:"discord_message_id"`
    ArticleTitle     string    `firestore:"article_title"`
    RelevanceScore   int       `firestore:"relevance_score"`
}
```

**Firestoreセキュリティルール**:
```javascript
match /notified_articles/{articleURL} {
  allow read, write: if false; // サービスアカウント経由のバックエンドのみアクセス
}
```

---

### 3.2 RejectedArticle（コレクション：rejected_articles）

**説明**: 興味がないと評価された記事の記録（再評価を回避）

**ドキュメントID**: `article_url`（例："https://dev.to/example/irrelevant-post"）

**属性**:
- `evaluated_at`（timestamp、必須）：LLMが記事を評価した日時
- `reason`（string、必須）：却下理由："low_relevance" | "no_topic_match" | "content_extraction_failed"
- `relevance_score`（int、オプション）：評価された場合のスコア（抽出失敗の場合はnull）

**インデックス**:
- プライマリ：ドキュメントID（article_url）- 自動
- セカンダリ：`evaluated_at`（TTL/クリーンアップクエリ用）

**TTLポリシー**: オプションで30日以上前のドキュメントを削除（記事が著者によって更新される可能性がある）

**検証ルール**:
- `reason`は次のいずれかでなければならない："low_relevance"、"no_topic_match"、"content_extraction_failed"
- 理由が"low_relevance"または"no_topic_match"の場合、`relevance_score`は必須

**例**:
```go
// Firestoreドキュメント: rejected_articles/https:--dev.to-example-irrelevant
{
    "evaluated_at": "2025-10-27T09:05:00Z",
    "reason": "low_relevance",
    "relevance_score": 35
}
```

**Go構造体**:
```go
type RejectedArticle struct {
    EvaluatedAt    time.Time `firestore:"evaluated_at"`
    Reason         string    `firestore:"reason"` // "low_relevance" | "no_topic_match" | "content_extraction_failed"
    RelevanceScore *int      `firestore:"relevance_score,omitempty"` // オプションフィールドのためのポインタ
}
```

**Firestoreセキュリティルール**:
```javascript
match /rejected_articles/{articleURL} {
  allow read, write: if false; // サービスアカウント経由のバックエンドのみアクセス
}
```

---

## 4. 外部APIペイロード

### 4.1 Discord Embed（アウトバウンド）

**説明**: リッチな記事通知のためのDiscord Webhook Embedsペイロード

**属性**:
- `content`（string、必須）：メッセージ本文（例："📰 Daily Tech Digest - 2025-10-27"）
- `embeds`（[]Embed、必須）：記事埋め込みの配列（メッセージあたり最大10）

**Embedオブジェクト**:
- `title`（string、必須）：記事のタイトル（クリック可能）
- `description`（string、必須）：LLM要約（最大200文字）
- `url`（string、必須）：記事URL
- `color`（int、必須）：埋め込みの色（10進数、例：5814783 = 青）
- `fields`（[]Field、必須）：キーと値のペア（関連性、トピック）
- `footer`（Footer、必須）：ソースフィード名

**検証ルール**:
- メッセージあたり最大10埋め込み（Discord API制限）
- `title`最大256文字（Discord制限）
- `description`最大4096文字（Discord制限、ただし200を使用）
- `color`は0〜16777215でなければならない（24ビットRGB）

**例**:
```json
{
  "content": "📰 Daily Tech Article Digest - 2025-10-27",
  "embeds": [
    {
      "title": "Building Microservices with Go",
      "description": "A comprehensive guide to scalable microservices...",
      "url": "https://dev.to/example/post",
      "color": 5814783,
      "fields": [
        {"name": "Relevance", "value": "85/100", "inline": true},
        {"name": "Topics", "value": "Go, Kubernetes", "inline": true}
      ],
      "footer": {"text": "Source: Dev.to"}
    }
  ]
}
```

---

### 4.2 Gemini APIリクエスト（アウトバウンド）

**説明**: 記事評価のためのGemini Flash APIリクエスト

**属性**:
- `contents`（[]Content、必須）：記事テキストを含むユーザープロンプト
- `generationConfig`（GenerationConfig、必須）：応答形式設定

**Contentオブジェクト**:
- `role`（string）："user"
- `parts`（[]Part）：プロンプトテキスト

**GenerationConfig**:
- `temperature`（float）：0.3（一貫したスコアリングのために低く）
- `responseMimeType`（string）："application/json"（構造化出力）

**例**:
```json
{
  "contents": [{
    "role": "user",
    "parts": [{
      "text": "Evaluate this article for relevance to [Go, Kubernetes]:\n\n[Article text...]"
    }]
  }],
  "generationConfig": {
    "temperature": 0.3,
    "responseMimeType": "application/json"
  }
}
```

---

### 4.3 Gemini API応答（インバウンド）

**説明**: スコアリングJSONを含むGemini Flash API応答

**属性**:
- `candidates`（[]Candidate）：応答の代替案（最初のものを使用）

**Candidateオブジェクト**:
- `content`（Content）：応答コンテンツ
- `finishReason`（string）："STOP"（正常完了）

**Content内の期待されるJSON**:
```json
{
  "relevance_score": 85,
  "matching_topics": ["Go", "Kubernetes"],
  "summary": "A comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters.",
  "reasoning": "Article covers Go best practices and Kubernetes deployment patterns."
}
```

**検証**:
- `relevance_score`は0〜100の整数でなければならない
- スコアが0より大きい場合、`matching_topics`は空でない配列でなければならない
- `summary`は50〜200文字でなければならない

---

## 5. データフロー

```
1. config.jsonを読み込む → Config
2. RSSフィードを取得 → []Article
3. Firestoreに対して重複排除 → []Article（フィルタ済み）
4. コンテンツを抽出（go-readability） → []Article（content_text付き）
5. Gemini APIで評価 → []ArticleEvaluation
6. 却下された記事を保存 → RejectedArticle（Firestore）
7. relevance_scoreでソート → []CuratedArticle（上位3〜5）
8. Discordに投稿 → DiscordEmbedペイロード
9. 通知された記事を保存 → NotifiedArticle（Firestore）
```

---

## 6. 状態遷移

### 記事のライフサイクル

```
NEW（RSSから）
  → DEDUP_CHECK（Firestoreルックアップ）
    → DUPLICATE（スキップ）
    → UNIQUE
      → CONTENT_FETCH（go-readability）
        → FAILED（→ RejectedArticle: "content_extraction_failed"）
        → SUCCESS
          → LLM_EVAL（Gemini API）
            → NOT_RELEVANT（→ RejectedArticle: "low_relevance"または"no_topic_match"）
            → RELEVANT
              → SELECTED（上位3〜5）
                → NOTIFIED（→ NotifiedArticle）
              → NOT_SELECTED（上位3〜5に入らない、保存されない）
```

---

## 7. 検証サマリー

| エンティティ | 主要制約 |
|--------|-----------------|
| Config | 1〜10 RSSソース、1〜50興味、重複トピックなし |
| RSSSource | 有効なURL、enabled=true/false、1〜50文字名 |
| InterestTopic | 一意のトピック、priority=high/medium/low、1〜50文字 |
| Article | 一意のURL、100〜50k文字コンテンツ、5〜500文字タイトル |
| ArticleEvaluation | スコア0〜100、50〜200文字要約、関連性がある場合は空でないトピック |
| NotifiedArticle | 有効なDiscord Snowflake ID、ドキュメントIDとしてのURL |
| RejectedArticle | 理由列挙型、評価された場合はスコアあり |
| Discord Embed | 最大10埋め込み、256文字タイトル、4096文字説明 |

---

## 8. Firestoreクォータと制限

**無料枠**:
- 50,000ドキュメント読み取り/日
- 20,000ドキュメント書き込み/日
- 1GBストレージ

**予想される日次使用量**:
- 読み取り：100〜200（重複排除チェック）= クォータの0.4%
- 書き込み：10〜15（3〜5通知 + 5〜10却下）= クォータの0.05%
- ストレージ：約1MB（10,000 URL × 100バイト/ドキュメント）= クォータの0.1%

**結論**: 無料枠の制限内に十分収まる。
