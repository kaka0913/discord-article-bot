// Package main はRSS記事キュレーションBotのCloud Functionsエントリーポイントです
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/kaka0913/discord-article-bot/internal/article"
	"github.com/kaka0913/discord-article-bot/internal/config"
	"github.com/kaka0913/discord-article-bot/internal/discord"
	"github.com/kaka0913/discord-article-bot/internal/llm"
	"github.com/kaka0913/discord-article-bot/internal/logging"
	"github.com/kaka0913/discord-article-bot/internal/rss"
	"github.com/kaka0913/discord-article-bot/internal/secrets"
	"github.com/kaka0913/discord-article-bot/internal/storage"
)

func init() {
	// Cloud Functionsイベントハンドラーを登録
	functions.CloudEvent("CurateArticles", curateArticles)
}

func main() {
	// Cloud Functionsフレームワークを起動
	// PORT環境変数が設定されている場合、HTTPサーバーとして起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}

// curateArticles はPub/Subトリガーによって実行されるメイン処理
func curateArticles(ctx context.Context, e event.Event) error {
	// ロガーを作成
	logger := logging.NewLogger()
	ctx = logging.ToContext(ctx, logger)

	logger.Info("記事キュレーション処理を開始します")

	// 環境変数を取得
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		return fmt.Errorf("GCP_PROJECT_ID環境変数が設定されていません")
	}

	configSource := os.Getenv("CONFIG_SOURCE")
	if configSource == "" {
		// デフォルトはローカルのconfig.json
		configSource = "config.json"
	}

	// Secret Managerクライアントを初期化
	secretMgr, err := secrets.NewManager(ctx, projectID)
	if err != nil {
		return fmt.Errorf("Secret Managerクライアントの初期化に失敗: %w", err)
	}
	defer secretMgr.Close()

	// シークレットを取得
	discordWebhookURL, err := secretMgr.GetSecret(ctx, "discord-webhook-url")
	if err != nil {
		return fmt.Errorf("Discord Webhook URLの取得に失敗: %w", err)
	}

	geminiAPIKey, err := secretMgr.GetSecret(ctx, "gemini-api-key")
	if err != nil {
		return fmt.Errorf("Gemini APIキーの取得に失敗: %w", err)
	}

	// Firestoreクライアントを初期化
	firestoreClient, err := storage.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("Firestoreクライアントの初期化に失敗: %w", err)
	}
	defer firestoreClient.Close()

	// 設定を読み込む
	configLoader := config.NewLoader()
	cfg, err := configLoader.Load(ctx, configSource)
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗: %w", err)
	}

	// 設定を検証
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("設定の検証に失敗: %w", err)
	}

	logger.Info("設定を読み込みました",
		"rssSources", len(cfg.RSSSources),
		"interests", len(cfg.Interests),
		"maxArticles", cfg.NotificationSettings.MaxArticles,
	)

	// 依存関係を初期化（設定値から取得）
	rssFetcher := rss.NewFetcher(time.Duration(cfg.TimeoutSettings.RSSFetchTimeoutSeconds) * time.Second)
	rssParser := rss.NewParser()
	articleFetcher := article.NewFetcher(time.Duration(cfg.TimeoutSettings.ArticleFetchTimeoutSeconds) * time.Second)
	articleExtractor := article.NewExtractor(cfg.TimeoutSettings.MinTextLength, cfg.TimeoutSettings.MaxTextLength)
	llmClient := llm.NewClient(geminiAPIKey)
	llmEvaluator := llm.NewEvaluator(llmClient)
	discordClient := discord.NewClient(discordWebhookURL, logger)

	// メインオーケストレーションを実行
	if err := orchestrateCuration(
		ctx,
		cfg,
		rssFetcher,
		rssParser,
		articleFetcher,
		articleExtractor,
		llmEvaluator,
		discordClient,
		firestoreClient,
		logger,
	); err != nil {
		logger.Error("記事キュレーション処理に失敗しました", "error", err)
		return err
	}

	logger.Info("記事キュレーション処理が正常に完了しました")
	return nil
}

