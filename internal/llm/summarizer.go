package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kaka0913/discord-article-bot/internal/config"
)

// SummaryResult はGemini APIからの要約結果を表します
type SummaryResult struct {
	Summary string `json:"summary"`
}

// Summarizer は記事の要約生成を行います
type Summarizer struct {
	client *Client
}

// NewSummarizer は新しいSummarizerを作成します
func NewSummarizer(client *Client) *Summarizer {
	return &Summarizer{
		client: client,
	}
}

// SummarizeArticle は記事の要約を生成します
func (s *Summarizer) SummarizeArticle(
	ctx context.Context,
	article *config.Article,
) (string, error) {
	// プロンプトを構築
	prompt := buildSummaryPrompt(article)

	// Gemini APIを呼び出し
	response, err := s.client.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// 応答からJSONテキストを抽出
	jsonText := response.Candidates[0].Content.Parts[0].Text

	// JSONをパース
	var result SummaryResult
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return "", fmt.Errorf("failed to parse summary result: %w", err)
	}

	// 検証
	summaryLen := len([]rune(result.Summary))
	if summaryLen < 50 || summaryLen > 200 {
		return "", fmt.Errorf("summary length must be between 50 and 200 characters, got %d", summaryLen)
	}

	return result.Summary, nil
}

// buildSummaryPrompt は要約用のプロンプトを構築します
func buildSummaryPrompt(article *config.Article) string {
	return fmt.Sprintf(`以下の技術記事を50-200文字で要約してください。主要なポイントを簡潔に強調してください。

記事タイトル: %s
記事内容: %s

JSON形式で要約を提供してください:
{
  "summary": "<50-200文字の要約>"
}`,
		article.Title,
		truncateContent(article.ContentText, 10000),
	)
}
