package contract

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kaka0913/discord-article-bot/internal/config"
	"github.com/kaka0913/discord-article-bot/internal/storage"
)

// TestMain はテストスイートの前後でセットアップとクリーンアップを実行します
func TestMain(m *testing.M) {
	// Firestoreエミュレータの環境変数をチェック
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		// エミュレータが起動していない場合はテストをスキップ
		os.Exit(0)
	}

	// テストを実行
	code := m.Run()

	os.Exit(code)
}

// setupTestClient はテスト用のFirestoreクライアントを作成します
func setupTestClient(t *testing.T) *storage.Client {
	ctx := context.Background()
	projectID := "test-project"

	client, err := storage.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to create Firestore client: %v", err)
	}

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close Firestore client: %v", err)
		}
	})

	return client
}

// cleanupCollection はテストデータをクリーンアップします
func cleanupCollection(t *testing.T, client *storage.Client, collection string) {
	ctx := context.Background()
	docs, err := client.GetClient().Collection(collection).Documents(ctx).GetAll()
	if err != nil {
		t.Logf("Warning: Failed to get documents for cleanup: %v", err)
		return
	}

	for _, doc := range docs {
		if _, err := doc.Ref.Delete(ctx); err != nil {
			t.Logf("Warning: Failed to delete document %s: %v", doc.Ref.ID, err)
		}
	}
}

// TestSaveAndCheckNotifiedArticle は通知済み記事の保存と確認をテストします
func TestSaveAndCheckNotifiedArticle(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// テストデータをクリーンアップ
	t.Cleanup(func() {
		cleanupCollection(t, client, storage.NotifiedArticlesCollection)
	})

	testCases := []struct {
		name             string
		articleURL       string
		discordMessageID string
		articleTitle     string
		relevanceScore   int
	}{
		{
			name:             "通常の記事URLを保存",
			articleURL:       "https://dev.to/example/building-microservices",
			discordMessageID: "1234567890123456789",
			articleTitle:     "Building Microservices with Go",
			relevanceScore:   85,
		},
		{
			name:             "特殊文字を含むURLを保存",
			articleURL:       "https://example.com/article?id=123&lang=ja",
			discordMessageID: "9876543210987654321",
			articleTitle:     "Special Characters Test",
			relevanceScore:   90,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 最初は通知されていないことを確認
			notified, err := client.IsArticleNotified(ctx, tc.articleURL)
			if err != nil {
				t.Fatalf("IsArticleNotified failed: %v", err)
			}
			if notified {
				t.Error("Expected article to not be notified initially")
			}

			// 記事を保存
			err = client.SaveNotifiedArticle(ctx, tc.articleURL, tc.discordMessageID, tc.articleTitle, tc.relevanceScore)
			if err != nil {
				t.Fatalf("SaveNotifiedArticle failed: %v", err)
			}

			// 保存後に通知済みであることを確認
			notified, err = client.IsArticleNotified(ctx, tc.articleURL)
			if err != nil {
				t.Fatalf("IsArticleNotified failed after save: %v", err)
			}
			if !notified {
				t.Error("Expected article to be notified after save")
			}
		})
	}
}

// TestNotifiedArticleTTL は通知済み記事のTTL機能をテストします
func TestNotifiedArticleTTL(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// テストデータをクリーンアップ
	t.Cleanup(func() {
		cleanupCollection(t, client, storage.NotifiedArticlesCollection)
	})

	articleURL := "https://dev.to/example/old-article"

	// 31日前の古い記事を手動で保存（TTL期限切れ）
	docID := storage.UrlToDocID(articleURL)
	oldArticle := config.NotifiedArticle{
		NotifiedAt:       time.Now().AddDate(0, 0, -31), // 31日前
		DiscordMessageID: "1234567890123456789",
		ArticleTitle:     "Old Article",
		RelevanceScore:   75,
	}

	_, err := client.GetClient().Collection(storage.NotifiedArticlesCollection).Doc(docID).Set(ctx, oldArticle)
	if err != nil {
		t.Fatalf("Failed to save old article: %v", err)
	}

	// TTL期限切れの記事は通知されていないとして扱われるべき
	notified, err := client.IsArticleNotified(ctx, articleURL)
	if err != nil {
		t.Fatalf("IsArticleNotified failed: %v", err)
	}
	if notified {
		t.Error("Expected old article to be treated as not notified (TTL expired)")
	}
}

