// Package config は設定ファイルの読み込み、検証、管理を提供します
package config

import "time"

// RSSSource はRSSフィードソースを表します
type RSSSource struct {
	URL     string `json:"url" validate:"required,url"`
	Name    string `json:"name" validate:"required,min=1,max=50"`
	Enabled bool   `json:"enabled"`
}

// InterestTopic はユーザーの興味のあるトピックを表します
type InterestTopic struct {
	Topic    string   `json:"topic" validate:"required,min=1,max=50"`
	Aliases  []string `json:"aliases,omitempty"`
	Priority string   `json:"priority" validate:"required,oneof=high medium low"`
}

// GetPriorityMultiplier は優先度に応じたスコアの倍率を返します
func (t *InterestTopic) GetPriorityMultiplier() float64 {
	switch t.Priority {
	case "high":
		return 2.0
	case "medium":
		return 1.0
	case "low":
		return 0.5
	default:
		return 1.0
	}
}

// NotificationSettings は通知に関する設定を表します
type NotificationSettings struct {
	MaxArticles       int `json:"max_articles" validate:"required,min=1,max=10"`
	MinArticles       int `json:"min_articles" validate:"required,min=1,max=10"`
	MinRelevanceScore int `json:"min_relevance_score" validate:"required,min=0,max=100"`
}

// Config はアプリケーション全体の設定を表します
type Config struct {
	RSSSources           []RSSSource          `json:"rss_sources" validate:"required,min=1,max=10,dive"`
	Interests            []InterestTopic      `json:"interests" validate:"required,min=1,max=50,dive"`
	NotificationSettings NotificationSettings `json:"notification_settings" validate:"required"`
}

// GetEnabledSources は有効なRSSソースのみを返します
func (c *Config) GetEnabledSources() []RSSSource {
	var enabled []RSSSource
	for _, source := range c.RSSSources {
		if source.Enabled {
			enabled = append(enabled, source)
		}
	}
	return enabled
}

// Article はRSSフィードから取得した記事を表します
type Article struct {
	Title         string    `json:"title"`
	URL           string    `json:"url" validate:"required,url"`
	PublishedDate time.Time `json:"published_date,omitempty"`
	SourceFeed    string    `json:"source_feed"`
	ContentText   string    `json:"content_text,omitempty"`
	FetchedAt     time.Time `json:"fetched_at"`
}

// ArticleEvaluation はLLMによる記事の評価結果を表します
type ArticleEvaluation struct {
	ArticleURL     string    `json:"article_url" validate:"required,url"`
	RelevanceScore int       `json:"relevance_score" validate:"min=0,max=100"`
	MatchingTopics []string  `json:"matching_topics"`
	Summary        string    `json:"summary" validate:"required,min=50,max=200"`
	EvaluatedAt    time.Time `json:"evaluated_at"`
	IsRelevant     bool      `json:"is_relevant"`
}

// CuratedArticle は選択された記事と評価の組み合わせを表します
type CuratedArticle struct {
	Article    Article           `json:"article"`
	Evaluation ArticleEvaluation `json:"evaluation"`
	Rank       int               `json:"rank" validate:"min=1,max=10"`
}

// NotifiedArticle はDiscordに通知済みの記事を表します（Firestore保存用）
type NotifiedArticle struct {
	NotifiedAt       time.Time `firestore:"notified_at"`
	DiscordMessageID string    `firestore:"discord_message_id"`
	ArticleTitle     string    `firestore:"article_title"`
	RelevanceScore   int       `firestore:"relevance_score"`
}

// RejectedArticle は却下された記事を表します（Firestore保存用）
type RejectedArticle struct {
	EvaluatedAt    time.Time `firestore:"evaluated_at"`
	Reason         string    `firestore:"reason"` // "low_relevance" | "no_topic_match" | "content_extraction_failed"
	RelevanceScore *int      `firestore:"relevance_score,omitempty"`
}

// RejectedArticleReason は記事が却下された理由を表す定数
const (
	ReasonLowRelevance            = "low_relevance"
	ReasonNoTopicMatch            = "no_topic_match"
	ReasonContentExtractionFailed = "content_extraction_failed"
)
