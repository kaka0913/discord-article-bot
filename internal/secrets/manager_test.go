package secrets

import (
	"context"
	"testing"
)

func TestNewMockManager(t *testing.T) {
	secrets := map[string]string{
		"api-key":     "test-api-key",
		"webhook-url": "https://discord.com/api/webhooks/test",
	}

	manager := NewMockManager(secrets)
	if manager == nil {
		t.Fatal("NewMockManager() が nil を返しました")
	}
}

func TestMockManager_GetSecret(t *testing.T) {
	secrets := map[string]string{
		"api-key":     "test-api-key-123",
		"webhook-url": "https://discord.com/api/webhooks/test",
	}

	manager := NewMockManager(secrets)
	ctx := context.Background()

	tests := []struct {
		name       string
		secretName string
		wantValue  string
		wantErr    bool
	}{
		{
			name:       "存在するシークレット - api-key",
			secretName: "api-key",
			wantValue:  "test-api-key-123",
			wantErr:    false,
		},
		{
			name:       "存在するシークレット - webhook-url",
			secretName: "webhook-url",
			wantValue:  "https://discord.com/api/webhooks/test",
			wantErr:    false,
		},
		{
			name:       "存在しないシークレット",
			secretName: "non-existent",
			wantValue:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := manager.GetSecret(ctx, tt.secretName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSecret() エラー = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && value != tt.wantValue {
				t.Errorf("GetSecret() = %v, 期待 %v", value, tt.wantValue)
			}
		})
	}
}

func TestMockManager_Close(t *testing.T) {
	secrets := map[string]string{
		"test-key": "test-value",
	}

	manager := NewMockManager(secrets)
	err := manager.Close()

	if err != nil {
		t.Errorf("Close() エラー = %v, 期待 nil", err)
	}
}

func TestMockManager_MultipleGetSecret(t *testing.T) {
	secrets := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	manager := NewMockManager(secrets)
	ctx := context.Background()

	// 複数回GetSecretを呼び出す
	for i := 0; i < 3; i++ {
		value1, err := manager.GetSecret(ctx, "key1")
		if err != nil {
			t.Errorf("GetSecret(key1) エラー = %v", err)
		}
		if value1 != "value1" {
			t.Errorf("GetSecret(key1) = %v, 期待 value1", value1)
		}

		value2, err := manager.GetSecret(ctx, "key2")
		if err != nil {
			t.Errorf("GetSecret(key2) エラー = %v", err)
		}
		if value2 != "value2" {
			t.Errorf("GetSecret(key2) = %v, 期待 value2", value2)
		}
	}
}
