package llm

import "fmt"

const (
	// MaxPromptContentLength はプロンプトに含めるコンテンツの最大文字数
	// Gemini Flash APIの入力トークン制限（約50kトークン）を考慮し、
	// 記事本文を適切な長さに制限する
	MaxPromptContentLength = 10000
)

// TruncateContent はコンテンツを指定した最大文字数に切り詰めます
func TruncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

// ValidateEvaluationResult は評価結果を検証します
func ValidateEvaluationResult(result *EvaluationResult) error {
	// スコアの範囲チェック
	if result.RelevanceScore < 0 || result.RelevanceScore > 100 {
		return fmt.Errorf("relevance_score must be between 0 and 100, got %d", result.RelevanceScore)
	}

	// 要約の長さチェック
	summaryLen := len([]rune(result.Summary))
	if summaryLen < 50 || summaryLen > 200 {
		return fmt.Errorf("summary must be between 50 and 200 characters, got %d", summaryLen)
	}

	// スコアが0より大きい場合、一致するトピックが必要
	if result.RelevanceScore > 0 && len(result.MatchingTopics) == 0 {
		return fmt.Errorf("matching_topics must not be empty when relevance_score > 0")
	}

	return nil
}
