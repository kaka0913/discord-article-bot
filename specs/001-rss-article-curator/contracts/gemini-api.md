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
          "text": "あなたは技術コンテンツキュレーションの専門家です。以下の記事を次のトピックとの関連性について評価してください: [\"Go\", \"Kubernetes\", \"Microservices\"]\n\n記事タイトル: Building Microservices with Go and Kubernetes\n記事内容: In this comprehensive guide, we explore best practices for building scalable microservices using the Go programming language and deploying them on Kubernetes clusters. We cover service mesh patterns, observability, and deployment strategies with actual code examples and deployment manifests...\n\nJSON形式で評価を提供してください:\n{\n  \"relevance_score\": <0-100の整数>,\n  \"matching_topics\": [<一致するトピック名の配列>],\n  \"summary\": \"<50-200文字の要約>\",\n  \"reasoning\": \"<スコアの簡単な説明>\"\n}\n\nスコアリング基準（加算方式、最大100点）:\n\n【AI生成記事の判定】（必須チェック）\n- 人間による執筆と判断: 継続して評価\n- AI生成記事の可能性が高い: 即座に0点を返す\n\n【トピックマッチング】（最大30点）\n- 3つ以上のトピックに詳細な実装例で言及: +30点\n- 2つのトピックに詳細な実装例で言及: +20点\n- 1つのトピックに詳細な実装例で言及: +15点\n\n【内容の具体性】（最大30点）\n- 実際のコード例・コマンド・設定ファイルを複数含む: +30点\n- 実装方法の詳細な手順とコード例を含む: +25点\n\n【実用性】（最大25点）\n- 実際のプロジェクトで即座に適用可能な実装: +25点\n- ステップバイステップのチュートリアル: +20点\n\n【記事の深さ】（最大15点）\n- 包括的で詳細な解説（実質2000文字以上）: +15点\n- 中程度の詳細な解説（実質1000-2000文字）: +10点"
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
            "text": "{\"relevance_score\": 95, \"matching_topics\": [\"Go\", \"Kubernetes\", \"Microservices\"], \"summary\": \"Goを使用したスケーラブルなマイクロサービスの構築とKubernetesクラスターへのデプロイに関するベストプラクティスを含む包括的なガイド。\", \"reasoning\": \"記事は3つのトピックすべてを実践的な例とデプロイ戦略を含めて深くカバーしています。\"}"
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
  "summary": "Goを使用したスケーラブルなマイクロサービスの構築とKubernetesクラスターへのデプロイに関するベストプラクティスを含む包括的なガイド。",
  "reasoning": "AI生成記事ではない（人間の経験に基づく執筆）。トピックマッチング: 3つのトピックに詳細な実装例で言及(+30点)。内容の具体性: コード例と設定ファイルを複数含む(+30点)。実用性: 即座に適用可能な実装(+25点)。記事の深さ: 2000文字以上の包括的な解説(+15点)。合計100点中95点。"
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
                                "text": `{"relevance_score": 85, "matching_topics": ["Go", "Kubernetes"], "summary": "マイクロサービス構築の包括的なガイド...", "reasoning": "トピックを深くカバーしています。"}`,
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
あなたは技術コンテンツキュレーションの専門家です。以下の記事を次のトピックとの関連性について評価してください: {TOPICS}

記事タイトル: {TITLE}
記事内容: {CONTENT}

JSON形式で評価を提供してください:
{
  "relevance_score": <0-100の整数>,
  "matching_topics": [<{TOPICS}から一致するトピック名の配列>],
  "summary": "<50-200文字の要約>",
  "reasoning": "<スコアの簡単な説明>"
}

スコアリング基準（加算方式、最大100点）:

【AI生成記事の判定】（必須チェック）
- 人間による執筆と判断: 継続して評価
- AI生成記事の可能性が高い: 即座に0点を返す
  判定基準:
  * 過度に形式的で個性のない文体
  * 具体的な実装や経験の欠如
  * 表面的な情報の羅列のみ
  * 表現が大袈裟で具体的でない

【トピックマッチング】（最大30点）
- 3つ以上のトピックに詳細な実装例で言及: +30点
- 2つのトピックに詳細な実装例で言及: +20点
- 1つのトピックに詳細な実装例で言及: +15点
- 複数トピックに言及するが表面的: +10点
- 1つのトピックに軽く言及: +5点
- トピックに全く言及なし: +0点

【内容の具体性】（最大30点）
- 実際のコード例・コマンド・設定ファイルを複数含む: +30点
- 実装方法の詳細な手順とコード例を含む: +25点
- アーキテクチャ図や設計パターンの具体的な解説: +20点
- ベストプラクティスと理由の説明: +15点
- 概念的な説明と簡単な例: +10点
- 抽象的な概念の説明のみ: +5点

【実用性】（最大25点）
- 実際のプロジェクトで即座に適用可能な実装: +25点
- ステップバイステップのチュートリアル: +20点
- 実務で参考になる設計思想と具体例: +15点
- 参考情報としての価値あり: +10点
- 一般的な情報の紹介のみ: +5点

【記事の深さ】（最大15点）
- 包括的で詳細な解説（実質2000文字以上）: +15点
- 中程度の詳細な解説（実質1000-2000文字）: +10点
- 簡潔だが要点を押さえた解説（実質500-1000文字）: +7点
- 短い紹介記事（実質500文字未満）: +3点

最終スコア = 合計点（最大100点、AI生成判定の場合は0点）

スコア区分の目安:
- 80-100点: 複数トピック + 詳細なコード例 + 即座に実用可能 + 包括的
- 60-79点: 1-2トピック + 具体的な実装 + 実用的 + 詳細
- 40-59点: トピック言及 + 概念説明 + やや実用的
- 0-39点: トピック言及なし or 表面的 or AI生成

重要な注意事項:
- matching_topicsには{TOPICS}からのトピックのみを含める（幻覚トピック禁止）
- 要約は簡潔（50-200文字）で主要なポイントを強調すること
- AI生成の疑いがある場合は必ず0点とし、reasoningに判定理由を記載
- 同じトピックへの複数の表面的言及より、1つのトピックへの深い言及を高く評価
```

**設計理由**:
- **加算方式のスコアリング**: 各評価項目の配点を明確化し、LLMの判断の一貫性を向上
- **AI生成記事の排除**: 質の低いAI生成コンテンツを事前にフィルタリング
- **具体性重視**: コード例や実装の有無など、客観的に判断できる基準を採用
- **トピックの深さ**: 表面的な言及より詳細な実装例を高く評価
- **プロンプト内のJSONスキーマ**: 構造化出力をガイド
- **文字制限**: 要約の簡潔性を強制（50-200文字）
- **幻覚トピック禁止**: LLMがトピックを発明するのを防止
- **透明性のあるreasoning**: 各項目の配点を明記することで、スコアの根拠を明確化

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
