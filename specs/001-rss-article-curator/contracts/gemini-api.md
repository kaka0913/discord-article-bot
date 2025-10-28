# Gemini API契約

**API**: Google Generative AI (Gemini)  
**モデル**: gemini-1.5-flashx  
**バージョン**: v1beta  
**ドキュメント**: https://ai.google.dev/api/rest  

## 概要

この契約は、記事の関連性評価と要約生成のためのGemini Flash API連携を定義します。システムは一貫した応答パースのために構造化JSON出力モードを使用します。

---

## エンドポイント

**POST** `https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent`

**認証**: APIキー（クエリパラメータ）
```
?key={GEMINI_API_KEY}
```

**レート制限**（無料枠）:
- 15リクエスト/分（RPM）
- 1500リクエスト/日（RPD）
- 100万トークン/分（TPM）

---

## リクエストペイロード

### Content Type
```
Content-Type: application/json
```

### スキーマ

```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {
          "text": "string (prompt + article content)"
        }
      ]
    }
  ],
  "generationConfig": {
    "temperature": 0.3,
    "maxOutputTokens": 500,
    "responseMimeType": "application/json"
  }
}
```

### フィールドの説明

- `contents`: 会話メッセージの配列（我々のユースケースでは単一のユーザーメッセージ）
- `role`: "user"（マルチターン会話ではないため常に"user"）
- `parts[].text`: 記事コンテンツを含むプロンプト（Flashモデルの最大約50kトークン）
- `temperature`: 0.3（一貫したスコアリングのために低く設定、範囲0-2）
- `maxOutputTokens`: 500（スコア+要約のJSON応答に十分）
- `responseMimeType`: "application/json"（構造化JSON出力を有効化）

---

## リクエスト例

### 記事評価

```json
{
  "contents": [
    {
      "role": "user",
      "parts": [
        {
          "text": "You are an expert tech content curator. Evaluate the following article for relevance to these topics: [\"Go\", \"Kubernetes\", \"Microservices\"]\n\nArticle Title: Building Microservices with Go and Kubernetes\nArticle Content: In this comprehensive guide, we explore best practices for building scalable microservices using the Go programming language and deploying them on Kubernetes clusters. We cover service mesh patterns, observability, and deployment strategies...\n\nProvide your evaluation in JSON format:\n{\n  \"relevance_score\": <0-100 integer>,\n  \"matching_topics\": [<array of matching topic names>],\n  \"summary\": \"<50-200 character summary>\",\n  \"reasoning\": \"<brief explanation>\"\n}\n\nScoring criteria:\n- 80-100: Highly relevant, covers multiple topics in depth\n- 60-79: Relevant, covers at least one topic well\n- 40-59: Partially relevant, mentions topics briefly\n- 0-39: Not relevant, unrelated content"
        }
      ]
    }
  ],
  "generationConfig": {
    "temperature": 0.3,
    "maxOutputTokens": 500,
    "responseMimeType": "application/json"
  }
}
```

---

## 応答

### 成功（200 OK）

```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "text": "{\"relevance_score\": 95, \"matching_topics\": [\"Go\", \"Kubernetes\", \"Microservices\"], \"summary\": \"Comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters with best practices.\", \"reasoning\": \"Article covers all three topics in depth with practical examples and deployment strategies.\"}"
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 1234,
    "candidatesTokenCount": 89,
    "totalTokenCount": 1323
  }
}
```

**主要フィールド**:
- `candidates[0].content.parts[0].text`: JSON文字列（評価を抽出するためにパース）
- `finishReason`: "STOP"（正常完了）、"SAFETY"（ブロック）、"MAX_TOKENS"（制限超過）
- `usageMetadata`: クォータ追跡のためのトークン使用状況

### 応答からパースされたJSON

```json
{
  "relevance_score": 95,
  "matching_topics": ["Go", "Kubernetes", "Microservices"],
  "summary": "Comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters with best practices.",
  "reasoning": "Article covers all three topics in depth with practical examples and deployment strategies."
}
```

---

## エラー応答

### 400 Bad Request（無効なプロンプト）
```json
{
  "error": {
    "code": 400,
    "message": "Invalid JSON payload",
    "status": "INVALID_ARGUMENT"
  }
}
```

**原因**:
- 不正なJSONリクエスト
- 無効なフィールドタイプ
- 必須フィールドの欠落

**処理**: エラーをログに記録し、記事をcontent_extraction_failedとして扱い、スキップ

### 401 Unauthorized
```json
{
  "error": {
    "code": 401,
    "message": "API key not valid",
    "status": "UNAUTHENTICATED"
  }
}
```

