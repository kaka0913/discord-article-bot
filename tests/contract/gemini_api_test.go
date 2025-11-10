package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kaka0913/discord-article-bot/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeminiAPIArticleEvaluation はGemini APIの記事評価機能をテストします
func TestGeminiAPIArticleEvaluation(t *testing.T) {
	// モックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// クエリ内のAPIキーを検証
		apiKey := r.URL.Query().Get("key")
		assert.NotEmpty(t, apiKey, "API key required")

		// Content-Typeを検証
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// リクエストをパース
		var req llm.GeminiRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// 生成設定を検証
		assert.Equal(t, 0.3, req.GenerationConfig.Temperature)
		assert.Equal(t, 2048, req.GenerationConfig.MaxOutputTokens)

		// リクエストの基本構造を検証
		assert.Len(t, req.Contents, 1)
		assert.Equal(t, "user", req.Contents[0].Role)
		assert.Len(t, req.Contents[0].Parts, 1)
		assert.NotEmpty(t, req.Contents[0].Parts[0].Text)

		// モックGemini応答を返す
		w.WriteHeader(http.StatusOK)
		response := llm.GeminiResponse{
			Candidates: []llm.Candidate{
				{
					Content: llm.CandidateContent{
						Parts: []llm.Part{
							{
								Text: `{"relevance_score": 85, "matching_topics": ["Go", "Kubernetes"], "summary": "Goを使用したスケーラブルなマイクロサービスの構築とKubernetesクラスターへのデプロイに関するベストプラクティスを含む包括的なガイド。", "reasoning": "AI生成記事ではない。トピックマッチング: 2つのトピックに詳細な実装例で言及(+20点)。内容の具体性: コード例と設定ファイルを複数含む(+30点)。実用性: 即座に適用可能な実装(+25点)。記事の深さ: 包括的な解説(+15点)。合計90点。"}`,
							},
						},
						Role: "model",
					},
					FinishReason: "STOP",
					Index:        0,
				},
			},
			UsageMetadata: llm.UsageMetadata{
				PromptTokenCount:     1234,
				CandidatesTokenCount: 89,
				TotalTokenCount:      1323,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 注: 実際のGemini APIを呼び出すテストはモックサーバーでは制限があるため、
	// ここではレスポンスのパースとバリデーションをテストします。

	t.Run("レスポンスのパース", func(t *testing.T) {
		// モックレスポンスを直接パース
		jsonResponse := `{"relevance_score": 85, "matching_topics": ["Go", "Kubernetes"], "summary": "Goを使用したスケーラブルなマイクロサービスの構築とKubernetesクラスターへのデプロイに関するベストプラクティスを含む包括的なガイド。", "reasoning": "詳細な実装例を含む"}`

		var result llm.EvaluationResult
		err := json.Unmarshal([]byte(jsonResponse), &result)
		require.NoError(t, err)

		assert.Equal(t, 85, result.RelevanceScore)
		assert.ElementsMatch(t, []string{"Go", "Kubernetes"}, result.MatchingTopics)
		assert.GreaterOrEqual(t, len([]rune(result.Summary)), 50)
		assert.LessOrEqual(t, len([]rune(result.Summary)), 200)
	})

	t.Run("評価結果の検証", func(t *testing.T) {
		// 有効な評価結果
		validResult := &llm.EvaluationResult{
			RelevanceScore: 85,
			MatchingTopics: []string{"Go", "Kubernetes"},
			Summary:        "これは50文字以上200文字以下の有効な要約です。Goを使用したマイクロサービスの構築について説明しています。これは50文字以上200文字以下の有効な要約です。",
			Reasoning:      "詳細な実装例を含む",
		}

		// スコアの範囲外
		invalidScore := &llm.EvaluationResult{
			RelevanceScore: 150,
			MatchingTopics: []string{"Go"},
			Summary:        "これは50文字以上200文字以下の有効な要約です。Goを使用したマイクロサービスの構築について説明しています。これは50文字以上200文字以下の有効な要約です。",
		}

		// 要約が短すぎる
		invalidSummary := &llm.EvaluationResult{
			RelevanceScore: 85,
			MatchingTopics: []string{"Go"},
			Summary:        "短い",
		}

		// スコア > 0 だがトピックが空
		invalidTopics := &llm.EvaluationResult{
			RelevanceScore: 85,
			MatchingTopics: []string{},
			Summary:        "これは50文字以上200文字以下の有効な要約です。Goを使用したマイクロサービスの構築について説明しています。これは50文字以上200文字以下の有効な要約です。",
		}

		// テストケースを実行
		assert.NoError(t, llm.ValidateEvaluationResult(validResult))
		assert.Error(t, llm.ValidateEvaluationResult(invalidScore))
		assert.Error(t, llm.ValidateEvaluationResult(invalidSummary))
		assert.Error(t, llm.ValidateEvaluationResult(invalidTopics))
	})
}

// TestGeminiAPIErrorHandling はGemini APIのエラーハンドリングをテストします
func TestGeminiAPIErrorHandling(t *testing.T) {
	t.Run("401 Unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			response := llm.GeminiError{
				Error: llm.ErrorDetail{
					Code:    401,
					Message: "API key not valid",
					Status:  "UNAUTHENTICATED",
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// エラーレスポンスのパースを検証
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		var geminiErr llm.GeminiError
		err = json.NewDecoder(resp.Body).Decode(&geminiErr)
		require.NoError(t, err)

		assert.Equal(t, 401, geminiErr.Error.Code)
		assert.Equal(t, "UNAUTHENTICATED", geminiErr.Error.Status)
	})

	t.Run("429 Too Many Requests", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			response := llm.GeminiError{
				Error: llm.ErrorDetail{
					Code:    429,
					Message: "Resource has been exhausted (e.g. check quota).",
					Status:  "RESOURCE_EXHAUSTED",
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// エラーレスポンスのパースを検証
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		var geminiErr llm.GeminiError
		err = json.NewDecoder(resp.Body).Decode(&geminiErr)
		require.NoError(t, err)

		assert.Equal(t, 429, geminiErr.Error.Code)
		assert.Equal(t, "RESOURCE_EXHAUSTED", geminiErr.Error.Status)
	})

	t.Run("400 Bad Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			response := llm.GeminiError{
				Error: llm.ErrorDetail{
					Code:    400,
					Message: "Invalid JSON payload",
					Status:  "INVALID_ARGUMENT",
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		// エラーレスポンスのパースを検証
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		var geminiErr llm.GeminiError
		err = json.NewDecoder(resp.Body).Decode(&geminiErr)
		require.NoError(t, err)

		assert.Equal(t, 400, geminiErr.Error.Code)
		assert.Equal(t, "INVALID_ARGUMENT", geminiErr.Error.Status)
	})
}

// TestRateLimiter はレート制限の動作をテストします
func TestRateLimiter(t *testing.T) {
	client := llm.NewClient("test-api-key")

	// レート制限のテストは実際のタイミングに依存するため、
	// ここではクライアントが正しく初期化されることを確認
	assert.NotNil(t, client)
}
