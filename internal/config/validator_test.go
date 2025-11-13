package config

import (
	"strings"
	"testing"
	"time"
)

// makeTestContent はテスト用の文字列を生成します
func makeTestContent(length int) string {
	return strings.Repeat("x", length)
}

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "有効な設定",
			config: &Config{
				RSSSources: []RSSSource{
					{URL: "https://dev.to/feed", Name: "Dev.to", Enabled: true},
				},
				Interests: []InterestTopic{
					{Topic: "Go", Priority: "high"},
				},
				NotificationSettings: NotificationSettings{
					MaxArticles:       5,
					MinArticles:       3,
					MinRelevanceScore: 70,
				},
				TimeoutSettings: TimeoutSettings{
					RSSFetchTimeoutSeconds:     10,
					ArticleFetchTimeoutSeconds: 10,
					MinTextLength:              100,
					MaxTextLength:              50000,
				},
			},
			wantErr: false,
		},
		{
			name: "有効なRSSソースが1つもない",
			config: &Config{
				RSSSources: []RSSSource{
					{URL: "https://dev.to/feed", Name: "Dev.to", Enabled: false},
				},
				Interests: []InterestTopic{
					{Topic: "Go", Priority: "high"},
				},
				NotificationSettings: NotificationSettings{
					MaxArticles:       5,
					MinArticles:       3,
					MinRelevanceScore: 70,
				},
				TimeoutSettings: TimeoutSettings{
					RSSFetchTimeoutSeconds:     10,
					ArticleFetchTimeoutSeconds: 10,
					MinTextLength:              100,
					MaxTextLength:              50000,
				},
			},
			wantErr: true,
			errMsg:  "少なくとも1つのRSSソースを有効にする必要があります",
		},
		{
			name: "重複する興味トピック",
			config: &Config{
				RSSSources: []RSSSource{
					{URL: "https://dev.to/feed", Name: "Dev.to", Enabled: true},
				},
				Interests: []InterestTopic{
					{Topic: "Go", Priority: "high"},
					{Topic: "Go", Priority: "medium"},
				},
				NotificationSettings: NotificationSettings{
					MaxArticles:       5,
					MinArticles:       3,
					MinRelevanceScore: 70,
				},
				TimeoutSettings: TimeoutSettings{
					RSSFetchTimeoutSeconds:     10,
					ArticleFetchTimeoutSeconds: 10,
					MinTextLength:              100,
					MaxTextLength:              50000,
				},
			},
			wantErr: true,
			errMsg:  "重複する興味トピックが見つかりました: Go",
		},
		{
			name: "MinArticlesがMaxArticlesより大きい",
			config: &Config{
				RSSSources: []RSSSource{
					{URL: "https://dev.to/feed", Name: "Dev.to", Enabled: true},
				},
				Interests: []InterestTopic{
					{Topic: "Go", Priority: "high"},
				},
				NotificationSettings: NotificationSettings{
					MaxArticles:       3,
					MinArticles:       5,
					MinRelevanceScore: 70,
				},
				TimeoutSettings: TimeoutSettings{
					RSSFetchTimeoutSeconds:     10,
					ArticleFetchTimeoutSeconds: 10,
					MinTextLength:              100,
					MaxTextLength:              50000,
				},
			},
			wantErr: true,
			errMsg:  "min_articles (5) はmax_articles (3) 以下である必要があります",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() エラー = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("エラーメッセージが一致しません:\n期待=%s\n実際=%s", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestValidateArticle(t *testing.T) {
	tests := []struct {
		name    string
		article *Article
		wantErr bool
	}{
		{
			name: "有効な記事",
			article: &Article{
				Title:       "Building Microservices with Go",
				URL:         "https://dev.to/example/post",
				SourceFeed:  "Dev.to",
				ContentText: "In this article, we explore how to build scalable microservices using Go and Kubernetes. " + makeTestContent(50),
				FetchedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "タイトルが短すぎる",
			article: &Article{
				Title:       "Go",
				URL:         "https://dev.to/example/post",
				SourceFeed:  "Dev.to",
				ContentText: makeTestContent(200),
				FetchedAt:   time.Now(),
			},
			wantErr: true,
		},
		{
			name: "コンテンツが短すぎる",
			article: &Article{
				Title:       "Building Microservices with Go",
				URL:         "https://dev.to/example/post",
				SourceFeed:  "Dev.to",
				ContentText: "Short",
				FetchedAt:   time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArticle(tt.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArticle() エラー = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateArticleEvaluation(t *testing.T) {
	tests := []struct {
		name    string
		eval    *ArticleEvaluation
		wantErr bool
	}{
		{
			name: "有効な評価",
			eval: &ArticleEvaluation{
				ArticleURL:     "https://dev.to/example/post",
				RelevanceScore: 85,
				MatchingTopics: []string{"Go", "Kubernetes"},
				Summary:        "A comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters.",
				EvaluatedAt:    time.Now(),
				IsRelevant:     true,
			},
			wantErr: false,
		},
		{
			name: "要約が短すぎる",
			eval: &ArticleEvaluation{
				ArticleURL:     "https://dev.to/example/post",
				RelevanceScore: 85,
				MatchingTopics: []string{"Go"},
				Summary:        "Short summary",
				EvaluatedAt:    time.Now(),
				IsRelevant:     true,
			},
			wantErr: true,
		},
		{
			name: "関連性ありだがマッチングトピックなし",
			eval: &ArticleEvaluation{
				ArticleURL:     "https://dev.to/example/post",
				RelevanceScore: 85,
				MatchingTopics: []string{},
				Summary:        "A comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters.",
				EvaluatedAt:    time.Now(),
				IsRelevant:     true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateArticleEvaluation(tt.eval)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArticleEvaluation() エラー = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInterestTopic_GetPriorityMultiplier(t *testing.T) {
	tests := []struct {
		priority string
		expected float64
	}{
		{"high", 2.0},
		{"medium", 1.0},
		{"low", 0.5},
		{"invalid", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			topic := &InterestTopic{Priority: tt.priority}
			result := topic.GetPriorityMultiplier()
			if result != tt.expected {
				t.Errorf("GetPriorityMultiplier() = %v, 期待 %v", result, tt.expected)
			}
		})
	}
}
