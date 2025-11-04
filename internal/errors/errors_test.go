package errors

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "メッセージのみのエラー",
			appError: &AppError{
				Type:    ErrorTypeConfig,
				Message: "設定の読み込みに失敗",
				Err:     nil,
			},
			expected: "[CONFIG] 設定の読み込みに失敗",
		},
		{
			name: "ラップされたエラー",
			appError: &AppError{
				Type:    ErrorTypeNetwork,
				Message: "API呼び出しに失敗",
				Err:     errors.New("connection timeout"),
			},
			expected: "[NETWORK] API呼び出しに失敗: connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, 期待 %v", result, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := Wrap(ErrorTypeInternal, "内部エラー", originalErr)

	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, 期待 %v", unwrapped, originalErr)
	}

	// errors.Is でラップされたエラーを検証
	if !errors.Is(appErr, originalErr) {
		t.Error("errors.Is() でラップされたエラーを検出できませんでした")
	}
}

func TestNew(t *testing.T) {
	err := New(ErrorTypeValidation, "バリデーションエラー")

	if err.Type != ErrorTypeValidation {
		t.Errorf("Type = %v, 期待 %v", err.Type, ErrorTypeValidation)
	}

	if err.Message != "バリデーションエラー" {
		t.Errorf("Message = %v, 期待 %v", err.Message, "バリデーションエラー")
	}

	if err.Err != nil {
		t.Errorf("Err = %v, 期待 nil", err.Err)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original")
	err := Wrap(ErrorTypeStorage, "ストレージエラー", originalErr)

	if err.Type != ErrorTypeStorage {
		t.Errorf("Type = %v, 期待 %v", err.Type, ErrorTypeStorage)
	}

	if err.Message != "ストレージエラー" {
		t.Errorf("Message = %v, 期待 %v", err.Message, "ストレージエラー")
	}

	if err.Err != originalErr {
		t.Errorf("Err = %v, 期待 %v", err.Err, originalErr)
	}
}

func TestErrorConstructors(t *testing.T) {
	originalErr := errors.New("test error")

	tests := []struct {
		name         string
		constructor  func() *AppError
		expectedType ErrorType
	}{
		{
			name:         "NewConfigError",
			constructor:  func() *AppError { return NewConfigError("config error", originalErr) },
			expectedType: ErrorTypeConfig,
		},
		{
			name:         "NewNetworkError",
			constructor:  func() *AppError { return NewNetworkError("network error", originalErr) },
			expectedType: ErrorTypeNetwork,
		},
		{
			name:         "NewValidationError",
			constructor:  func() *AppError { return NewValidationError("validation error", originalErr) },
			expectedType: ErrorTypeValidation,
		},
		{
			name:         "NewStorageError",
			constructor:  func() *AppError { return NewStorageError("storage error", originalErr) },
			expectedType: ErrorTypeStorage,
		},
		{
			name:         "NewLLMError",
			constructor:  func() *AppError { return NewLLMError("llm error", originalErr) },
			expectedType: ErrorTypeLLM,
		},
		{
			name:         "NewDiscordError",
			constructor:  func() *AppError { return NewDiscordError("discord error", originalErr) },
			expectedType: ErrorTypeDiscord,
		},
		{
			name:         "NewRSSError",
			constructor:  func() *AppError { return NewRSSError("rss error", originalErr) },
			expectedType: ErrorTypeRSS,
		},
		{
			name:         "NewArticleError",
			constructor:  func() *AppError { return NewArticleError("article error", originalErr) },
			expectedType: ErrorTypeArticle,
		},
		{
			name:         "NewInternalError",
			constructor:  func() *AppError { return NewInternalError("internal error", originalErr) },
			expectedType: ErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()
			if err.Type != tt.expectedType {
				t.Errorf("Type = %v, 期待 %v", err.Type, tt.expectedType)
			}
			if err.Err != originalErr {
				t.Errorf("Err = %v, 期待 %v", err.Err, originalErr)
			}
		})
	}
}
