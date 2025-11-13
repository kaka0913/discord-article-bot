// Package function はRSS記事キュレーションBotのCloud Functionsエントリーポイントです
package function

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

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
	// Cloud Functions HTTPハンドラーを登録
	functions.HTTP("CuratorHandler", curatorHandler)
}

// handleError はエラーをログに記録し、HTTPエラーレスポンスを返す
// セキュリティ上の理由から、エラーの詳細はログにのみ記録し、HTTPレスポンスには含めない
func handleError(w http.ResponseWriter, logger logging.Logger, statusCode int, message string, err error) {
	if err != nil {
		logger.Error(message, "error", err)
		// 本番環境では詳細なエラーメッセージを返さない
		http.Error(w, message, statusCode)
	} else {
		logger.Error(message)
		http.Error(w, message, statusCode)
	}
}

// curatorHandler はHTTPトリガーによって実行されるメイン処理
func curatorHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.NewLogger()
	ctx = logging.ToContext(ctx, logger)

	logger.Info("記事キュレーション処理を開始します")

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		handleError(w, logger, http.StatusInternalServerError, "GCP_PROJECT_ID環境変数が設定されていません", nil)
		return
	}

	// CONFIG_URLまたはCONFIG_SOURCE環境変数から設定ソースを取得
	// CONFIG_URLを優先（gcloudデプロイで使用）
	configSource := os.Getenv("CONFIG_URL")
	if configSource == "" {
		configSource = os.Getenv("CONFIG_SOURCE")
	}
	if configSource == "" {
		// デフォルトはローカルのconfig.json
		configSource = "config.json"
	}

	secretMgr, err := secrets.NewManager(ctx, projectID)
	if err != nil {
		handleError(w, logger, http.StatusInternalServerError, "Secret Managerクライアントの初期化に失敗", err)
		return
	}
	defer secretMgr.Close()

	discordWebhookURL, err := secretMgr.GetSecret(ctx, "discord-webhook-url")
	if err != nil {
		handleError(w, logger, http.StatusInternalServerError, "Discord Webhook URLの取得に失敗", err)
		return
	}

	geminiAPIKey, err := secretMgr.GetSecret(ctx, "gemini-api-key")
	if err != nil {
		handleError(w, logger, http.StatusInternalServerError, "Gemini APIキーの取得に失敗", err)
		return
	}

	firestoreClient, err := storage.NewClient(ctx, projectID)
	if err != nil {
		handleError(w, logger, http.StatusInternalServerError, "Firestoreクライアントの初期化に失敗", err)
		return
	}
	defer firestoreClient.Close()

	configLoader := config.NewLoader()
	cfg, err := configLoader.Load(ctx, configSource)
	if err != nil {
		handleError(w, logger, http.StatusInternalServerError, "設定の読み込みに失敗", err)
		return
	}

	if err := config.ValidateConfig(cfg); err != nil {
		handleError(w, logger, http.StatusBadRequest, "設定の検証に失敗", err)
		return
	}

	logger.Info("設定を読み込みました",
		"rssSources", len(cfg.RSSSources),
		"interests", len(cfg.Interests),
		"maxArticles", cfg.NotificationSettings.MaxArticles,
	)

	rssFetcher := rss.NewFetcher(time.Duration(cfg.TimeoutSettings.RSSFetchTimeoutSeconds) * time.Second)
	rssParser := rss.NewParser()
	articleFetcher := article.NewFetcher(time.Duration(cfg.TimeoutSettings.ArticleFetchTimeoutSeconds) * time.Second)
	articleExtractor := article.NewExtractor(cfg.TimeoutSettings.MinTextLength, cfg.TimeoutSettings.MaxTextLength)
	llmClient := llm.NewClient(geminiAPIKey)
	llmEvaluator := llm.NewEvaluator(llmClient)
	discordClient := discord.NewClient(discordWebhookURL, logger)

	orchestrateCuration(
		w,
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
		0, // 本番環境では記事数制限なし
	)
}

func getArticleTitle(articlesByURL map[string]rss.Article, articleURL string) string {
	if article, ok := articlesByURL[articleURL]; ok {
		return article.Title
	}
	return "Unknown"
}

