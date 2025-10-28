# Discord Webhook API契約

**API**: Discord Webhook API
**バージョン**: v10
**ドキュメント**: https://discord.com/developers/docs/resources/webhook

## 概要

この契約は、キュレーションされた記事ダイジェストを投稿するためのDiscord Webhook API連携を定義します。システムはリッチメッセージフォーマットのためにWebhook Embedsを使用します。

---

## エンドポイント

**POST** `https://discord.com/api/webhooks/{webhook.id}/{webhook.token}`

**認証**: なし（URL内のwebhookトークンが認可を提供）

**レート制限**:
- 30リクエスト/分（webhook単位）
- 5リクエスト/秒のバースト（その後レート制限）

---

## リクエストペイロード

### Content Type
```
Content-Type: application/json
```

### スキーマ

```json
{
  "content": "string (max 2000 chars, optional)",
  "embeds": [
    {
      "title": "string (max 256 chars, required)",
      "description": "string (max 4096 chars, required)",
      "url": "string (valid URL, optional)",
      "color": "integer (0-16777215, optional)",
      "fields": [
        {
          "name": "string (max 256 chars, required)",
          "value": "string (max 1024 chars, required)",
          "inline": "boolean (optional)"
        }
      ],
      "footer": {
        "text": "string (max 2048 chars, required)"
      }
    }
  ]
}
```

### 制約

- メッセージあたり最大10埋め込み
- メッセージの合計サイズ < 6000文字（content + embedsの合計）
- `color`はRGB 16進数の10進数表現（例：#58A5EF = 5814783）

---

## リクエスト例

### 3件の記事を含む日次ダイジェスト

```json
{
  "content": "📰 Daily Tech Article Digest - 2025-10-27",
  "embeds": [
    {
      "title": "Building Microservices with Go and Kubernetes",
      "description": "A comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters with best practices.",
      "url": "https://dev.to/example/building-microservices",
      "color": 5814783,
      "fields": [
        {
          "name": "Relevance",
          "value": "95/100",
          "inline": true
        },
        {
          "name": "Topics",
          "value": "Go, Kubernetes",
          "inline": true
        }
      ],
      "footer": {
        "text": "Source: Dev.to"
      }
    },
    {
      "title": "WebAssembly Performance Optimization Tips",
      "description": "Learn advanced techniques for optimizing WebAssembly modules to achieve near-native performance in web browsers.",
      "url": "https://zenn.dev/example/wasm-optimization",
      "color": 5814783,
      "fields": [
        {
          "name": "Relevance",
          "value": "88/100",
          "inline": true
        },
        {
          "name": "Topics",
          "value": "WebAssembly, Rust",
          "inline": true
        }
      ],
      "footer": {
        "text": "Source: Zenn"
      }
    },
    {
      "title": "Rust Async Runtime Internals",
      "description": "Deep dive into Tokio runtime architecture and how async/await works under the hood in Rust applications.",
      "url": "https://hashnode.dev/example/rust-async",
      "color": 5814783,
      "fields": [
        {
          "name": "Relevance",
          "value": "82/100",
          "inline": true
        },
        {
          "name": "Topics",
          "value": "Rust, Async",
          "inline": true
        }
      ],
      "footer": {
        "text": "Source: Hashnode"
      }
    }
  ]
}
```

---

## 応答

### 成功（200 OK）

```json
{
  "id": "1234567890123456789",
  "type": 0,
  "content": "📰 Daily Tech Article Digest - 2025-10-27",
  "channel_id": "987654321098765432",
  "embeds": [...],
  "timestamp": "2025-10-27T00:15:00.000Z"
}
```

**主要フィールド**:
- `id`: 投稿されたメッセージのDiscord Snowflake ID（NotifiedArticleに保存）
- `timestamp`: Discordがメッセージを受信した日時

### エラー応答

#### 400 Bad Request
```json
{
  "code": 50035,
  "message": "Invalid Form Body",
  "errors": {
    "embeds": {
      "0": {
        "title": {
          "_errors": [
            {
              "code": "BASE_TYPE_MAX_LENGTH",
              "message": "Must be 256 or fewer in length."
            }
          ]
        }
      }
    }
  }
}
```

**原因**:
- タイトル > 256文字
- 説明 > 4096文字
- 10個以上の埋め込み
- 無効なURL形式

