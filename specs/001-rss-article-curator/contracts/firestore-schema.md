# Firestoreスキーマ契約

**データベース**: Google Cloud Firestore（ネイティブモード）
**バージョン**: v1
**ドキュメント**: https://cloud.google.com/firestore/docs

## 概要

この契約は、記事の重複排除追跡のためのFirestoreデータベーススキーマを定義します。システムは2つのコレクションを使用します：`notified_articles`（Discordに投稿された記事）と`rejected_articles`（関連性がないと評価された記事）。

---

## コレクション

### 1. notified_articles

**目的**: Discordに正常に投稿された記事を追跡して重複通知を防止

**ドキュメントID形式**: URL安全な記事URL（例：`https:--dev.to-example-post`）

**スキーマ**:
```json
{
  "notified_at": "timestamp",
  "discord_message_id": "string",
  "article_title": "string",
  "relevance_score": "number"
}
```

**フィールドの説明**:
- `notified_at`（timestamp、必須）：記事がDiscordに投稿された日時
- `discord_message_id`（string、必須）：Discord Snowflake ID（17〜19桁）
- `article_title`（string、必須）：記事のタイトル（デバッグ/ログ用）
- `relevance_score`（number、必須）：LLM関連性スコア（0〜100）

**インデックス**:
- プライマリ：ドキュメントID（自動）
- セカンダリ：`notified_at`（TTLクエリと日付範囲検索用）

**ドキュメント例**:
```
コレクション: notified_articles
ドキュメントID: https:--dev.to-example-building-microservices
{
  "notified_at": Timestamp(2025-10-27T09:15:00Z),
  "discord_message_id": "1234567890123456789",
  "article_title": "Building Microservices with Go and Kubernetes",
  "relevance_score": 95
}
```

---

### 2. rejected_articles

**目的**: LLMによる再評価を避けるために、興味がないと評価された記事を追跡

**ドキュメントID形式**: URL安全な記事URL（例：`https:--dev.to-example-irrelevant`）

**スキーマ**:
```json
{
  "evaluated_at": "timestamp",
  "reason": "string",
  "relevance_score": "number | null"
}
```

**フィールドの説明**:
- `evaluated_at`（timestamp、必須）：記事がLLMによって評価された日時
- `reason`（string、必須）：却下理由の列挙型："low_relevance" | "no_topic_match" | "content_extraction_failed"
- `relevance_score`（number、オプション）：評価された場合はLLMスコア、コンテンツ抽出が失敗した場合はnull

**理由の列挙値**:
- `low_relevance`: LLMが記事を評価したがスコアがmin_relevance_scoreしきい値未満
- `no_topic_match`: LLMがユーザーの興味から一致するトピックを見つけられなかった
- `content_extraction_failed`: go-readabilityが記事テキストの抽出に失敗（404、ペイウォール、タイムアウト）

**インデックス**:
- プライマリ：ドキュメントID（自動）
- セカンダリ：`evaluated_at`（TTLクエリ用）

**ドキュメント例**:
```
コレクション: rejected_articles

# 低い関連性スコア
ドキュメントID: https:--dev.to-example-intro-to-javascript
{
  "evaluated_at": Timestamp(2025-10-27T09:05:00Z),
  "reason": "low_relevance",
  "relevance_score": 35
}

# コンテンツ抽出失敗（ペイウォール）
ドキュメントID: https:--medium.com-example-paywalled-article
{
  "evaluated_at": Timestamp(2025-10-27T09:10:00Z),
  "reason": "content_extraction_failed",
  "relevance_score": null
}
```

---

## 操作

### 記事が通知済みか確認（重複排除）

**操作**: IDによるドキュメント`Get`

**Go例**:
```go
docRef := firestoreClient.Collection("notified_articles").Doc(articleURL)
doc, err := docRef.Get(ctx)
if err != nil {
    if status.Code(err) == codes.NotFound {
        // 記事はまだ通知されていない
        return false, nil
    }
    return false, err // Firestoreエラー
}
// 記事は既に通知済み
return true, nil
```

**パフォーマンス**: O(1)検索（ドキュメントIDでインデックス化）

**コスト**: チェックごとに1ドキュメント読み取り

---

### 記事が却下されたか確認（再評価を回避）

