package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kaka0913/discord-article-bot/internal/logging"
)

// Client は Discord Webhook API クライアント
type Client struct {
	webhookURL string
	httpClient *http.Client
	logger     logging.Logger
}

// NewClient は新しいDiscordクライアントを作成
func NewClient(webhookURL string, logger logging.Logger) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Article はキュレーションされた記事を表す
type Article struct {
	Title       string
	Description string
	URL         string
	Relevance   int
	Topics      []string
	Source      string
}

// WebhookPayload はDiscord Webhook APIのリクエストペイロード
type WebhookPayload struct {
	Content string        `json:"content,omitempty"`
	Embeds  []EmbedObject `json:"embeds"`
}

// EmbedObject はDiscord Embedsオブジェクト
type EmbedObject struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	URL         string        `json:"url,omitempty"`
	Color       int           `json:"color,omitempty"`
	Fields      []EmbedField  `json:"fields,omitempty"`
	Footer      *EmbedFooter  `json:"footer,omitempty"`
}

// EmbedField はEmbedsのフィールド
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// EmbedFooter はEmbedsのフッター
type EmbedFooter struct {
	Text string `json:"text"`
}

// WebhookResponse はDiscord Webhook APIのレスポンス
type WebhookResponse struct {
	ID        string          `json:"id"`
	Type      int             `json:"type"`
	Content   string          `json:"content"`
	ChannelID string          `json:"channel_id"`
	Embeds    []EmbedObject   `json:"embeds"`
	Timestamp string          `json:"timestamp"`
}

// ErrorResponse はDiscord APIのエラーレスポンス
type ErrorResponse struct {
	Message    string          `json:"message"`
	Code       int             `json:"code"`
	Errors     json.RawMessage `json:"errors,omitempty"`
	RetryAfter float64         `json:"retry_after,omitempty"`
}

// PostArticles は記事リストをDiscordに投稿
func (c *Client) PostArticles(ctx context.Context, articles []Article, date string) (string, error) {
	// Embedsペイロードをフォーマット
	payload := FormatArticlesPayload(articles, date)

	// ペイロードの検証
	if len(payload.Embeds) > 10 {
		return "", fmt.Errorf("too many embeds: %d (max 10)", len(payload.Embeds))
	}

	// リトライロジック付きで送信
	return c.postWithRetry(ctx, payload, 3)
}

// postWithRetry はリトライロジック付きでWebhookにPOST
func (c *Client) postWithRetry(ctx context.Context, payload WebhookPayload, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 指数バックオフ: 5s, 10s, 20s
			backoff := time.Duration(5*(1<<uint(attempt-1))) * time.Second
			c.logger.Info(fmt.Sprintf("Discord webhookを%v後に再試行します (試行 %d/%d)", backoff, attempt, maxRetries))

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		messageID, err := c.post(ctx, payload)
		if err == nil {
			return messageID, nil
		}

		lastErr = err

		// エラーの種類を判定
		if isRateLimitError(err) {
			// レート制限エラーの場合、retry_after秒待つ
			if retryAfter := extractRetryAfter(err); retryAfter > 0 {
				c.logger.Warn(fmt.Sprintf("レート制限に達しました。%v秒待機します", retryAfter))
				select {
				case <-time.After(time.Duration(retryAfter) * time.Second):
				case <-ctx.Done():
					return "", ctx.Err()
				}
				// レート制限待機はリトライカウントに含めない
				attempt--
				continue
			}
		}

		// 致命的なエラー（404など）の場合はリトライしない
		if isFatalError(err) {
			c.logger.Error("致命的なエラーが発生しました。リトライしません", "error", err)
			return "", err
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// post はWebhookにPOSTリクエストを送信
func (c *Client) post(ctx context.Context, payload WebhookPayload) (string, error) {
	// JSON エンコード
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// リクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み込み
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return "", c.handleErrorResponse(resp.StatusCode, body)
	}

	// 204 No Contentの場合はbodyが空なのでダミーIDを返す
	if resp.StatusCode == http.StatusNoContent || len(body) == 0 {
		c.logger.Info("Discordへのメッセージ送信に成功しました (No Content)")
		return "success-no-content", nil
	}

	// 成功レスポンスをパース
	var webhookResp WebhookResponse
	if err := json.Unmarshal(body, &webhookResp); err != nil {
		return "", fmt.Errorf("failed to parse success response: %w", err)
	}

	c.logger.Info(fmt.Sprintf("Discordへのメッセージ送信に成功しました: %s", webhookResp.ID))
	return webhookResp.ID, nil
}

// handleErrorResponse はエラーレスポンスを処理
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// パース失敗時も DiscordAPIError を返す（型安全性のため）
		return &DiscordAPIError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("failed to parse error response: %s", string(body)),
		}
	}

	return &DiscordAPIError{
		StatusCode: statusCode,
		Code:       errResp.Code,
		Message:    errResp.Message,
		Errors:     errResp.Errors,
		RetryAfter: errResp.RetryAfter,
	}
}

// DiscordAPIError はDiscord APIのエラー
type DiscordAPIError struct {
	StatusCode int
	Code       int
	Message    string
	Errors     json.RawMessage
	RetryAfter float64
}

func (e *DiscordAPIError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("Discord API error (HTTP %d, code %d): %s, details: %s",
			e.StatusCode, e.Code, e.Message, string(e.Errors))
	}
	return fmt.Sprintf("Discord API error (HTTP %d, code %d): %s",
		e.StatusCode, e.Code, e.Message)
}

// isRateLimitError はレート制限エラーかどうかをチェック
func isRateLimitError(err error) bool {
	if apiErr, ok := err.(*DiscordAPIError); ok {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// isFatalError は致命的なエラーかどうかをチェック（リトライ不要）
func isFatalError(err error) bool {
	if apiErr, ok := err.(*DiscordAPIError); ok {
		// 404 Not Found (無効なwebhook)
		if apiErr.StatusCode == http.StatusNotFound {
			return true
		}
		// 400 Bad Request (検証エラー)
		if apiErr.StatusCode == http.StatusBadRequest {
			return true
		}
	}
	return false
}

// extractRetryAfter はエラーからretry_after値を抽出
func extractRetryAfter(err error) float64 {
	if apiErr, ok := err.(*DiscordAPIError); ok {
		return apiErr.RetryAfter
	}
	return 0
}