func orchestrateCuration(
	w http.ResponseWriter,
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
	maxEvaluationArticles int, // 評価する記事の最大数（0=無制限、ローカルテストでは3）
) {
	logger.Info("RSSフィードから記事を取得中")
	allArticles := []rss.Article{}

	enabledSources := cfg.GetEnabledSources()
	for _, source := range enabledSources {
		logger.Info("RSSソースを処理中", "source", source.Name, "url", source.URL)

		xmlData, err := rssFetcher.Fetch(ctx, source.URL)
		if err != nil {
			logger.Error("RSSフィードの取得に失敗しました。スキップします", "source", source.Name, "error", err)
			continue
		}

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
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "処理可能な記事が見つかりませんでした\n")
		return
	}

	logger.Info("すべてのRSSフィードから記事を取得しました", "totalCount", len(allArticles))

	logger.Info("重複チェックを実行中")
	filteredArticles := []rss.Article{}
	var firestoreErrorCount int
	var notifiedSkipCount, rejectedSkipCount int
	const maxFirestoreErrors = 10

	for _, article := range allArticles {
		notified, err := firestoreClient.IsArticleNotified(ctx, article.URL)
		if err != nil {
			firestoreErrorCount++
			logger.Error("通知済みチェックに失敗しました", "url", article.URL, "error", err)
			if firestoreErrorCount >= maxFirestoreErrors {
				logger.Error("Firestoreエラーが多すぎます。処理を中止します",
					"firestoreErrorCount", firestoreErrorCount,
					"processedArticles", len(allArticles),
					"filteredArticles", len(filteredArticles))
				handleError(w, logger, http.StatusInternalServerError, fmt.Sprintf("Firestoreエラーが多すぎます（%d件）。処理を中止します", firestoreErrorCount), nil)
				return
			}
			filteredArticles = append(filteredArticles, article)
			continue
		}
		if notified {
			notifiedSkipCount++
			continue
		}

		rejected, err := firestoreClient.IsArticleRejected(ctx, article.URL)
		if err != nil {
			firestoreErrorCount++
			logger.Error("却下済みチェックに失敗しました", "url", article.URL, "error", err)
			if firestoreErrorCount >= maxFirestoreErrors {
				handleError(w, logger, http.StatusInternalServerError, fmt.Sprintf("Firestoreエラーが多すぎます（%d件）。処理を中止します", firestoreErrorCount), nil)
				return
			}
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
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "新しい記事が見つかりませんでした\n")
		return
	}

	if maxEvaluationArticles > 0 && len(filteredArticles) > maxEvaluationArticles {
		logger.Info("記事数を制限します",
			"originalCount", len(filteredArticles),
			"limitedCount", maxEvaluationArticles,
			"reason", "API制限またはテスト環境",
		)
		filteredArticles = filteredArticles[:maxEvaluationArticles]
	}

	logger.Info("記事を評価中")
	evaluatedArticles := []config.ArticleEvaluation{}

	interestTopics := make([]string, len(cfg.Interests))
	for i, interest := range cfg.Interests {
		interestTopics[i] = interest.Topic
	}

	for _, rssArticle := range filteredArticles {
		htmlContent, err := articleFetcher.Fetch(ctx, rssArticle.URL)
		if err != nil {
			logger.Warn("記事HTMLの取得に失敗しました。スキップします", "url", rssArticle.URL, "error", err)
			if saveErr := firestoreClient.SaveRejectedArticle(ctx, rssArticle.URL, config.ReasonContentExtractionFailed, nil); saveErr != nil {
				logger.Error("却下記事の保存に失敗", "url", rssArticle.URL, "error", saveErr)
			}
			continue
		}

		extractedTitle, extractedText, err := articleExtractor.ExtractWithTitle(ctx, htmlContent, rssArticle.URL)
		if err != nil {
			logger.Warn("記事本文の抽出に失敗しました。スキップします", "url", rssArticle.URL, "error", err)
			if saveErr := firestoreClient.SaveRejectedArticle(ctx, rssArticle.URL, config.ReasonContentExtractionFailed, nil); saveErr != nil {
				logger.Error("却下記事の保存に失敗", "url", rssArticle.URL, "error", saveErr)
			}
			continue
		}

		title := rssArticle.Title
		if extractedTitle != "" {
			title = extractedTitle
		}

		configArticle := &config.Article{
			Title:         title,
			URL:           rssArticle.URL,
			PublishedDate: rssArticle.PublishedDate,
			SourceFeed:    rssArticle.SourceFeed,
			ContentText:   extractedText,
			FetchedAt:     rssArticle.FetchedAt,
		}

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

		if !evaluation.IsRelevant {
			logger.Debug("関連性がない記事を却下", "url", rssArticle.URL, "score", evaluation.RelevanceScore)
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
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "関連性のある記事が見つかりませんでした\n")
		return
	}

	sort.Slice(evaluatedArticles, func(i, j int) bool {
		return evaluatedArticles[i].RelevanceScore > evaluatedArticles[j].RelevanceScore
	})

	maxArticles := cfg.NotificationSettings.MaxArticles
	if len(evaluatedArticles) > maxArticles {
		evaluatedArticles = evaluatedArticles[:maxArticles]
	}

	logger.Info("上位記事を選択しました", "count", len(evaluatedArticles))

	articlesByURL := make(map[string]rss.Article, len(filteredArticles))
	for _, article := range filteredArticles {
		articlesByURL[article.URL] = article
	}

	discordArticles := make([]discord.Article, len(evaluatedArticles))
	for i, eval := range evaluatedArticles {
		article, ok := articlesByURL[eval.ArticleURL]
		sourceFeed := "Unknown"
		if ok {
			sourceFeed = article.SourceFeed
		}

		discordArticles[i] = discord.Article{
			Title:       getArticleTitle(articlesByURL, eval.ArticleURL),
			Description: eval.Summary,
			URL:         eval.ArticleURL,
			Relevance:   eval.RelevanceScore,
			Topics:      eval.MatchingTopics,
			Source:      sourceFeed,
		}
	}

	logger.Info("記事全体のサマリーを生成中", "articleCount", len(evaluatedArticles))

	llmArticles := make([]llm.ArticleForSummary, len(evaluatedArticles))
	for i, eval := range evaluatedArticles {
		llmArticles[i] = llm.ArticleForSummary{
			Title:          getArticleTitle(articlesByURL, eval.ArticleURL),
			Summary:        eval.Summary,
			RelevanceScore: eval.RelevanceScore,
			MatchingTopics: eval.MatchingTopics,
		}
	}

	summaryResult, err := llmEvaluator.GenerateArticlesSummary(ctx, llmArticles)
	if err != nil {
		logger.Warn("サマリー生成に失敗しました。サマリーなしで通知します", "error", err)
		summaryResult = nil
	} else {
		logger.Info("サマリー生成に成功しました")
	}

	var discordSummary *discord.ArticlesSummary
	if summaryResult != nil {
		discordSummary = &discord.ArticlesSummary{
			OverallSummary:  summaryResult.OverallSummary,
			MustRead:        summaryResult.MustRead,
			Recommendations: summaryResult.Recommendations,
		}
	}

	logger.Info("Discordに通知中", "articleCount", len(discordArticles))

	date := time.Now().Format("2006-01-02")
	messageID, err := discordClient.PostArticles(ctx, discordArticles, date, discordSummary)
	if err != nil {
		handleError(w, logger, http.StatusInternalServerError, "Discord通知に失敗", err)
		return
	}

	logger.Info("Discordへの通知に成功しました", "messageID", messageID)

	logger.Info("通知済み記事をFirestoreに保存中")
	for _, eval := range evaluatedArticles {
		title := getArticleTitle(articlesByURL, eval.ArticleURL)
		if err := firestoreClient.SaveNotifiedArticle(ctx, eval.ArticleURL, messageID, title, eval.RelevanceScore); err != nil {
			logger.Error("通知済み記事の保存に失敗", "url", eval.ArticleURL, "error", err)
		}
	}

	logger.Info("通知済み記事の保存完了")

	// HTTPレスポンスを返す
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "記事キュレーション処理が完了しました。%d件の記事を通知しました。\n", len(discordArticles))
}
