package contract

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kaka0913/discord-article-bot/internal/discord"
	"github.com/kaka0913/discord-article-bot/internal/logging"
)

// TestDiscordWebhookValidPayload は有効なペイロードのテスト
func TestDiscordWebhookValidPayload(t *testing.T) {
	// テスト用のサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content-Typeを検証
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got: %s", r.Header.Get("Content-Type"))
		}

		// リクエストボディをパース
		var payload discord.WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// 制約を検証
		if len(payload.Embeds) > 10 {
			t.Errorf("Too many embeds: %d (max 10)", len(payload.Embeds))
		}

		for i, embed := range payload.Embeds {
			if len(embed.Title) > 256 {
				t.Errorf("Embed %d: title too long (%d chars, max 256)", i, len(embed.Title))
			}
			if len(embed.Description) > 4096 {
				t.Errorf("Embed %d: description too long (%d chars, max 4096)", i, len(embed.Description))
			}
			if embed.Color < 0 || embed.Color > 16777215 {
				t.Errorf("Embed %d: invalid color %d (must be 0-16777215)", i, embed.Color)
			}
			for j, field := range embed.Fields {
				if len(field.Name) > 256 {
					t.Errorf("Embed %d, field %d: name too long (%d chars, max 256)", i, j, len(field.Name))
				}
				if len(field.Value) > 1024 {
					t.Errorf("Embed %d, field %d: value too long (%d chars, max 1024)", i, j, len(field.Value))
				}
			}
			if embed.Footer != nil && len(embed.Footer.Text) > 2048 {
				t.Errorf("Embed %d: footer too long (%d chars, max 2048)", i, len(embed.Footer.Text))
			}
		}

		// モックDiscord応答を返す
		w.WriteHeader(http.StatusOK)
		response := discord.WebhookResponse{
			ID:        "1234567890123456789",
			Type:      0,
			Content:   payload.Content,
			ChannelID: "987654321098765432",
			Embeds:    payload.Embeds,
			Timestamp: time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Discordクライアントをテスト
	logger := logging.NewLogger()
	client := discord.NewClient(server.URL, logger)

	articles := []discord.Article{
		{
			Title:       "Building Microservices with Go and Kubernetes",
			Description: "A comprehensive guide to building scalable microservices using Go and deploying them on Kubernetes clusters with best practices.",
			URL:         "https://dev.to/example/building-microservices",
			Relevance:   95,
			Topics:      []string{"Go", "Kubernetes"},
			Source:      "Dev.to",
		},
		{
			Title:       "WebAssembly Performance Optimization Tips",
			Description: "Learn advanced techniques for optimizing WebAssembly modules to achieve near-native performance in web browsers.",
			URL:         "https://zenn.dev/example/wasm-optimization",
			Relevance:   88,
			Topics:      []string{"WebAssembly", "Rust"},
			Source:      "Zenn",
		},
		{
			Title:       "Rust Async Runtime Internals",
			Description: "Deep dive into Tokio runtime architecture and how async/await works under the hood in Rust applications.",
			URL:         "https://hashnode.dev/example/rust-async",
			Relevance:   82,
			Topics:      []string{"Rust", "Async"},
			Source:      "Hashnode",
		},
	}

	messageID, err := client.PostArticles(context.Background(), articles, "2025-10-27")
	if err != nil {
		t.Fatalf("Failed to post articles: %v", err)
	}

	if messageID != "1234567890123456789" {
		t.Errorf("Expected message ID: 1234567890123456789, got: %s", messageID)
	}
}

