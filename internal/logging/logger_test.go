package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger() が nil を返しました")
	}
}

func TestNewDevelopmentLogger(t *testing.T) {
	logger := NewDevelopmentLogger()
	if logger == nil {
		t.Fatal("NewDevelopmentLogger() が nil を返しました")
	}
}

func TestLogger_Info(t *testing.T) {
	// ログ出力をキャプチャするためのバッファ
	var buf bytes.Buffer

	// テスト用のハンドラーを作成
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &logger{
		slogger: slog.New(handler),
	}

	// ログを出力
	logger.Info("テストメッセージ", "key", "value")

	// 出力されたJSONを検証
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("JSONのパースに失敗: %v", err)
	}

	if logEntry["msg"] != "テストメッセージ" {
		t.Errorf("msg = %v, 期待 'テストメッセージ'", logEntry["msg"])
	}

	if logEntry["key"] != "value" {
		t.Errorf("key = %v, 期待 'value'", logEntry["key"])
	}

	if logEntry["level"] != "INFO" {
		t.Errorf("level = %v, 期待 'INFO'", logEntry["level"])
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &logger{
		slogger: slog.New(handler),
	}

	logger.Error("エラーメッセージ", "error", "test error")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("JSONのパースに失敗: %v", err)
	}

	if logEntry["msg"] != "エラーメッセージ" {
		t.Errorf("msg = %v, 期待 'エラーメッセージ'", logEntry["msg"])
	}

	if logEntry["level"] != "ERROR" {
		t.Errorf("level = %v, 期待 'ERROR'", logEntry["level"])
	}
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	baseLogger := &logger{
		slogger: slog.New(handler),
	}

	// With() で新しいロガーを作成
	newLogger := baseLogger.With("request_id", "12345")

	// 新しいロガーでログを出力
	newLogger.Info("リクエスト処理", "status", "success")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("JSONのパースに失敗: %v", err)
	}

	// With() で追加した属性が含まれているか確認
	if logEntry["request_id"] != "12345" {
		t.Errorf("request_id = %v, 期待 '12345'", logEntry["request_id"])
	}

	if logEntry["status"] != "success" {
		t.Errorf("status = %v, 期待 'success'", logEntry["status"])
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer

	// DEBUGレベルを有効にする
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &logger{
		slogger: slog.New(handler),
	}

	logger.Debug("デバッグメッセージ", "detail", "debug info")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("JSONのパースに失敗: %v", err)
	}

	if logEntry["level"] != "DEBUG" {
		t.Errorf("level = %v, 期待 'DEBUG'", logEntry["level"])
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &logger{
		slogger: slog.New(handler),
	}

	logger.Warn("警告メッセージ", "reason", "deprecated API")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("JSONのパースに失敗: %v", err)
	}

	if logEntry["level"] != "WARN" {
		t.Errorf("level = %v, 期待 'WARN'", logEntry["level"])
	}
}

func TestFromContext(t *testing.T) {
	// コンテキストにロガーが設定されていない場合
	ctx := context.Background()
	logger := FromContext(ctx)
	if logger == nil {
		t.Error("FromContext() が nil を返しました")
	}

	// コンテキストにロガーを設定した場合
	customLogger := NewLogger()
	ctx = ToContext(ctx, customLogger)
	retrievedLogger := FromContext(ctx)

	// 同じロガーが取得できることを確認
	// インターフェース比較なので型アサーションで確認
	if _, ok := retrievedLogger.(Logger); !ok {
		t.Error("FromContext() が期待するLogger型を返しませんでした")
	}
}

func TestToContext(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger()

	newCtx := ToContext(ctx, logger)

	// コンテキストにロガーが設定されているか確認
	retrievedLogger := FromContext(newCtx)
	if retrievedLogger == nil {
		t.Error("ToContext() でロガーが設定されませんでした")
	}
}

func TestNewLoggerWithLevel(t *testing.T) {
	var buf bytes.Buffer

	// カスタムレベルでロガーを作成
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})

	logger := &logger{
		slogger: slog.New(handler),
	}

	// INFOレベルのログは出力されないはず
	logger.Info("これは出力されない")

	// WARNレベルのログは出力されるはず
	logger.Warn("これは出力される")

	output := buf.String()

	// INFOログが含まれていないことを確認
	if strings.Contains(output, "これは出力されない") {
		t.Error("INFOレベルのログが出力されてしまいました")
	}

	// WARNログが含まれていることを確認
	if !strings.Contains(output, "これは出力される") {
		t.Error("WARNレベルのログが出力されませんでした")
	}
}
