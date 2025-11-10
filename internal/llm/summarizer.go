package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

// ArticleForSummary はサマリー生成用の記事情報
type ArticleForSummary struct {
	Title          string
	Summary        string
	RelevanceScore int
	MatchingTopics []string
}

// ArticlesSummaryResult はGemini APIからのサマリー生成結果
type ArticlesSummaryResult struct {
	OverallSummary  string   `json:"overall_summary"`
	MustRead        string   `json:"must_read"`
	Recommendations []string `json:"recommendations"`
}

// GenerateArticlesSummary は複数記事の全体サマリーを生成します
func (e *Evaluator) GenerateArticlesSummary(
	ctx context.Context,
	articles []ArticleForSummary,
) (*ArticlesSummaryResult, error) {
	// プロンプトを構築
	prompt := buildSummaryPrompt(articles)

	// Gemini APIを呼び出し
	response, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// 応答の安全性チェック
	if len(response.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}
	if len(response.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no parts in candidate content")
	}

	// 応答からJSONテキストを抽出
	jsonText := response.Candidates[0].Content.Parts[0].Text

	// マークダウンコードブロックを除去（Gemini 2.0対応）
	jsonText = extractJSONFromMarkdown(jsonText)

	// JSONをパース
	var result ArticlesSummaryResult
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse summary result: %w", err)
	}

	return &result, nil
}

// buildSummaryPrompt はサマリー生成用のプロンプトを構築します
func buildSummaryPrompt(articles []ArticleForSummary) string {
	// 記事情報をJSON形式で整形 (構造体のマーシャルは常に成功する)
	articlesJSON, _ := json.MarshalIndent(articles, "", "  ")

	// recommendations配列の例を動的に生成
	recommendationExamples := ""
	for i := 1; i <= len(articles); i++ {
		if i > 1 {
			recommendationExamples += ",\n    "
		} else {
			recommendationExamples += "\n    "
		}
		recommendationExamples += fmt.Sprintf("\"<記事%dのタイトルと推奨理由を1文で>\"", i)
	}

	return fmt.Sprintf(`あなたは技術記事のキュレーターです。以下の厳選された記事について、読者に向けた魅力的なサマリーを作成してください。

記事リスト:
%s

以下のJSON形式でサマリーを提供してください:
{
  "overall_summary": "<全体的なテーマや傾向を1-2文で説明>",
  "must_read": "<最もスコアが高い記事について、なぜ特に読むべきかを1文で説明>",
  "recommendations": [%s
  ]
}

作成ガイドライン:
- overall_summary: 全記事を俯瞰して、共通テーマや今回の特徴を簡潔に説明
- must_read: 最高スコアの記事について「なぜ特に読むべきか」を強調
- recommendations: 各記事について「○○な人におすすめ」「○○を学べる」など具体的な推奨理由
- 簡潔で読みやすく、技術者の興味を引く表現を使う
- スコアの数値は直接的に言及せず、「特におすすめ」「実践的」などの表現を使う
- recommendations配列には必ず%d個の要素を含めること（記事数と一致）

重要: 必ずJSON形式で応答してください。マークダウンコードブロックは使用しても構いません。`, string(articlesJSON), recommendationExamples, len(articles))
}
