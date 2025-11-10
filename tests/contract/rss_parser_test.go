package contract

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kaka0913/discord-article-bot/internal/logging"
	"github.com/kaka0913/discord-article-bot/internal/rss"
)

func TestRSSParser_ParseRSS2Feed(t *testing.T) {
	// テスト用のロガーを設定
	ctx := logging.ToContext(context.Background(), logging.NewLogger())

	// RSS 2.0フィクスチャを読み込み
	fixtureData, err := os.ReadFile(filepath.Join("fixtures", "rss2_sample.xml"))
	if err != nil {
		t.Fatalf("フィクスチャファイルの読み込みに失敗: %v", err)
	}

	// パーサーを作成
	parser := rss.NewParser()

	// フィードをパース
	articles, err := parser.Parse(ctx, fixtureData, "Tech Blog Sample")
	if err != nil {
		t.Fatalf("RSS 2.0フィードのパースに失敗: %v", err)
	}

	// 記事数を確認
	expectedCount := 3
	if len(articles) != expectedCount {
		t.Errorf("記事数が期待値と異なる: got %d, want %d", len(articles), expectedCount)
	}

	// 最初の記事の内容を検証
	if len(articles) > 0 {
		firstArticle := articles[0]

		if firstArticle.Title != "Building Microservices with Go" {
			t.Errorf("記事タイトルが期待値と異なる: got %q, want %q",
				firstArticle.Title, "Building Microservices with Go")
		}

		expectedURL := "https://example.com/articles/building-microservices-with-go"
		if firstArticle.URL != expectedURL {
			t.Errorf("記事URLが期待値と異なる: got %q, want %q",
				firstArticle.URL, expectedURL)
		}

		if firstArticle.SourceFeed != "Tech Blog Sample" {
			t.Errorf("ソースフィード名が期待値と異なる: got %q, want %q",
				firstArticle.SourceFeed, "Tech Blog Sample")
		}

		if firstArticle.PublishedDate.IsZero() {
			t.Error("公開日時が設定されていない")
		}

		if firstArticle.FetchedAt.IsZero() {
			t.Error("取得日時が設定されていない")
		}
	}
}

func TestRSSParser_ParseAtomFeed(t *testing.T) {
	// テスト用のロガーを設定
	ctx := logging.ToContext(context.Background(), logging.NewLogger())

	// Atomフィクスチャを読み込み
	fixtureData, err := os.ReadFile(filepath.Join("fixtures", "atom_sample.xml"))
	if err != nil {
		t.Fatalf("フィクスチャファイルの読み込みに失敗: %v", err)
	}

	// パーサーを作成
	parser := rss.NewParser()

	// フィードをパース
	articles, err := parser.Parse(ctx, fixtureData, "Tech Blog Atom")
	if err != nil {
		t.Fatalf("Atomフィードのパースに失敗: %v", err)
	}

	// 記事数を確認
	expectedCount := 2
	if len(articles) != expectedCount {
		t.Errorf("記事数が期待値と異なる: got %d, want %d", len(articles), expectedCount)
	}

	// 最初の記事の内容を検証
	if len(articles) > 0 {
		firstArticle := articles[0]

		if firstArticle.Title != "Understanding Docker Containers" {
			t.Errorf("記事タイトルが期待値と異なる: got %q, want %q",
				firstArticle.Title, "Understanding Docker Containers")
		}

		expectedURL := "https://example.com/articles/understanding-docker-containers"
		if firstArticle.URL != expectedURL {
			t.Errorf("記事URLが期待値と異なる: got %q, want %q",
				firstArticle.URL, expectedURL)
		}

		if firstArticle.SourceFeed != "Tech Blog Atom" {
			t.Errorf("ソースフィード名が期待値と異なる: got %q, want %q",
				firstArticle.SourceFeed, "Tech Blog Atom")
		}

		if firstArticle.PublishedDate.IsZero() {
			t.Error("公開日時が設定されていない")
		}

		if firstArticle.FetchedAt.IsZero() {
			t.Error("取得日時が設定されていない")
		}
	}
}

func TestRSSParser_ParseInvalidFeed(t *testing.T) {
	// テスト用のロガーを設定
	ctx := logging.ToContext(context.Background(), logging.NewLogger())

	// 無効なXMLデータ
	invalidXML := []byte("<invalid>xml</data>")

	// パーサーを作成
	parser := rss.NewParser()

	// フィードをパース（エラーが期待される）
	_, err := parser.Parse(ctx, invalidXML, "Invalid Feed")
	if err == nil {
		t.Error("無効なフィードのパースでエラーが期待されたが、エラーが返されなかった")
	}
}

func TestRSSParser_ParseEmptyFeed(t *testing.T) {
	// テスト用のロガーを設定
	ctx := logging.ToContext(context.Background(), logging.NewLogger())

	// 空のRSSフィード
	emptyRSS := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Empty Feed</title>
    <link>https://example.com</link>
    <description>Empty RSS feed</description>
  </channel>
</rss>`)

	// パーサーを作成
	parser := rss.NewParser()

	// フィードをパース
	articles, err := parser.Parse(ctx, emptyRSS, "Empty Feed")
	if err != nil {
		t.Fatalf("空のフィードのパースに失敗: %v", err)
	}

	// 記事数を確認（0件であることを期待）
	if len(articles) != 0 {
		t.Errorf("記事数が期待値と異なる: got %d, want 0", len(articles))
	}
}
