package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kaka0913/discord-article-bot/internal/config"
)

const (
	// NotifiedArticlesCollection は通知済み記事を保存するコレクション名
	NotifiedArticlesCollection = "notified_articles"

	// NotifiedArticleTTLDays は通知済み記事のTTL（日数）
	NotifiedArticleTTLDays = 30
)

// SaveNotifiedArticle は通知済み記事をFirestoreに保存します
func (c *Client) SaveNotifiedArticle(ctx context.Context, articleURL, discordMessageID, articleTitle string, relevanceScore int) error {
	// URLをFirestoreドキュメントIDに変換（/ を - に置換）
	docID := urlToDocID(articleURL)

	docRef := c.client.Collection(NotifiedArticlesCollection).Doc(docID)

	notifiedArticle := config.NotifiedArticle{
		NotifiedAt:       time.Now(),
		DiscordMessageID: discordMessageID,
		ArticleTitle:     articleTitle,
		RelevanceScore:   relevanceScore,
	}

	_, err := docRef.Set(ctx, notifiedArticle)
	if err != nil {
		return fmt.Errorf("failed to save notified article: %w", err)
	}

	return nil
}

// IsArticleNotified は記事が既に通知済みかどうかをチェックします
func (c *Client) IsArticleNotified(ctx context.Context, articleURL string) (bool, error) {
	// URLをFirestoreドキュメントIDに変換
	docID := urlToDocID(articleURL)

	docRef := c.client.Collection(NotifiedArticlesCollection).Doc(docID)
	doc, err := docRef.Get(ctx)

	if err != nil {
		// NotFoundエラーは記事がまだ通知されていないことを意味する
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check notified article: %w", err)
	}

	// ドキュメントが存在する場合、TTLチェックを実行
	if doc.Exists() {
		var notifiedArticle config.NotifiedArticle
		if err := doc.DataTo(&notifiedArticle); err != nil {
			return false, fmt.Errorf("failed to parse notified article: %w", err)
		}

		// TTLチェック: 30日以上経過している場合は古いデータとして扱う
		if time.Since(notifiedArticle.NotifiedAt) > NotifiedArticleTTLDays*24*time.Hour {
			// TTL期限切れの場合はfalseを返す（新しい記事として扱う）
			return false, nil
		}

		return true, nil
	}

	return false, nil
}

// urlToDocID はURLをFirestoreドキュメントIDに変換します
// Firestoreのドキュメント文字列制限に対応するため、/を--に置換
func urlToDocID(url string) string {
	// : を - に、/ を -- に置換
	docID := strings.ReplaceAll(url, "://", "-")
	docID = strings.ReplaceAll(docID, "/", "--")
	return docID
}
