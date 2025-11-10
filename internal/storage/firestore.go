// Package storage はFirestoreを使用した記事の重複排除追跡を提供します
package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cloud.google.com/go/firestore"
)

// Client はFirestoreクライアントをラップします
type Client struct {
	client *firestore.Client
}

// NewClient は新しいFirestoreクライアントを作成します
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("firestore client creation failed: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

// Close はFirestoreクライアントを閉じます
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// GetClient は内部のFirestoreクライアントを返します（テスト用）
func (c *Client) GetClient() *firestore.Client {
	return c.client
}

// urlToDocID はURLをFirestoreドキュメントIDに変換します
// SHA256ハッシュを使用することで、以下の問題を解決します：
// - クエリパラメータ（?、&、=）や特殊文字への対応
// - Firestoreのドキュメント文字列制限（1,500バイト）への対応
// - URLエンコード文字への対応
func urlToDocID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}
