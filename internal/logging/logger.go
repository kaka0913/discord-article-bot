// Package logging は構造化ログを提供します
package logging

import (
	"context"
	"log/slog"
	"os"
)

// Logger は構造化ログを出力するためのインターフェース
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// logger はslogベースのロガー実装
type logger struct {
	slogger *slog.Logger
}

// NewLogger は新しいロガーを作成します
func NewLogger() Logger {
	// JSON形式の構造化ログを標準出力に出力
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &logger{
		slogger: slog.New(handler),
	}
}

// NewLoggerWithLevel はログレベルを指定して新しいロガーを作成します
func NewLoggerWithLevel(level slog.Level) Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return &logger{
		slogger: slog.New(handler),
	}
}

// NewDevelopmentLogger は開発用のロガーを作成します（テキスト形式、DEBUGレベル）
func NewDevelopmentLogger() Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return &logger{
		slogger: slog.New(handler),
	}
}

// Debug はDEBUGレベルのログを出力します
func (l *logger) Debug(msg string, args ...any) {
	l.slogger.Debug(msg, args...)
}

// Info はINFOレベルのログを出力します
func (l *logger) Info(msg string, args ...any) {
	l.slogger.Info(msg, args...)
}

// Warn はWARNレベルのログを出力します
func (l *logger) Warn(msg string, args ...any) {
	l.slogger.Warn(msg, args...)
}

// Error はERRORレベルのログを出力します
func (l *logger) Error(msg string, args ...any) {
	l.slogger.Error(msg, args...)
}

// With は指定された属性を持つ新しいロガーを返します
func (l *logger) With(args ...any) Logger {
	return &logger{
		slogger: l.slogger.With(args...),
	}
}

// FromContext はコンテキストからロガーを取得します
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey{}).(Logger); ok {
		return logger
	}
	// デフォルトのロガーを返す
	return NewLogger()
}

// ToContext はコンテキストにロガーを設定します
func ToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// loggerKey はコンテキストのキー
type loggerKey struct{}
