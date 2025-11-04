// Package errors はアプリケーション固有のカスタムエラー型を提供します
package errors

import "fmt"

// ErrorType はエラーの種類を表す列挙型
type ErrorType string

const (
	// ErrorTypeConfig は設定関連のエラー
	ErrorTypeConfig ErrorType = "CONFIG"
	// ErrorTypeNetwork はネットワーク関連のエラー
	ErrorTypeNetwork ErrorType = "NETWORK"
	// ErrorTypeValidation はバリデーション関連のエラー
	ErrorTypeValidation ErrorType = "VALIDATION"
	// ErrorTypeStorage はストレージ（Firestore）関連のエラー
	ErrorTypeStorage ErrorType = "STORAGE"
	// ErrorTypeLLM はLLM（Gemini）関連のエラー
	ErrorTypeLLM ErrorType = "LLM"
	// ErrorTypeDiscord はDiscord API関連のエラー
	ErrorTypeDiscord ErrorType = "DISCORD"
	// ErrorTypeRSS はRSSフィード関連のエラー
	ErrorTypeRSS ErrorType = "RSS"
	// ErrorTypeArticle は記事処理関連のエラー
	ErrorTypeArticle ErrorType = "ARTICLE"
	// ErrorTypeInternal は内部エラー
	ErrorTypeInternal ErrorType = "INTERNAL"
)

// AppError はアプリケーション固有のエラー型
type AppError struct {
	Type    ErrorType // エラーの種類
	Message string    // エラーメッセージ
	Err     error     // 元のエラー（ラップされている場合）
}

// Error はerrorインターフェースを実装します
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap は元のエラーを返します（errors.Unwrap対応）
func (e *AppError) Unwrap() error {
	return e.Err
}

// New は新しいAppErrorを作成します
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     nil,
	}
}

// Wrap は既存のエラーをラップして新しいAppErrorを作成します
func Wrap(errType ErrorType, message string, err error) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// NewConfigError は設定関連のエラーを作成します
func NewConfigError(message string, err error) *AppError {
	return Wrap(ErrorTypeConfig, message, err)
}

// NewNetworkError はネットワーク関連のエラーを作成します
func NewNetworkError(message string, err error) *AppError {
	return Wrap(ErrorTypeNetwork, message, err)
}

// NewValidationError はバリデーション関連のエラーを作成します
func NewValidationError(message string, err error) *AppError {
	return Wrap(ErrorTypeValidation, message, err)
}

// NewStorageError はストレージ関連のエラーを作成します
func NewStorageError(message string, err error) *AppError {
	return Wrap(ErrorTypeStorage, message, err)
}

// NewLLMError はLLM関連のエラーを作成します
func NewLLMError(message string, err error) *AppError {
	return Wrap(ErrorTypeLLM, message, err)
}

// NewDiscordError はDiscord API関連のエラーを作成します
func NewDiscordError(message string, err error) *AppError {
	return Wrap(ErrorTypeDiscord, message, err)
}

// NewRSSError はRSSフィード関連のエラーを作成します
func NewRSSError(message string, err error) *AppError {
	return Wrap(ErrorTypeRSS, message, err)
}

// NewArticleError は記事処理関連のエラーを作成します
func NewArticleError(message string, err error) *AppError {
	return Wrap(ErrorTypeArticle, message, err)
}

// NewInternalError は内部エラーを作成します
func NewInternalError(message string, err error) *AppError {
	return Wrap(ErrorTypeInternal, message, err)
}
