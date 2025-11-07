package rss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kaka0913/discord-article-bot/internal/errors"
	"github.com/kaka0913/discord-article-bot/internal/logging"
)

// Fetcher はRSSフィードのHTTPリクエストを担当する
type Fetcher struct {
	client *http.Client
}

// NewFetcher は新しいFetcherインスタンスを作成する
func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Fetch は指定されたRSSフィードURLからXMLコンテンツを取得する
// エラーが発生した場合はエラーを返す
func (f *Fetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	logger := logging.FromContext(ctx)
	logger.Info("RSSフィードを取得中", "url", url)

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.NewValidationError("HTTPリクエストの作成に失敗", err)
	}

	// User-Agentヘッダーを設定（一部のサイトではUser-Agentが必要）
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; discord-article-bot/1.0)")
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml")

	// HTTPリクエストを実行
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, errors.NewRSSError("RSSフィードの取得に失敗", err)
	}
	defer resp.Body.Close()

	// ステータスコードを確認
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(
			errors.ErrorTypeRSS,
			fmt.Sprintf("RSSフィードの取得に失敗: HTTPステータス %d", resp.StatusCode),
		)
	}

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewRSSError("レスポンスボディの読み取りに失敗", err)
	}

	logger.Info("RSSフィードの取得に成功", "url", url, "size", len(body))
	return body, nil
}
