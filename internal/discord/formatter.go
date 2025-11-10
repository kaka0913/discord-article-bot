package discord

import (
	"fmt"
	"strings"
)

const (
	// Discord Embedsの制約
	maxTitleLength       = 256
	maxDescriptionLength = 4096
	maxFieldNameLength   = 256
	maxFieldValueLength  = 1024
	maxFooterLength      = 2048

	// デフォルトのEmbed色（#58A5EF = 5814783）
	defaultEmbedColor = 5814783
)

// ArticlesSummary は記事全体のサマリー情報
type ArticlesSummary struct {
	OverallSummary  string
	MustRead        string
	Recommendations []string
}

// FormatArticlesPayload は記事リストをDiscord Webhook用のペイロードにフォーマット
func FormatArticlesPayload(articles []Article, date string, summary *ArticlesSummary) WebhookPayload {
	embeds := make([]EmbedObject, 0, len(articles)+1)

	// サマリーが提供されている場合、最初にサマリーEmbedを追加
	if summary != nil {
		summaryEmbed := formatSummaryEmbed(summary, len(articles))
		embeds = append(embeds, summaryEmbed)
	}

	// 各記事のEmbedを追加
	for _, article := range articles {
		embed := formatArticleEmbed(article)
		embeds = append(embeds, embed)
	}

	return WebhookPayload{
		Content: fmt.Sprintf("📰 Daily Tech Article Digest - %s", date),
		Embeds:  embeds,
	}
}

// formatSummaryEmbed はサマリー情報をEmbedオブジェクトにフォーマット
func formatSummaryEmbed(summary *ArticlesSummary, articleCount int) EmbedObject {
	// サマリーの説明文を構築
	var descriptionParts []string

	// 全体サマリー
	if summary.OverallSummary != "" {
		descriptionParts = append(descriptionParts, fmt.Sprintf("**📋 今日のハイライト**\n%s", summary.OverallSummary))
	}

	// 必読記事
	if summary.MustRead != "" {
		descriptionParts = append(descriptionParts, fmt.Sprintf("\n**⭐ 特におすすめ**\n%s", summary.MustRead))
	}

	// 各記事の推奨理由
	if len(summary.Recommendations) > 0 {
		descriptionParts = append(descriptionParts, "\n**📚 記事ガイド**")
		for i, rec := range summary.Recommendations {
			// 番号付きで推奨理由を表示
			descriptionParts = append(descriptionParts, fmt.Sprintf("%d. %s", i+1, rec))
		}
	}

	description := strings.Join(descriptionParts, "\n")
	description = truncateString(description, maxDescriptionLength)

	return EmbedObject{
		Title:       fmt.Sprintf("📊 本日の厳選記事 (%d件)", articleCount),
		Description: description,
		Color:       0x3498db, // 青色（#3498db = 3447003）
	}
}

// formatArticleEmbed は個別の記事をEmbedオブジェクトにフォーマット
func formatArticleEmbed(article Article) EmbedObject {
	// タイトルを制限内に収める
	title := truncateString(article.Title, maxTitleLength)

	// 説明を制限内に収める
	description := truncateString(article.Description, maxDescriptionLength)

	// Topicsをカンマ区切りの文字列に変換
	topicsValue := strings.Join(article.Topics, ", ")
	if topicsValue == "" {
		topicsValue = "N/A"
	}

	// フィールドを作成
	fields := []EmbedField{
		{
			Name:   "Relevance",
			Value:  fmt.Sprintf("%d/100", article.Relevance),
			Inline: true,
		},
		{
			Name:   "Topics",
			Value:  truncateString(topicsValue, maxFieldValueLength),
			Inline: true,
		},
	}

	// フッターを作成
	footer := &EmbedFooter{
		Text: truncateString(fmt.Sprintf("Source: %s", article.Source), maxFooterLength),
	}

	return EmbedObject{
		Title:       title,
		Description: description,
		URL:         article.URL,
		Color:       defaultEmbedColor,
		Fields:      fields,
		Footer:      footer,
	}
}

// truncateString は文字列を指定された長さに切り詰める（末尾に"..."を付ける）
func truncateString(s string, maxLen int) string {
	// 文字数（Unicodeコードポイント数）でカウント
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	// "..."を追加するため、maxLen-3の位置で切り詰める
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}

	return string(runes[:maxLen-3]) + "..."
}