**操作**: IDによるドキュメント`Get`

**Go例**:
```go
docRef := firestoreClient.Collection("rejected_articles").Doc(articleURL)
doc, err := docRef.Get(ctx)
if err != nil {
    if status.Code(err) == codes.NotFound {
        // 記事はまだ却下されていない
        return false, nil
    }
    return false, err // Firestoreエラー
}
// 記事は既に却下済み
return true, nil
```

**パフォーマンス**: O(1)検索（ドキュメントIDでインデックス化）

**コスト**: チェックごとに1ドキュメント読み取り

---

### 通知済み記事を保存

**操作**: マージ付きドキュメント`Set`

**Go例**:
```go
docRef := firestoreClient.Collection("notified_articles").Doc(articleURL)
_, err := docRef.Set(ctx, map[string]interface{}{
    "notified_at":        firestore.ServerTimestamp,
    "discord_message_id": messageID,
    "article_title":      article.Title,
    "relevance_score":    evaluation.RelevanceScore,
})
if err != nil {
    return err // Firestoreエラー
}
```

**パフォーマンス**: O(1)書き込み（ドキュメントIDでインデックス化）

**コスト**: 記事ごとに1ドキュメント書き込み

**べき等性**: `Set`を使用（`Create`ではなく）することで、ドキュメントが存在する場合でもエラーなしで再試行可能

---

### 却下された記事を保存

**操作**: マージ付きドキュメント`Set`

**Go例**:
```go
docRef := firestoreClient.Collection("rejected_articles").Doc(articleURL)
data := map[string]interface{}{
    "evaluated_at": firestore.ServerTimestamp,
    "reason":       reason, // "low_relevance" | "no_topic_match" | "content_extraction_failed"
}
if relevanceScore != nil {
    data["relevance_score"] = *relevanceScore
} else {
    data["relevance_score"] = nil
}
_, err := docRef.Set(ctx, data)
if err != nil {
    return err // Firestoreエラー
}
```

**パフォーマンス**: O(1)書き込み（ドキュメントIDでインデックス化）

**コスト**: 却下された記事ごとに1ドキュメント書き込み

---

## バッチ操作

### 複数記事の確認（バッチGet）

**操作**: バッチあたり最大500ドキュメントの`GetAll`

**Go例**:
```go
var docRefs []*firestore.DocumentRef
for _, url := range articleURLs {
    docRefs = append(docRefs, firestoreClient.Collection("notified_articles").Doc(url))
}

docs, err := firestoreClient.GetAll(ctx, docRefs)
if err != nil {
    return nil, err
}

notifiedURLs := make(map[string]bool)
for i, doc := range docs {
    if doc.Exists() {
        notifiedURLs[articleURLs[i]] = true
    }
}
```

**パフォーマンス**: 単一ラウンドトリップでO(n)検索（最大500ドキュメント/バッチ）

**コスト**: nドキュメント読み取り（個別Getと同じだが高速）

**ユースケース**: RSSフィードからの100〜200記事を1〜2バッチリクエストでチェック

---

## セキュリティルール

**Firestoreセキュリティルール**（firestore.rules）:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // すべてのクライアントアクセスを拒否（サービスアカウント経由のバックエンドのみ）
    match /{document=**} {
      allow read, write: if false;
    }
  }
}
```

**理由**:
- Cloud Functionsはサービスアカウント資格情報を使用（自動、ルール不要）
- Web/モバイルクライアントはFirestoreにアクセスしない（バックエンドのみアーキテクチャ）
- すべてのクライアントアクセスを拒否することで、誤った公開露出を防止

---

## インデックス

### 自動インデックス

Firestoreは以下の単一フィールドインデックスを自動的に作成します:
- ドキュメントID（重複排除検索に使用）
- `notified_at`（timestampフィールド）
- `evaluated_at`（timestampフィールド）

### 複合インデックス

現在のユースケースでは不要。将来のクエリでフィルタリングが必要な場合は以下で作成:

```yaml
# firestore.indexes.json
{
  "indexes": [
    {
      "collectionGroup": "notified_articles",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "relevance_score", "order": "DESCENDING" },
        { "fieldPath": "notified_at", "order": "DESCENDING" }
      ]
    }
  ]
}
```

---

## TTL（Time-To-Live）ポリシー

### オプション：古いドキュメントの自動削除

**理由**: ストレージクォータ（1GB無料枠）を節約し、混雑を減らす

**実装**: Cloud Functions定期タスク（キュレーターとは別）

```go
// 90日以上前の通知済み記事を削除
cutoff := time.Now().AddDate(0, 0, -90)
query := firestoreClient.Collection("notified_articles").Where("notified_at", "<", cutoff)
docs, _ := query.Documents(ctx).GetAll()
for _, doc := range docs {
    doc.Ref.Delete(ctx)
}

