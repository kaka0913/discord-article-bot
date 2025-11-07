package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kaka0913/discord-article-bot/internal/config"
	"github.com/kaka0913/discord-article-bot/internal/logging"
)

const (
	// RejectedArticlesCollection は却下済み記事を保存するコレクション名
	RejectedArticlesCollection = "rejected_articles"

	// RejectedArticleTTLDays は却下済み記事のTTL（日数）
	RejectedArticleTTLDays = 30
)

// SaveRejectedArticle は却下された記事をFirestoreに保存します
func (c *Client) SaveRejectedArticle(ctx context.Context, articleURL, reason string, relevanceScore *int) error {
	// URLをFirestoreドキュメントIDに変換（SHA256ハッシュ）
	docID := urlToDocID(articleURL)

	docRef := c.client.Collection(RejectedArticlesCollection).Doc(docID)

	rejectedArticle := map[string]interface{}{
		"evaluated_at": firestore.ServerTimestamp,
		"reason":       reason,
	}

	if relevanceScore != nil {
		rejectedArticle["relevance_score"] = *relevanceScore
	} else {
		rejectedArticle["relevance_score"] = nil
	}

	_, err := docRef.Set(ctx, rejectedArticle)
	if err != nil {
		return fmt.Errorf("failed to save rejected article: %w", err)
	}

	return nil
}

// IsArticleRejected は記事が既に却下済みかどうかをチェックします
func (c *Client) IsArticleRejected(ctx context.Context, articleURL string) (bool, error) {
	// URLをFirestoreドキュメントIDに変換
	docID := urlToDocID(articleURL)

	docRef := c.client.Collection(RejectedArticlesCollection).Doc(docID)
	doc, err := docRef.Get(ctx)

	if err != nil {
		// NotFoundエラーは記事がまだ却下されていないことを意味する
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to check rejected article: %w", err)
	}

	// ドキュメントが存在する場合、TTLチェックを実行
	if doc.Exists() {
		var rejectedArticle config.RejectedArticle
		if err := doc.DataTo(&rejectedArticle); err != nil {
			// データのパースに失敗した場合、破損データの可能性があるのでログを出力
			logger := logging.FromContext(ctx)
			logger.Warn("Failed to parse rejected article data", "url", articleURL, "error", err)
			return false, fmt.Errorf("failed to parse rejected article: %w", err)
		}

		// TTLチェック: 30日以上経過している場合は古いデータとして扱う
		if time.Since(rejectedArticle.EvaluatedAt) > RejectedArticleTTLDays*24*time.Hour {
			// TTL期限切れの場合はfalseを返す（新しい記事として扱う）
			return false, nil
		}

		return true, nil
	}

	return false, nil
}