**原因**:
- 無効なAPIキー
- 期限切れのAPIキー
- `?key=`パラメータの欠落

**処理**: 致命的エラー、関数を終了し、管理者に警告（Secret Managerの更新が必要）

### 429 Too Many Requests（レート制限）
```json
{
  "error": {
    "code": 429,
    "message": "Resource has been exhausted (e.g. check quota).",
    "status": "RESOURCE_EXHAUSTED",
    "details": [
      {
        "@type": "type.googleapis.com/google.rpc.QuotaFailure",
        "violations": [
          {
            "subject": "requests_per_minute",
            "description": "Quota exceeded for requests per minute."
          }
        ]
      }
    ]
  }
}
```

**原因**:
- 15 RPM（リクエスト/分）の超過
- 1500 RPD（リクエスト/日）の超過
- 100万TPM（トークン/分）の超過

**処理**:
- **RPM**: 60秒待機して再試行（golang.org/x/time/rateを使用して事前に防止）
- **RPD**: 処理を停止し、警告をログに記録（日次クォータに到達、明日再開）
- **TPM**: 60秒待機して再試行（記事サイズのコンテンツでは発生しないはず）

### 500 Internal Server Error
```json
{
  "error": {
    "code": 500,
    "message": "Internal error encountered.",
    "status": "INTERNAL"
  }
}
```

**原因**:
- Gemini APIの一時的な停止
- モデルの過負荷

**処理**: 指数バックオフで再試行（5秒、15秒、45秒）、最大3回再試行

---

## 契約テスト

### テストケース

1. **有効な記事評価**
   - 有効なトピックを含む記事コンテンツを送信
   - 200 OK応答を検証
   - JSON応答に必須フィールドが含まれることを検証
   - `relevance_score`が0-100の整数であることを検証
   - `summary`が50-200文字であることを検証

2. **応答JSONスキーマ検証**
   - `candidates[0].content.parts[0].text`をJSONとしてパース
   - `relevance_score`の型が数値であることを検証
   - `matching_topics`の型が文字列配列であることを検証
   - `summary`の型が文字列であることを検証
   - `reasoning`の型が文字列であることを検証

3. **レート制限処理**
   - 60秒間に16リクエストを送信
   - 16番目のリクエストで429応答を検証
   - エラーに"requests_per_minute"が含まれることを検証
   - 60秒待機後の再試行成功を検証

4. **無効なAPIキー**
   - 偽のAPIキーでリクエストを送信
   - 401応答を検証
   - エラーステータス"UNAUTHENTICATED"を検証

5. **無効なJSONリクエスト**
   - 不正なJSON（必須フィールドの欠落）を送信
   - 400応答を検証
   - エラーステータス"INVALID_ARGUMENT"を検証

6. **セーフティフィルタのトリガー**
   - 有害なコンテンツを含む記事を送信（エッジケース）
   - `finishReason: "SAFETY"`を検証
   - 適切な処理（low_relevanceとして扱う）を検証

### Goテスト例

```go
func TestGeminiAPIArticleEvaluation(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // クエリ内のAPIキーを検証
        assert.NotEmpty(t, r.URL.Query().Get("key"), "API key required")

        // Content-Typeを検証
        assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

        // リクエストをパース
        var req struct {
            Contents []struct {
                Role  string `json:"role"`
                Parts []struct {
                    Text string `json:"text"`
                } `json:"parts"`
            } `json:"contents"`
            GenerationConfig struct {
                Temperature       float64 `json:"temperature"`
                ResponseMimeType  string  `json:"responseMimeType"`
            } `json:"generationConfig"`
        }
        json.NewDecoder(r.Body).Decode(&req)

        // 構造化出力がリクエストされていることを検証
        assert.Equal(t, "application/json", req.GenerationConfig.ResponseMimeType)
        assert.Equal(t, 0.3, req.GenerationConfig.Temperature)

        // モックGemini応答を返す
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "candidates": []map[string]interface{}{
                {
                    "content": map[string]interface{}{
                        "parts": []map[string]interface{}{
                            {
                                "text": `{"relevance_score": 85, "matching_topics": ["Go", "Kubernetes"], "summary": "Comprehensive guide to building microservices...", "reasoning": "Covers topics in depth."}`,
                            },
                        },
                        "role": "model",
                    },
                    "finishReason": "STOP",
                    "index": 0,
                },
            },
            "usageMetadata": map[string]int{
                "promptTokenCount":     1234,
                "candidatesTokenCount": 89,
                "totalTokenCount":      1323,
            },
        })
    }))
    defer server.Close()

    // Geminiクライアントをテスト
    client := NewGeminiClient(server.URL, "test-api-key")
    evaluation, err := client.EvaluateArticle(Article{...}, []string{"Go", "Kubernetes"})

    assert.NoError(t, err)
    assert.Equal(t, 85, evaluation.RelevanceScore)
    assert.ElementsMatch(t, []string{"Go", "Kubernetes"}, evaluation.MatchingTopics)
    assert.GreaterOrEqual(t, len(evaluation.Summary), 50)
    assert.LessOrEqual(t, len(evaluation.Summary), 200)
}
```