**処理**: エラーをログに記録し、記事をスキップして残りの記事を続行

#### 404 Not Found
```json
{
  "message": "Unknown Webhook",
  "code": 10015
}
```

**原因**:
- 無効なwebhook ID/トークン
- webhookが削除された

**処理**: 致命的エラー、ログを記録して終了（管理者がSecret Manager内のwebhook URLを修正する必要がある）

#### 429 Too Many Requests
```json
{
  "message": "You are being rate limited.",
  "retry_after": 64.0,
  "global": false
}
```

**原因**:
- 30リクエスト/分の超過
- バースト制限の超過（5リクエスト/秒）

**処理**: `retry_after`秒後に指数バックオフで再試行（1メッセージ/日では発生しないはず）

---

## 契約テスト

### テストケース

1. **有効なEmbeds ペイロード**
   - すべてのフィールドを含む3〜5個の埋め込みを送信
   - 200 OK応答を検証
   - メッセージIDが返されることを検証
   - 埋め込みがDiscordで正しくレンダリングされることを検証

2. **最大埋め込み数（10個）**
   - 10個の埋め込みを送信（エッジケース）
   - 200 OK応答を検証

3. **タイトルが長すぎる（> 256文字）**
   - 257文字のタイトルを持つ埋め込みを送信
   - 400エラー応答を検証
   - エラーメッセージに"title"が含まれることを検証

4. **説明が長すぎる（> 4096文字）**
   - 4097文字の説明を持つ埋め込みを送信
   - 400エラー応答を検証

5. **無効なWebhookトークン**
   - 偽のwebhook URLにリクエストを送信
   - 404エラー応答を検証
   - エラーコード10015を検証

6. **レート制限処理**
   - 60秒間に31リクエストを送信
   - 31番目のリクエストで429応答を検証
   - `retry_after`フィールドが存在することを検証
   - 待機後の再試行成功を検証

### Goテスト例

```go
func TestDiscordWebhookEmbedsPayload(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Content-Typeを検証
        assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

        // リクエストボディをパース
        var payload struct {
            Content string         `json:"content"`
            Embeds  []DiscordEmbed `json:"embeds"`
        }
        json.NewDecoder(r.Body).Decode(&payload)

        // 制約を検証
        assert.LessOrEqual(t, len(payload.Embeds), 10, "Max 10 embeds")
        for _, embed := range payload.Embeds {
            assert.LessOrEqual(t, len(embed.Title), 256, "Title max 256 chars")
            assert.LessOrEqual(t, len(embed.Description), 4096, "Description max 4096 chars")
            assert.True(t, embed.Color >= 0 && embed.Color <= 16777215, "Color 0-16777215")
        }

        // モックDiscord応答を返す
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id": "1234567890123456789",
            "type": 0,
            "content": payload.Content,
            "timestamp": time.Now().Format(time.RFC3339),
        })
    }))
    defer server.Close()

    // Discordクライアントをテスト
    client := NewDiscordClient(server.URL)
    messageID, err := client.PostArticles([]CuratedArticle{...})

    assert.NoError(t, err)
    assert.Equal(t, "1234567890123456789", messageID)
}
```

---

## エラー処理戦略

| エラーコード | HTTPステータス | アクション |
|------------|-------------|--------|
| 50035 | 400 | 検証エラーをログに記録し、不正な埋め込みをスキップして続行 |
| 10015 | 404 | 致命的：無効なwebhook、関数を終了し、管理者に警告 |
| Rate Limit | 429 | `retry_after`秒待機し、最大3回再試行 |
| Timeout | - | 指数バックオフで再試行（5秒、10秒、20秒） |
| Network Error | - | 指数バックオフで再試行（5秒、10秒、20秒） |

---

## 監視

### 追跡するメトリクス

- `discord_webhook_requests_total`（カウンター）：送信された総リクエスト数
- `discord_webhook_errors_total{code}`（カウンター）：ステータスコード別のエラー
- `discord_webhook_latency_seconds`（ヒストグラム）：リクエスト時間
- `discord_messages_posted_total`（カウンター）：成功したメッセージ投稿

### アラート

- `discord_webhook_errors_total{code="404"} > 0`：無効なwebhook（重大）
- `discord_webhook_errors_total{code="429"} > 0`：レート制限に到達（警告）
- `discord_webhook_latency_seconds > 5s`：Discord APIが遅い（警告）
