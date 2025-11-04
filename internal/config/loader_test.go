package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadFromFile(t *testing.T) {
	// テスト用の一時ファイルを作成
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	testConfig := `{
		"rss_sources": [
			{
				"url": "https://dev.to/feed",
				"name": "Dev.to",
				"enabled": true
			}
		],
		"interests": [
			{
				"topic": "Go",
				"aliases": ["Golang"],
				"priority": "high"
			}
		],
		"notification_settings": {
			"max_articles": 5,
			"min_articles": 3,
			"min_relevance_score": 70
		}
	}`

	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("テスト設定ファイルの作成に失敗: %v", err)
	}

	// ローダーを作成してファイルを読み込み
	loader := NewLoader()
	config, err := loader.Load(context.Background(), configPath)
	if err != nil {
		t.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// 検証
	if len(config.RSSSources) != 1 {
		t.Errorf("RSSソース数が不正: 期待=1, 実際=%d", len(config.RSSSources))
	}

	if config.RSSSources[0].Name != "Dev.to" {
		t.Errorf("RSSソース名が不正: 期待=Dev.to, 実際=%s", config.RSSSources[0].Name)
	}

	if len(config.Interests) != 1 {
		t.Errorf("興味トピック数が不正: 期待=1, 実際=%d", len(config.Interests))
	}

	if config.Interests[0].Topic != "Go" {
		t.Errorf("トピック名が不正: 期待=Go, 実際=%s", config.Interests[0].Topic)
	}

	if config.NotificationSettings.MaxArticles != 5 {
		t.Errorf("最大記事数が不正: 期待=5, 実際=%d", config.NotificationSettings.MaxArticles)
	}
}

func TestLoader_LoadFromURL(t *testing.T) {
	// テスト用のHTTPサーバーを起動
	testConfig := `{
		"rss_sources": [
			{
				"url": "https://dev.to/feed",
				"name": "Dev.to",
				"enabled": true
			}
		],
		"interests": [
			{
				"topic": "Kubernetes",
				"priority": "medium"
			}
		],
		"notification_settings": {
			"max_articles": 5,
			"min_articles": 3,
			"min_relevance_score": 70
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testConfig))
	}))
	defer server.Close()

	// ローダーを作成してURLから読み込み
	loader := NewLoader()
	config, err := loader.Load(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("URLからの設定読み込みに失敗: %v", err)
	}

	// 検証
	if len(config.Interests) != 1 {
		t.Errorf("興味トピック数が不正: 期待=1, 実際=%d", len(config.Interests))
	}

	if config.Interests[0].Topic != "Kubernetes" {
		t.Errorf("トピック名が不正: 期待=Kubernetes, 実際=%s", config.Interests[0].Topic)
	}
}

func TestConvertToRawGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GitHub blob URL",
			input:    "https://github.com/user/repo/blob/main/config.json",
			expected: "https://raw.githubusercontent.com/user/repo/main/config.json",
		},
		{
			name:     "GitHub blob URL with branch",
			input:    "https://github.com/user/repo/blob/develop/config.json",
			expected: "https://raw.githubusercontent.com/user/repo/develop/config.json",
		},
		{
			name:     "Already raw URL",
			input:    "https://raw.githubusercontent.com/user/repo/main/config.json",
			expected: "https://raw.githubusercontent.com/user/repo/main/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToRawGitHubURL(tt.input)
			if result != tt.expected {
				t.Errorf("変換結果が不正:\n期待=%s\n実際=%s", tt.expected, result)
			}
		})
	}
}