---

## レート制限実装

### golang.org/x/time/rate戦略

```go
import "golang.org/x/time/rate"

// レート制限を作成：15リクエスト/分（1リクエスト/4秒）
limiter := rate.NewLimiter(rate.Every(4*time.Second), 1) // バースト1を許可

// 各Gemini API呼び出しの前
err := limiter.Wait(context.Background()) // トークンが利用可能になるまでブロック
if err != nil {
    return err // コンテキストがキャンセルされた
}

// Gemini APIリクエストを実行
response, err := geminiClient.EvaluateArticle(article, interests)
```

**理由**:
- 15 RPM = 平均4秒に1リクエスト
- バースト1でトークンが利用可能な場合は即座にリクエスト可能
- `Wait()`が次のトークンが利用可能になるまで自動的にスリープ（手動再試行ロジック不要）

---

## プロンプトエンジニアリング

### 評価プロンプトテンプレート

```
You are an expert tech content curator. Evaluate the following article for relevance to these topics: {TOPICS}

Article Title: {TITLE}
Article Content: {CONTENT}

Provide your evaluation in JSON format:
{
  "relevance_score": <0-100 integer>,
  "matching_topics": [<array of matching topic names from {TOPICS}>],
  "summary": "<50-200 character summary>",
  "reasoning": "<brief explanation of score>"
}

Scoring criteria:
- 80-100: Highly relevant, covers multiple topics in depth with examples
- 60-79: Relevant, covers at least one topic well with practical content
- 40-59: Partially relevant, mentions topics briefly without depth
- 0-39: Not relevant, unrelated content or only tangential mentions

Important:
- Only include topics from {TOPICS} in matching_topics (no hallucinated topics)
- Summary must be concise (50-200 chars) and highlight key takeaways
- Be strict: generic mentions without substance should score low
```

**設計理由**:
- 明確なスコアリング基準によりLLMの変動を削減
- プロンプト内のJSONスキーマが構造化出力をガイド
- 文字制限により要約の簡潔性を強制
- 「幻覚トピック禁止」によりLLMがトピックを発明するのを防止

---

## エラー処理戦略

| エラーコード | HTTPステータス | アクション |
|------------|-------------|--------|
| INVALID_ARGUMENT | 400 | エラーをログに記録し、content_extraction_failedとして扱い、記事をスキップ |
| UNAUTHENTICATED | 401 | 致命的：無効なAPIキー、関数を終了し、管理者に警告 |
| RESOURCE_EXHAUSTED (RPM) | 429 | 60秒待機して再試行（レート制限により発生しないはず） |
| RESOURCE_EXHAUSTED (RPD) | 429 | 警告をログに記録し、処理を停止（日次クォータに到達） |
| INTERNAL | 500 | 指数バックオフで再試行（5秒、15秒、45秒）、最大3回 |
| finishReason: SAFETY | 200 | 警告をログに記録し、low_relevance（スコア0）として扱う |
| finishReason: MAX_TOKENS | 200 | 警告をログに記録し、low_relevance（切り詰められた応答）として扱う |

---

## 監視

### 追跡するメトリクス

- `gemini_api_requests_total`（カウンター）：API呼び出し総数
- `gemini_api_errors_total{code}`（カウンター）：ステータスコード別のエラー
- `gemini_api_latency_seconds`（ヒストグラム）：リクエスト時間
- `gemini_api_tokens_used_total{type}`（カウンター）：プロンプト/応答トークン
- `gemini_rate_limit_waits_total`（カウンター）：レート制限によるブロック回数

### アラート

- `gemini_api_errors_total{code="401"} > 0`：無効なAPIキー（重大）
- `gemini_api_errors_total{code="429"} > 5`：クォータ枯渇（警告）
- `gemini_api_latency_seconds > 10s`：APIの応答が遅い（警告）
- `gemini_api_tokens_used_total > 1000000/day`：TPM制限に近づいている（情報）
