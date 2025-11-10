package rss

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mmcdole/gofeed"

	"github.com/kaka0913/discord-article-bot/internal/errors"
	"github.com/kaka0913/discord-article-bot/internal/logging"
)

const (
	// maxTitleLength は記事タイトルの最大長
	// 500文字以上のタイトルは異常に長いため切り詰める
	maxTitleLength = 500
)

// Article は取得したRSS記事を表す
type Article struct {
	Title         string    // 記事のタイトル
	URL           string    // 記事のURL
	PublishedDate time.Time // 公開日時
	SourceFeed    string    // ソースフィード名
	FetchedAt     time.Time // 取得日時
}

// Parser はRSSフィードのXMLをパースする
type Parser struct {
	parser *gofeed.Parser
}

// NewParser は新しいParserインスタンスを作成する
func NewParser() *Parser {
	return &Parser{
		parser: gofeed.NewParser(),
	}
}

// Parse はRSSフィードのXMLをパースしてArticleのリストを返す
// RSS 2.0とAtomフィードの両方に対応
func (p *Parser) Parse(ctx context.Context, xmlData []byte, sourceFeedName string) ([]Article, error) {
	logger := logging.FromContext(ctx)
	logger.Info("RSSフィードをパース中", "source", sourceFeedName)

	// gofeedを使用してフィードをパース
	feed, err := p.parser.ParseString(string(xmlData))
	if err != nil {
		return nil, errors.NewRSSError("RSSフィードのパースに失敗", err)
	}

	// フィードの基本情報をログに記録
	logger.Info("RSSフィードのパース成功",
		"source", sourceFeedName,
		"title", feed.Title,
		"type", feed.FeedType,
		"itemCount", len(feed.Items),
	)

	// 記事のリストを作成
	articles := make([]Article, 0, len(feed.Items))
	now := time.Now()

	for _, item := range feed.Items {
		// 必須フィールドの検証
		if item.Link == "" {
			logger.Warn("記事のリンクが空のためスキップ", "title", item.Title)
			continue
		}

		if item.Title == "" {
			logger.Warn("記事のタイトルが空のためスキップ", "link", item.Link)
			continue
		}

		// タイトルをサニタイズ（前後の空白を削除）
		title := strings.TrimSpace(item.Title)
		if len(title) > maxTitleLength {
			title = truncateUTF8(title, maxTitleLength)
		}

		// 公開日時を取得（存在しない場合は現在時刻を使用）
		publishedDate := now
		if item.PublishedParsed != nil {
			publishedDate = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedDate = *item.UpdatedParsed
		}

		article := Article{
			Title:         title,
			URL:           item.Link,
			PublishedDate: publishedDate,
			SourceFeed:    sourceFeedName,
			FetchedAt:     now,
		}

		articles = append(articles, article)
	}

	logger.Info("記事のパース完了",
		"source", sourceFeedName,
		"totalItems", len(feed.Items),
		"validArticles", len(articles),
	)

	return articles, nil
}

// truncateUTF8 はUTF-8文字列を安全に切り詰める
// マルチバイト文字の途中で切断されることを防ぐ
func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	// maxBytes以下で最大のUTF-8安全な位置を見つける
	for i := maxBytes; i > 0; i-- {
		if utf8.RuneStart(s[i]) {
			return s[:i]
		}
	}

	// 最初の1文字も切り詰められない場合は空文字列を返す
	return ""
}
