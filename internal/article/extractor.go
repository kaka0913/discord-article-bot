package article

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-shiori/go-readability"

	"github.com/kaka0913/discord-article-bot/internal/errors"
	"github.com/kaka0913/discord-article-bot/internal/logging"
)

// Extractor は記事の本文抽出を担当する
type Extractor struct {
	minTextLength int
	maxTextLength int
}

// NewExtractor は新しいExtractorインスタンスを作成する
func NewExtractor(minTextLength, maxTextLength int) *Extractor {
	return &Extractor{
		minTextLength: minTextLength,
		maxTextLength: maxTextLength,
	}
}

// Extract はHTMLから記事の本文を抽出する
// go-readabilityを使用して本文を抽出し、テキストのみを返す
func (e *Extractor) Extract(ctx context.Context, htmlContent, articleURL string) (string, error) {
	logger := logging.FromContext(ctx)
	logger.Info("記事本文を抽出中", "url", articleURL)

	// URLを検証してパース
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return "", errors.NewValidationError("記事URLのパースに失敗", err)
	}

	// go-readabilityで本文を抽出
	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		return "", errors.NewArticleError("記事本文の抽出に失敗", err)
	}

	// テキストコンテンツを取得（HTMLタグを削除）
	text := article.TextContent

	// テキストをサニタイズ（前後の空白を削除、連続する空白を1つに）
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")

	// テキストの長さを検証
	textLength := len(text)
	if textLength < e.minTextLength {
		return "", errors.New(
			errors.ErrorTypeArticle,
			fmt.Sprintf("記事本文が短すぎる: %d文字 (最小 %d文字)", textLength, e.minTextLength),
		)
	}

	if textLength > e.maxTextLength {
		logger.Warn("記事本文が長すぎるため切り詰める",
			"url", articleURL,
			"originalLength", textLength,
			"maxLength", e.maxTextLength,
		)
		text = text[:e.maxTextLength]
		textLength = e.maxTextLength
	}

	logger.Info("記事本文の抽出に成功",
		"url", articleURL,
		"title", article.Title,
		"textLength", textLength,
	)

	return text, nil
}

// ExtractWithTitle はHTMLから記事の本文とタイトルを抽出する
func (e *Extractor) ExtractWithTitle(ctx context.Context, htmlContent, articleURL string) (title, text string, err error) {
	logger := logging.FromContext(ctx)
	logger.Info("記事本文とタイトルを抽出中", "url", articleURL)

	// URLを検証してパース
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return "", "", errors.NewValidationError("記事URLのパースに失敗", err)
	}

	// go-readabilityで本文を抽出
	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		return "", "", errors.NewArticleError("記事本文の抽出に失敗", err)
	}

	// タイトルを取得
	title = strings.TrimSpace(article.Title)

	// テキストコンテンツを取得（HTMLタグを削除）
	text = article.TextContent

	// テキストをサニタイズ（前後の空白を削除、連続する空白を1つに）
	text = strings.TrimSpace(text)
	text = strings.Join(strings.Fields(text), " ")

	// テキストの長さを検証
	textLength := len(text)
	if textLength < e.minTextLength {
		return "", "", errors.New(
			errors.ErrorTypeArticle,
			fmt.Sprintf("記事本文が短すぎる: %d文字 (最小 %d文字)", textLength, e.minTextLength),
		)
	}

	if textLength > e.maxTextLength {
		logger.Warn("記事本文が長すぎるため切り詰める",
			"url", articleURL,
			"originalLength", textLength,
			"maxLength", e.maxTextLength,
		)
		text = text[:e.maxTextLength]
		textLength = e.maxTextLength
	}

	logger.Info("記事本文とタイトルの抽出に成功",
		"url", articleURL,
		"title", title,
		"textLength", textLength,
	)

	return title, text, nil
}
