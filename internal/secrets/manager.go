// Package secrets はGoogle Cloud Secret Managerとの統合を提供します
package secrets

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// Manager はSecret Managerからシークレットを取得するためのインターフェース
type Manager interface {
	GetSecret(ctx context.Context, secretName string) (string, error)
	Close() error
}

// manager はSecret Managerクライアントの実装
type manager struct {
	client    *secretmanager.Client
	projectID string
}

// NewManager は新しいSecret Managerクライアントを作成します
func NewManager(ctx context.Context, projectID string) (Manager, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Secret Managerクライアントの作成に失敗しました: %w", err)
	}

	return &manager{
		client:    client,
		projectID: projectID,
	}, nil
}

// GetSecret は指定された名前のシークレットの最新バージョンを取得します
func (m *manager) GetSecret(ctx context.Context, secretName string) (string, error) {
	// シークレットのフルパスを構築
	// フォーマット: projects/{project}/secrets/{secret}/versions/latest
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", m.projectID, secretName)

	// シークレットにアクセス
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := m.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("シークレット '%s' の取得に失敗しました: %w", secretName, err)
	}

	// シークレットデータを文字列として返す
	return string(result.Payload.Data), nil
}

// Close はSecret Managerクライアントをクローズします
func (m *manager) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}

// mockManager はテスト用のモックSecret Manager
type mockManager struct {
	secrets map[string]string
}

// NewMockManager はテスト用のモックSecret Managerを作成します
func NewMockManager(secrets map[string]string) Manager {
	return &mockManager{
		secrets: secrets,
	}
}

// GetSecret はモックシークレットを返します
func (m *mockManager) GetSecret(ctx context.Context, secretName string) (string, error) {
	secret, ok := m.secrets[secretName]
	if !ok {
		return "", fmt.Errorf("シークレット '%s' が見つかりません", secretName)
	}
	return secret, nil
}

// Close は何もしません（モック用）
func (m *mockManager) Close() error {
	return nil
}