// 30日以上前の却下された記事を削除（再評価を許可）
cutoff = time.Now().AddDate(0, 0, -30)
query = firestoreClient.Collection("rejected_articles").Where("evaluated_at", "<", cutoff)
docs, _ = query.Documents(ctx).GetAll()
for _, doc := range docs {
    doc.Ref.Delete(ctx)
}
```

**スケジュール**: 週次（日曜日午前3時JST、Cloud Scheduler経由）

**コスト影響**: 最小限（10〜50削除/週 = 書き込みクォータの0.25%）

---

## エラー処理

### Firestoreエラーコード

| コード | 説明 | アクション |
|------|-------------|--------|
| `NotFound` | ドキュメントが存在しない | 想定内（記事が初めて）、続行 |
| `PermissionDenied` | サービスアカウントにIAMロールがない | 致命的：IAM権限を確認し、関数を終了 |
| `Unavailable` | Firestoreが一時的にダウン | 指数バックオフで再試行（5秒、15秒、45秒） |
| `DeadlineExceeded` | リクエストタイムアウト | 指数バックオフで再試行 |
| `ResourceExhausted` | クォータ超過 | エラーをログに記録し、処理を停止（無料枠では発生しないはず） |
| `AlreadyExists` | ドキュメントが存在（Createのみ） | べき等性のため`Create`の代わりに`Set`を使用 |

### Goエラー処理例

```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

doc, err := firestoreClient.Collection("notified_articles").Doc(url).Get(ctx)
if err != nil {
    switch status.Code(err) {
    case codes.NotFound:
        // 想定内：記事はまだ通知されていない
        return false, nil
    case codes.Unavailable, codes.DeadlineExceeded:
        // 一時的エラー：再試行
        time.Sleep(5 * time.Second)
        return checkIfNotified(ctx, url) // 再試行
    case codes.PermissionDenied:
        // 致命的：サービスアカウントIAM問題
        log.Fatalf("Firestore permission denied: %v", err)
    default:
        // 予期しないエラー：ログを記録して失敗
        return false, fmt.Errorf("firestore error: %w", err)
    }
}
```

---

**注記**: Firestoreの統合テストは省略します。本番環境での動作検証のみを行います。

---

## 監視

### 追跡するメトリクス

- `firestore_reads_total{collection}`（カウンター）：コレクション別のドキュメント読み取り
- `firestore_writes_total{collection}`（カウンター）：コレクション別のドキュメント書き込み
- `firestore_errors_total{code}`（カウンター）：ステータスコード別のエラー
- `firestore_latency_seconds{operation}`（ヒストグラム）：リクエスト時間

### アラート

- `firestore_errors_total{code="PermissionDenied"} > 0`：IAM問題（重大）
- `firestore_errors_total{code="ResourceExhausted"} > 0`：クォータ超過（警告）
- `firestore_reads_total > 40000/day`：無料枠制限に近づいている（情報）
- `firestore_writes_total > 15000/day`：無料枠制限に近づいている（情報）

---

## クォータと制限

**無料枠**:
- 50,000ドキュメント読み取り/日
- 20,000ドキュメント書き込み/日
- 20,000ドキュメント削除/日
- 1GBストレージ

**予想される日次使用量**:
- 読み取り：100〜200（重複排除チェック）+ 100〜200（却下チェック）= 合計200〜400
- 書き込み：3〜5（通知済み）+ 5〜10（却下）= 合計8〜15
- ストレージ：約1MB（10,000 URL × 100バイト/ドキュメント）

**マージン**: 無料枠の99%以上の余裕（制限を大幅に下回る）
