package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// バリデーション用の定数
const (
	MinTitleLength   = 5
	MaxTitleLength   = 500
	MinContentLength = 100
	MaxContentLength = 50000
	MinSummaryLength = 50
	MaxSummaryLength = 200
)

// Validator は設定を検証するためのインターフェース
type Validator interface {
	Validate(config *Config) error
}

// configValidator は設定バリデーターの実装
type configValidator struct {
	validate *validator.Validate
}

// NewValidator は新しい設定バリデーターを作成します
func NewValidator() Validator {
	return &configValidator{
		validate: validator.New(),
	}
}

// Validate は設定の妥当性を検証します
func (v *configValidator) Validate(config *Config) error {
	// 構造体のバリデーションタグを検証
	if err := v.validate.Struct(config); err != nil {
		return fmt.Errorf("設定の検証に失敗しました: %w", err)
	}

	// カスタムバリデーション: 少なくとも1つのRSSソースが有効である必要がある
	hasEnabledSource := false
	for _, source := range config.RSSSources {
		if source.Enabled {
			hasEnabledSource = true
			break
		}
	}
	if !hasEnabledSource {
		return fmt.Errorf("少なくとも1つのRSSソースを有効にする必要があります")
	}

	// カスタムバリデーション: 興味トピックに重複がないか確認
	topicMap := make(map[string]bool)
	for _, interest := range config.Interests {
		if topicMap[interest.Topic] {
			return fmt.Errorf("重複する興味トピックが見つかりました: %s", interest.Topic)
		}
		topicMap[interest.Topic] = true
	}

	// カスタムバリデーション: MinArticlesがMaxArticles以下であることを確認
	if config.NotificationSettings.MinArticles > config.NotificationSettings.MaxArticles {
		return fmt.Errorf("min_articles (%d) はmax_articles (%d) 以下である必要があります",
			config.NotificationSettings.MinArticles,
			config.NotificationSettings.MaxArticles)
	}

	return nil
}

// ValidateArticle は記事の妥当性を検証します
func ValidateArticle(article *Article) error {
	validate := validator.New()
	if err := validate.Struct(article); err != nil {
		return fmt.Errorf("記事の検証に失敗しました: %w", err)
	}

	// タイトルの長さチェック
	titleLen := len([]rune(article.Title))
	if titleLen < MinTitleLength || titleLen > MaxTitleLength {
		return fmt.Errorf("記事タイトルは%d〜%d文字である必要があります (現在: %d文字)", MinTitleLength, MaxTitleLength, titleLen)
	}

	// コンテンツテキストの長さチェック
	if article.ContentText != "" {
		contentLen := len([]rune(article.ContentText))
		if contentLen < MinContentLength || contentLen > MaxContentLength {
			return fmt.Errorf("記事コンテンツは%d〜%d文字である必要があります (現在: %d文字)", MinContentLength, MaxContentLength, contentLen)
		}
	}

	return nil
}

// ValidateArticleEvaluation は記事評価の妥当性を検証します
func ValidateArticleEvaluation(eval *ArticleEvaluation) error {
	validate := validator.New()
	if err := validate.Struct(eval); err != nil {
		return fmt.Errorf("記事評価の検証に失敗しました: %w", err)
	}

	// 要約の長さチェック
	summaryLen := len([]rune(eval.Summary))
	if summaryLen < MinSummaryLength || summaryLen > MaxSummaryLength {
		return fmt.Errorf("要約は%d〜%d文字である必要があります (現在: %d文字)", MinSummaryLength, MaxSummaryLength, summaryLen)
	}

	// 関連性がある場合はマッチングトピックが必要
	if eval.IsRelevant && len(eval.MatchingTopics) == 0 {
		return fmt.Errorf("関連性がある記事には少なくとも1つのマッチングトピックが必要です")
	}

	return nil
}
