package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Loader は設定ファイルを読み込むためのインターフェース
type Loader interface {
	Load(ctx context.Context, source string) (*Config, error)
}

// loader は設定ファイルローダーの実装
type loader struct {
	httpClient *http.Client
}

// NewLoader は新しい設定ローダーを作成します
func NewLoader() Loader {
	return &loader{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Load は指定されたソースから設定を読み込みます
// sourceはファイルパスまたはURL（HTTPSまたはGitHub URL）を指定できます
func (l *loader) Load(ctx context.Context, source string) (*Config, error) {
	var data []byte
	var err error

	// GitHub URLの場合はraw.githubusercontent.comに変換
	// より厳密な判定: github.com/ と /blob/ の両方が含まれる場合のみ変換
	if strings.Contains(source, "github.com/") && strings.Contains(source, "/blob/") {
		source = convertToRawGitHubURL(source)
	}

	// URLかローカルファイルかを判定
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		data, err = l.loadFromURL(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("URLから設定を読み込めませんでした: %w", err)
		}
	} else {
		data, err = l.loadFromFile(source)
		if err != nil {
			return nil, fmt.Errorf("ファイルから設定を読み込めませんでした: %w", err)
		}
	}

	// JSONをパース
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("設定のJSONパースに失敗しました: %w", err)
	}

	return &config, nil
}

// loadFromFile はローカルファイルから設定を読み込みます
func (l *loader) loadFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}
	return data, nil
}

// loadFromURL はHTTP/HTTPS URLから設定を読み込みます
func (l *loader) loadFromURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエスト作成エラー: %w", err)
	}

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTPステータスエラー: %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み込みエラー: %w", err)
	}

	return data, nil
}

// convertToRawGitHubURL はGitHub URLをraw.githubusercontent.com形式に変換します
// 例: https://github.com/user/repo/blob/main/config.json
//  -> https://raw.githubusercontent.com/user/repo/main/config.json
func convertToRawGitHubURL(url string) string {
	// github.com を raw.githubusercontent.com に置換
	url = strings.Replace(url, "github.com", "raw.githubusercontent.com", 1)
	// /blob/ を削除
	url = strings.Replace(url, "/blob/", "/", 1)
	return url
}