// TestDiscordWebhookMaxEmbeds は最大埋め込み数（10個）のテスト
func TestDiscordWebhookMaxEmbeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload discord.WebhookPayload
		json.NewDecoder(r.Body).Decode(&payload)

		if len(payload.Embeds) != 10 {
			t.Errorf("Expected 10 embeds, got: %d", len(payload.Embeds))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(discord.WebhookResponse{
			ID:        "test-message-id",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	logger := logging.NewLogger()
	client := discord.NewClient(server.URL, logger)

	// 10個の記事を生成
	articles := make([]discord.Article, 10)
	for i := 0; i < 10; i++ {
		articles[i] = discord.Article{
			Title:       "Test Article",
			Description: "Test Description",
			URL:         "https://example.com",
			Relevance:   90,
			Topics:      []string{"Test"},
			Source:      "Test Source",
		}
	}

	_, err := client.PostArticles(context.Background(), articles, "2025-10-27")
	if err != nil {
		t.Fatalf("Failed to post 10 embeds: %v", err)
	}
}

// TestDiscordWebhookTitleTooLong はタイトルが長すぎる場合のテスト
func TestDiscordWebhookTitleTooLong(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload discord.WebhookPayload
		json.NewDecoder(r.Body).Decode(&payload)

		// タイトルが切り詰められていることを確認
		if len(payload.Embeds) > 0 {
			title := payload.Embeds[0].Title
			if len([]rune(title)) > 256 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(discord.ErrorResponse{
					Code:    50035,
					Message: "Invalid Form Body",
				})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(discord.WebhookResponse{
			ID:        "test-message-id",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	logger := logging.NewLogger()
	client := discord.NewClient(server.URL, logger)

	// 257文字のタイトルを持つ記事
	articles := []discord.Article{
		{
			Title:       strings.Repeat("a", 257),
			Description: "Test",
			URL:         "https://example.com",
			Relevance:   90,
			Topics:      []string{"Test"},
			Source:      "Test",
		},
	}

	// フォーマッターが自動的に切り詰めるので、エラーにならないはず
	_, err := client.PostArticles(context.Background(), articles, "2025-10-27")
	if err != nil {
		t.Fatalf("Expected success with truncated title, got error: %v", err)
	}
}

// TestDiscordWebhookDescriptionTooLong は説明が長すぎる場合のテスト
func TestDiscordWebhookDescriptionTooLong(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload discord.WebhookPayload
		json.NewDecoder(r.Body).Decode(&payload)

		// 説明が切り詰められていることを確認
		if len(payload.Embeds) > 0 {
			desc := payload.Embeds[0].Description
			if len([]rune(desc)) > 4096 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(discord.ErrorResponse{
					Code:    50035,
					Message: "Invalid Form Body",
				})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(discord.WebhookResponse{
			ID:        "test-message-id",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	logger := logging.NewLogger()
	client := discord.NewClient(server.URL, logger)

	// 4097文字の説明を持つ記事
	articles := []discord.Article{
		{
			Title:       "Test",
			Description: strings.Repeat("a", 4097),
			URL:         "https://example.com",
			Relevance:   90,
			Topics:      []string{"Test"},
			Source:      "Test",
		},
	}

	// フォーマッターが自動的に切り詰めるので、エラーにならないはず
	_, err := client.PostArticles(context.Background(), articles, "2025-10-27")
	if err != nil {
		t.Fatalf("Expected success with truncated description, got error: %v", err)
	}
}

// TestDiscordWebhookInvalidWebhook は無効なwebhookのテスト
func TestDiscordWebhookInvalidWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(discord.ErrorResponse{
			Code:    10015,
			Message: "Unknown Webhook",
		})
	}))
	defer server.Close()

	logger := logging.NewLogger()
	client := discord.NewClient(server.URL, logger)

	articles := []discord.Article{
		{
			Title:       "Test",
			Description: "Test",
			URL:         "https://example.com",
			Relevance:   90,
			Topics:      []string{"Test"},
			Source:      "Test",
		},
	}

	_, err := client.PostArticles(context.Background(), articles, "2025-10-27")
	if err == nil {
		t.Fatal("Expected error for invalid webhook, got nil")
	}

	// エラーメッセージに"Unknown Webhook"が含まれることを確認
	if !strings.Contains(err.Error(), "Unknown Webhook") && !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected error message to contain 'Unknown Webhook' or '404', got: %v", err)
	}
}

// TestDiscordWebhookRateLimit はレート制限のテスト
func TestDiscordWebhookRateLimit(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// 最初のリクエストは429を返す
		if requestCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			// エラーレスポンスボディに直接retry_afterを含める
			w.Write([]byte(`{"message": "You are being rate limited.", "retry_after": 0.1, "global": false}`))
			return
		}

		// 2回目のリクエストは成功
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(discord.WebhookResponse{
			ID:        "test-message-id",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	logger := logging.NewLogger()
	client := discord.NewClient(server.URL, logger)

	articles := []discord.Article{
		{
			Title:       "Test",
			Description: "Test",
			URL:         "https://example.com",
			Relevance:   90,
			Topics:      []string{"Test"},
			Source:      "Test",
		},
	}

	// レート制限後に再試行して成功することを確認
	messageID, err := client.PostArticles(context.Background(), articles, "2025-10-27")
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	if messageID != "test-message-id" {
		t.Errorf("Expected message ID: test-message-id, got: %s", messageID)
	}

	// 2回のリクエストが行われたことを確認
	if requestCount != 2 {
		t.Errorf("Expected 2 requests (1 rate limited + 1 retry), got: %d", requestCount)
	}
}

// TestDiscordWebhookTooManyEmbeds は埋め込みが10個を超える場合のテスト
func TestDiscordWebhookTooManyEmbeds(t *testing.T) {
	logger := logging.NewLogger()
	client := discord.NewClient("http://example.com", logger)

	// 11個の記事を生成
	articles := make([]discord.Article, 11)
	for i := 0; i < 11; i++ {
		articles[i] = discord.Article{
			Title:       "Test Article",
			Description: "Test Description",
			URL:         "https://example.com",
			Relevance:   90,
			Topics:      []string{"Test"},
			Source:      "Test Source",
		}
	}

	_, err := client.PostArticles(context.Background(), articles, "2025-10-27")
	if err == nil {
		t.Fatal("Expected error for too many embeds, got nil")
	}

	if !strings.Contains(err.Error(), "too many embeds") {
		t.Errorf("Expected error message to contain 'too many embeds', got: %v", err)
	}
}