// TestSaveAndCheckRejectedArticle は却下済み記事の保存と確認をテストします
func TestSaveAndCheckRejectedArticle(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// テストデータをクリーンアップ
	t.Cleanup(func() {
		cleanupCollection(t, client, storage.RejectedArticlesCollection)
	})

	testCases := []struct {
		name           string
		articleURL     string
		reason         string
		relevanceScore *int
	}{
		{
			name:           "低い関連性スコアで却下",
			articleURL:     "https://dev.to/example/low-relevance",
			reason:         config.ReasonLowRelevance,
			relevanceScore: intPtr(35),
		},
		{
			name:           "トピックマッチなしで却下",
			articleURL:     "https://dev.to/example/no-topic-match",
			reason:         config.ReasonNoTopicMatch,
			relevanceScore: intPtr(40),
		},
		{
			name:           "コンテンツ抽出失敗で却下",
			articleURL:     "https://medium.com/example/paywalled",
			reason:         config.ReasonContentExtractionFailed,
			relevanceScore: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 最初は却下されていないことを確認
			rejected, err := client.IsArticleRejected(ctx, tc.articleURL)
			if err != nil {
				t.Fatalf("IsArticleRejected failed: %v", err)
			}
			if rejected {
				t.Error("Expected article to not be rejected initially")
			}

			// 記事を保存
			err = client.SaveRejectedArticle(ctx, tc.articleURL, tc.reason, tc.relevanceScore)
			if err != nil {
				t.Fatalf("SaveRejectedArticle failed: %v", err)
			}

			// 保存後に却下済みであることを確認
			rejected, err = client.IsArticleRejected(ctx, tc.articleURL)
			if err != nil {
				t.Fatalf("IsArticleRejected failed after save: %v", err)
			}
			if !rejected {
				t.Error("Expected article to be rejected after save")
			}
		})
	}
}

// TestRejectedArticleTTL は却下済み記事のTTL機能をテストします
func TestRejectedArticleTTL(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// テストデータをクリーンアップ
	t.Cleanup(func() {
		cleanupCollection(t, client, storage.RejectedArticlesCollection)
	})

	articleURL := "https://dev.to/example/old-rejected-article"

	// 31日前の古い却下記事を手動で保存（TTL期限切れ）
	docID := storage.UrlToDocID(articleURL)
	oldArticle := config.RejectedArticle{
		EvaluatedAt:    time.Now().AddDate(0, 0, -31), // 31日前
		Reason:         config.ReasonLowRelevance,
		RelevanceScore: intPtr(30),
	}

	_, err := client.GetClient().Collection(storage.RejectedArticlesCollection).Doc(docID).Set(ctx, oldArticle)
	if err != nil {
		t.Fatalf("Failed to save old rejected article: %v", err)
	}

	// TTL期限切れの記事は却下されていないとして扱われるべき
	rejected, err := client.IsArticleRejected(ctx, articleURL)
	if err != nil {
		t.Fatalf("IsArticleRejected failed: %v", err)
	}
	if rejected {
		t.Error("Expected old rejected article to be treated as not rejected (TTL expired)")
	}
}

// TestDuplicateNotificationPrevention は重複通知を防ぐテストです
func TestDuplicateNotificationPrevention(t *testing.T) {
	client := setupTestClient(t)
	ctx := context.Background()

	// テストデータをクリーンアップ
	t.Cleanup(func() {
		cleanupCollection(t, client, storage.NotifiedArticlesCollection)
	})

	articleURL := "https://dev.to/example/duplicate-test"

	// 記事を最初に保存
	err := client.SaveNotifiedArticle(ctx, articleURL, "1111111111111111111", "Duplicate Test", 80)
	if err != nil {
		t.Fatalf("First SaveNotifiedArticle failed: %v", err)
	}

	// 通知済みであることを確認
	notified, err := client.IsArticleNotified(ctx, articleURL)
	if err != nil {
		t.Fatalf("IsArticleNotified failed: %v", err)
	}
	if !notified {
		t.Fatal("Expected article to be notified")
	}

	// 同じ記事を再度保存しても成功すべき（べき等性）
	err = client.SaveNotifiedArticle(ctx, articleURL, "2222222222222222222", "Duplicate Test", 85)
	if err != nil {
		t.Errorf("Second SaveNotifiedArticle should succeed (idempotent): %v", err)
	}
}

// intPtr はint値へのポインタを返すヘルパー関数です
func intPtr(i int) *int {
	return &i
}
