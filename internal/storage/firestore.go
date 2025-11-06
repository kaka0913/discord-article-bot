// Package storage はFirestoreを使用した記事の重複排除追跡を提供します
package storage

import (
	"context"
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