// orchestrateCuration はメインのオーケストレーションロジックを実行します
func orchestrateCuration(
	ctx context.Context,
	cfg *config.Config,
	rssFetcher *rss.Fetcher,
	rssParser *rss.Parser,
	articleFetcher *article.Fetcher,
	articleExtractor *article.Extractor,
	llmEvaluator *llm.Evaluator,
	discordClient *discord.Client,
	firestoreClient *storage.Client,
	logger logging.Logger,
) error {
	// 1. RSSフィードから記事を取得
	logger.Info("RSSフィードから記事を取得中")
	allArticles := []rss.Article{}

	enabledSources := cfg.GetEnabledSources()
	for _, source := range enabledSources {
		logger.Info("RSSソースを処理中", "source", source.Name, "url", source.URL)

		// RSSフィードを取得
		xmlData, err := rssFetcher.Fetch(ctx, source.URL)
		if err != nil {
			// エラーをログに記録して続行（FR-014: 1つのソースが失敗しても処理を続行）
			logger.Error("RSSフィードの取得に失敗しました。スキップします", "source", source.Name, "error", err)
			continue
		}

		// RSSフィードをパース
		articles, err := rssParser.Parse(ctx, xmlData, source.Name)
		if err != nil {
			logger.Error("RSSフィードのパースに失敗しました。スキップします", "source", source.Name, "error", err)
			continue
		}

		logger.Info("RSSフィードから記事を取得しました", "source", source.Name, "count", len(articles))
		allArticles = append(allArticles, articles...)
	}

	if len(allArticles) == 0 {
		logger.Warn("処理可能な記事が見つかりませんでした")
		return nil
	}

	logger.Info("すべてのRSSフィードから記事を取得しました", "totalCount", len(allArticles))

	// 2. 重複チェック（通知済み・却下済み記事を除外）
	logger.Info("重複チェックを実行中")
	filteredArticles := []rss.Article{}
	var firestoreErrorCount int
	var notifiedSkipCount, rejectedSkipCount int
	const maxFirestoreErrors = 10

	for _, article := range allArticles {
		// 通知済みチェック
		notified, err := firestoreClient.IsArticleNotified(ctx, article.URL)
		if err != nil {
			firestoreErrorCount++
			logger.Error("通知済みチェックに失敗しました", "url", article.URL, "error", err)
			if firestoreErrorCount >= maxFirestoreErrors {
				return fmt.Errorf("Firestoreエラーが多すぎます（%d件）。処理を中止します", firestoreErrorCount)
			}
			// エラー時は安全側に倒して処理を続行（重複のリスクはあるが、記事を見逃すよりまし）
			filteredArticles = append(filteredArticles, article)
			continue
		}
		if notified {
			notifiedSkipCount++
			continue
		}

		// 却下済みチェック
		rejected, err := firestoreClient.IsArticleRejected(ctx, article.URL)
		if err != nil {
			firestoreErrorCount++
			logger.Error("却下済みチェックに失敗しました", "url", article.URL, "error", err)
			if firestoreErrorCount >= maxFirestoreErrors {
				return fmt.Errorf("Firestoreエラーが多すぎます（%d件）。処理を中止します", firestoreErrorCount)
			}
			// エラー時は安全側に倒して処理を続行
			filteredArticles = append(filteredArticles, article)
			continue
		}
		if rejected {
			rejectedSkipCount++
			continue
		}

		filteredArticles = append(filteredArticles, article)
	}

	logger.Info("重複チェック完了",
		"originalCount", len(allArticles),
		"filteredCount", len(filteredArticles),
		"notifiedSkipped", notifiedSkipCount,
		"rejectedSkipped", rejectedSkipCount,
		"firestoreErrors", firestoreErrorCount,
	)

	if len(filteredArticles) == 0 {
		logger.Info("新しい記事が見つかりませんでした")
		return nil
	}

	// 3. 記事コンテンツを取得して評価
	logger.Info("記事を評価中")
	evaluatedArticles := []config.ArticleEvaluation{}

	// 興味トピックをリストに変換
	interestTopics := make([]string, len(cfg.Interests))
	for i, interest := range cfg.Interests {
		interestTopics[i] = interest.Topic
	}

	for _, rssArticle := range filteredArticles {
		// 記事HTMLを取得
		htmlContent, err := articleFetcher.Fetch(ctx, rssArticle.URL)
		if err != nil {
			logger.Warn("記事HTMLの取得に失敗しました。スキップします", "url", rssArticle.URL, "error", err)
			// コンテンツ抽出失敗として記録
			if saveErr := firestoreClient.SaveRejectedArticle(ctx, rssArticle.URL, config.ReasonContentExtractionFailed, nil); saveErr != nil {
				logger.Error("却下記事の保存に失敗", "url", rssArticle.URL, "error", saveErr)
			}
			continue
		}

		// 記事本文とタイトルを抽出
		extractedTitle, extractedText, err := articleExtractor.ExtractWithTitle(ctx, htmlContent, rssArticle.URL)
		if err != nil {
			logger.Warn("記事本文の抽出に失敗しました。スキップします", "url", rssArticle.URL, "error", err)
			// コンテンツ抽出失敗として記録
			if saveErr := firestoreClient.SaveRejectedArticle(ctx, rssArticle.URL, config.ReasonContentExtractionFailed, nil); saveErr != nil {
				logger.Error("却下記事の保存に失敗", "url", rssArticle.URL, "error", saveErr)
			}
			continue
		}

		// タイトルが抽出された場合は使用、そうでなければRSSのタイトルを使用
		title := rssArticle.Title
		if extractedTitle != "" {
			title = extractedTitle
		}

		// config.Articleを作成
		configArticle := &config.Article{
			Title:         title,
			URL:           rssArticle.URL,
			PublishedDate: rssArticle.PublishedDate,
			SourceFeed:    rssArticle.SourceFeed,
			ContentText:   extractedText,
			FetchedAt:     rssArticle.FetchedAt,
		}

		// LLMで評価
		evaluation, err := llmEvaluator.EvaluateArticle(ctx, configArticle, interestTopics, cfg.NotificationSettings.MinRelevanceScore)
		if err != nil {
			logger.Error("記事の評価に失敗しました。スキップします", "url", rssArticle.URL, "error", err)
			continue
		}

		logger.Info("記事を評価しました",
			"url", rssArticle.URL,
			"score", evaluation.RelevanceScore,
			"isRelevant", evaluation.IsRelevant,
		)

		// 関連性がない記事は却下
		if !evaluation.IsRelevant {
			logger.Debug("関連性がない記事を却下", "url", rssArticle.URL, "score", evaluation.RelevanceScore)
			// 却下理由を判定して保存
			reason := config.ReasonLowRelevance
			if len(evaluation.MatchingTopics) == 0 {
				reason = config.ReasonNoTopicMatch
			}
			if saveErr := firestoreClient.SaveRejectedArticle(ctx, rssArticle.URL, reason, &evaluation.RelevanceScore); saveErr != nil {
				logger.Error("却下記事の保存に失敗", "url", rssArticle.URL, "error", saveErr)
			}
			continue
		}

		evaluatedArticles = append(evaluatedArticles, *evaluation)
	}

	logger.Info("記事の評価完了", "relevantCount", len(evaluatedArticles))

	if len(evaluatedArticles) == 0 {
		logger.Info("関連性のある記事が見つかりませんでした")
		return nil
	}

	// 4. スコア順にソートして上位記事を選択
	sort.Slice(evaluatedArticles, func(i, j int) bool {
		return evaluatedArticles[i].RelevanceScore > evaluatedArticles[j].RelevanceScore
	})

	// 最大記事数を超える場合は切り詰める
	maxArticles := cfg.NotificationSettings.MaxArticles
	if len(evaluatedArticles) > maxArticles {
		evaluatedArticles = evaluatedArticles[:maxArticles]
	}

	logger.Info("上位記事を選択しました", "count", len(evaluatedArticles))

	// 5. Discord通知用のペイロードを作成
	// パフォーマンス最適化: 記事URLからの高速検索用マップを作成（O(n²) → O(n)）
	articlesByURL := make(map[string]rss.Article, len(filteredArticles))
	for _, article := range filteredArticles {
		articlesByURL[article.URL] = article
	}

	discordArticles := make([]discord.Article, len(evaluatedArticles))
	for i, eval := range evaluatedArticles {
		// マップから記事を取得（O(1)検索）
		article, ok := articlesByURL[eval.ArticleURL]
		title := "Unknown"
		sourceFeed := "Unknown"
		if ok {
			title = article.Title
			sourceFeed = article.SourceFeed
		}

		discordArticles[i] = discord.Article{
			Title:       title,
			Description: eval.Summary,
			URL:         eval.ArticleURL,
			Relevance:   eval.RelevanceScore,
			Topics:      eval.MatchingTopics,
			Source:      sourceFeed,
		}
	}

	// 6. Discordに通知
	logger.Info("Discordに通知中", "articleCount", len(discordArticles))

	date := time.Now().Format("2006-01-02")
	messageID, err := discordClient.PostArticles(ctx, discordArticles, date)
	if err != nil {
		return fmt.Errorf("Discord通知に失敗: %w", err)
	}

	logger.Info("Discordへの通知に成功しました", "messageID", messageID)

	// 7. 通知済み記事をFirestoreに保存
	logger.Info("通知済み記事をFirestoreに保存中")
	for _, eval := range evaluatedArticles {
		// マップから記事タイトルを取得（O(1)検索）
		article, ok := articlesByURL[eval.ArticleURL]
		title := "Unknown"
		if ok {
			title = article.Title
		}
		if err := firestoreClient.SaveNotifiedArticle(ctx, eval.ArticleURL, messageID, title, eval.RelevanceScore); err != nil {
			logger.Error("通知済み記事の保存に失敗", "url", eval.ArticleURL, "error", err)
			// エラーをログに記録するが、処理は続行
		}
	}

	logger.Info("通知済み記事の保存完了")
	return nil
}
