// Package llm はGemini API統合を提供します
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	// GeminiAPIURL はGemini Flash APIのエンドポイント
	GeminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"

	// Temperature はGemini APIの温度パラメータ（一貫性のために低く設定）
	Temperature = 0.3

	// MaxOutputTokens は応答の最大トークン数
	MaxOutputTokens = 500

	// RequestsPerMinute はGemini APIの無料枠のレート制限（15 RPM）
	RequestsPerMinute = 15

	// RequestInterval は1リクエストあたりの平均間隔（4秒）
	RequestInterval = 4 * time.Second
)

// Client はGemini APIクライアントを表します
type Client struct {
	apiKey     string
	httpClient *http.Client
	limiter    *rate.Limiter
}

// NewClient は新しいGemini APIクライアントを作成します
func NewClient(apiKey string) *Client {
	// 15 RPM = 4秒に1リクエスト、バースト1を許可
	limiter := rate.NewLimiter(rate.Every(RequestInterval), 1)

	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		limiter:    limiter,
	}
}

// GeminiRequest はGemini APIへのリクエストを表します
type GeminiRequest struct {
	Contents         []Content         `json:"contents"`
	GenerationConfig GenerationConfig  `json:"generationConfig"`
}

// Content はリクエストのコンテンツを表します
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part はコンテンツの一部を表します
type Part struct {
	Text string `json:"text"`
}

// GenerationConfig は生成設定を表します
type GenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	MaxOutputTokens  int     `json:"maxOutputTokens"`
	ResponseMimeType string  `json:"responseMimeType"`
}

// GeminiResponse はGemini APIからの応答を表します
type GeminiResponse struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

// Candidate は応答の候補を表します
type Candidate struct {
	Content      CandidateContent `json:"content"`
	FinishReason string           `json:"finishReason"`
	Index        int              `json:"index"`
}

// CandidateContent は候補のコンテンツを表します
type CandidateContent struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

// UsageMetadata はトークン使用量のメタデータを表します
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GeminiError はGemini APIからのエラー応答を表します
type GeminiError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail はエラーの詳細を表します
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// GenerateContent はGemini APIにリクエストを送信してコンテンツを生成します
func (c *Client) GenerateContent(ctx context.Context, prompt string) (*GeminiResponse, error) {
	// レート制限を適用（次のトークンが利用可能になるまで待機）
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// リクエストボディを構築
	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			Temperature:      Temperature,
			MaxOutputTokens:  MaxOutputTokens,
			ResponseMimeType: "application/json",
		},
	}

	// JSONにエンコード
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// HTTPリクエストを作成
	url := fmt.Sprintf("%s?key=%s", GeminiAPIURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// エラー応答の処理
	if resp.StatusCode != http.StatusOK {
		var geminiErr GeminiError
		if err := json.Unmarshal(body, &geminiErr); err != nil {
			return nil, fmt.Errorf("unexpected error response (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("gemini api error (status %d): %s - %s",
			resp.StatusCode, geminiErr.Error.Status, geminiErr.Error.Message)
	}

	// 成功応答をパース
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 応答の検証
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	if geminiResp.Candidates[0].FinishReason != "STOP" {
		return nil, fmt.Errorf("unexpected finish reason: %s", geminiResp.Candidates[0].FinishReason)
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no parts in candidate content")
	}

	return &geminiResp, nil
}
