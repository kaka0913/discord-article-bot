package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kaka0913/discord-article-bot/internal/config"
)

// EvaluationResult はGemini APIからの評価結果を表します
type EvaluationResult struct {
	RelevanceScore int      `json:"relevance_score"`
	MatchingTopics []string `json:"matching_topics"`
	Summary        string   `json:"summary"`
	Reasoning      string   `json:"reasoning"`
	IsAIGenerated  bool     `json:"is_ai_generated"`
}

// Evaluator は記事の関連性評価を行います
type Evaluator struct {
	client *Client
}

// NewEvaluator は新しいEvaluatorを作成します
func NewEvaluator(client *Client) *Evaluator {
	return &Evaluator{
		client: client,
	}
}

// EvaluateArticle は記事を評価し、ArticleEvaluationを返します
func (e *Evaluator) EvaluateArticle(
	ctx context.Context,
	article *config.Article,
	topics []string,
	minRelevanceScore int,
) (*config.ArticleEvaluation, error) {
	// プロンプトを構築
	prompt := buildEvaluationPrompt(article, topics)

	// Gemini APIを呼び出し
	response, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
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
	var result EvaluationResult
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse evaluation result: %w", err)
	}

	// 検証
	if err := ValidateEvaluationResult(&result); err != nil {
		return nil, fmt.Errorf("invalid evaluation result: %w", err)
	}

	// ArticleEvaluationに変換
	evaluation := &config.ArticleEvaluation{
		ArticleURL:     article.URL,
		RelevanceScore: result.RelevanceScore,
		MatchingTopics: result.MatchingTopics,
		Summary:        result.Summary,
		EvaluatedAt:    time.Now(),
		IsRelevant:     result.RelevanceScore >= minRelevanceScore,
	}

	return evaluation, nil
}

// buildEvaluationPrompt は評価用のプロンプトを構築します
func buildEvaluationPrompt(article *config.Article, topics []string) string {
	topicsJSON, _ := json.Marshal(topics) // []string のマーシャルは常に成功する

	return fmt.Sprintf(`あなたは技術コンテンツキュレーションの専門家です。以下の記事を次のトピックとの関連性について評価してください: %s

記事タイトル: %s
記事内容: %s

JSON形式で評価を提供してください:
{
  "relevance_score": <0-100の整数>,
  "matching_topics": [<一致するトピック名の配列>],
  "summary": "<50-200文字の要約>",
  "reasoning": "<スコアの簡単な説明>",
  "is_ai_generated": <true/false>
}

スコアリング基準（加算方式、最大100点）:

【AI生成記事の判定】（必須チェック）
- 人間による執筆と判断: 継続して評価
- AI生成記事の可能性が高い: 即座に0点を返す
  判定基準:
  * 過度に形式的で個性のない文体
  * 具体的な実装や経験の欠如
  * 表面的な情報の羅列のみ
  * 表現が大袈裟で具体的でない

【トピックマッチング】（最大30点）
- 3つ以上のトピックに詳細な実装例で言及: +30点
- 2つのトピックに詳細な実装例で言及: +20点
- 1つのトピックに詳細な実装例で言及: +15点
- 複数トピックに言及するが表面的: +10点
- 1つのトピックに軽く言及: +5点
- トピックに全く言及なし: +0点

【内容の具体性】（最大30点）
- 実際のコード例・コマンド・設定ファイルを複数含む: +30点
- 実装方法の詳細な手順とコード例を含む: +25点
- アーキテクチャ図や設計パターンの具体的な解説: +20点
- ベストプラクティスと理由の説明: +15点
- 概念的な説明と簡単な例: +10点
- 抽象的な概念の説明のみ: +5点

【実用性】（最大25点）
- 実際のプロジェクトで即座に適用可能な実装: +25点
- ステップバイステップのチュートリアル: +20点
- 実務で参考になる設計思想と具体例: +15点
- 参考情報としての価値あり: +10点
- 一般的な情報の紹介のみ: +5点

【記事の深さ】（最大15点）
- 包括的で詳細な解説（実質2000文字以上）: +15点
- 中程度の詳細な解説（実質1000-2000文字）: +10点
- 簡潔だが要点を押さえた解説（実質500-1000文字）: +7点
- 短い紹介記事（実質500文字未満）: +3点

最終スコア = 合計点（最大100点、AI生成判定の場合は0点）

スコア区分の目安:
- 80-100点: 複数トピック + 詳細なコード例 + 即座に実用可能 + 包括的
- 60-79点: 1-2トピック + 具体的な実装 + 実用的 + 詳細
- 40-59点: トピック言及 + 概念説明 + やや実用的
- 0-39点: トピック言及なし or 表面的 or AI生成

重要な注意事項:
- matching_topicsには%sからのトピックのみを含める（幻覚トピック禁止）
- 要約は簡潔（50-200文字）で主要なポイントを強調すること
- AI生成の疑いがある場合は必ず0点とし、reasoningに判定理由を記載
- 同じトピックへの複数の表面的言及より、1つのトピックへの深い言及を高く評価`,
		string(topicsJSON),
		article.Title,
		TruncateContent(article.ContentText, MaxPromptContentLength),
		string(topicsJSON),
	)
}

// DetermineRejectionReason は評価結果から却下理由を判定します
func DetermineRejectionReason(result *EvaluationResult) string {
	// AI生成記事と判定された場合
	if result.IsAIGenerated {
		return config.ReasonLowRelevance
	}

	// スコアが0の場合
	if result.RelevanceScore == 0 {
		// トピックマッチがない場合
		if len(result.MatchingTopics) == 0 {
			return config.ReasonNoTopicMatch
		}
		return config.ReasonLowRelevance
	}

	// トピックマッチなし
	if len(result.MatchingTopics) == 0 {
		return config.ReasonNoTopicMatch
	}

	// 低関連性
	return config.ReasonLowRelevance
}

// extractJSONFromMarkdown はマークダウンコードブロックからJSONを抽出します
// Gemini 2.0は ```json ... ``` の形式でJSONを返すことがあるため、これを除去します
func extractJSONFromMarkdown(text string) string {
	text = strings.TrimSpace(text)

	// ```json で始まり ``` で終わる場合
	if strings.HasPrefix(text, "```json") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
		return text
	}

	// ``` で始まり ``` で終わる場合（言語指定なし）
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
		return text
	}

	// マークダウンコードブロックがない場合はそのまま返す
	return text
}
