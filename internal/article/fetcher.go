package article

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kaka0913/discord-article-bot/internal/errors"
	"github.com/kaka0913/discord-article-bot/internal/logging"
)

// Fetcher は記事のHTMLコンテンツの取得を担当する
type Fetcher struct {
	client  *http.Client
	timeout time.Duration
}

// NewFetcher は新しいFetcherインスタンスを作成する
func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Fetch は指定されたURLから記事のHTMLコンテンツを取得する
// エラーが発生した場合はエラーを返す
func (f *Fetcher) Fetch(ctx context.Context, url string) (string, error) {
	logger := logging.FromContext(ctx)
	logger.Info("記事HTMLを取得中", "url", url)

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", errors.NewValidationError("HTTPリクエストの作成に失敗", err)
	}

	// User-Agentヘッダーを設定（一部のサイトではUser-Agentが必要）
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; discord-article-bot/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8")

	// HTTPリクエストを実行
	resp, err := f.client.Do(req)
	if err != nil {
		return "", errors.NewArticleError("記事HTMLの取得に失敗", err)
	}
	defer resp.Body.Close()

	// ステータスコードを確認
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(
			errors.ErrorTypeArticle,
			fmt.Sprintf("記事HTMLの取得に失敗: HTTPステータス %d", resp.StatusCode),
		)
	}

	// Content-Typeがtext/htmlかどうかを確認（警告のみ）
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !containsHTML(contentType) {
		logger.Warn("Content-Typeがtext/htmlではない", "url", url, "contentType", contentType)
	}

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.NewArticleError("レスポンスボディの読み取りに失敗", err)
	}

	// HTMLコンテンツのサイズを検証
	if len(body) == 0 {
		return "", errors.New(errors.ErrorTypeArticle, "記事HTMLが空")
	}

	// 最大サイズを制限（10MB）
	const maxSize = 10 * 1024 * 1024
	if len(body) > maxSize {
		return "", errors.New(
			errors.ErrorTypeArticle,
			fmt.Sprintf("記事HTMLが大きすぎる: %d bytes (最大 %d bytes)", len(body), maxSize),
		)
	}

	logger.Info("記事HTMLの取得に成功", "url", url, "size", len(body))
	return string(body), nil
}

// containsHTML はContent-Typeヘッダーにtext/htmlが含まれているかを確認する
func containsHTML(contentType string) bool {
	return len(contentType) >= 9 && contentType[:9] == "text/html"
}
