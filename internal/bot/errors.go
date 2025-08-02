package bot

import (
	"fmt"
	"strings"

	"telegram-reminder/internal/logger"
)

// ErrorHandler provides utilities for handling and formatting errors
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler instance
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleOpenAIError formats OpenAI API errors for user consumption
func (eh *ErrorHandler) HandleOpenAIError(err error, model string) string {
	if err == nil {
		return ""
	}

	logger.L.Error("openai error", "model", model, "error", err)
	return formatOpenAIError(err, model)
}

// HandleTaskError logs and formats task execution errors
func (eh *ErrorHandler) HandleTaskError(err error, taskName, model string) string {
	if err == nil {
		return ""
	}

	logger.L.Error("task execution error", "task", taskName, "model", model, "error", err)
	return fmt.Sprintf("❌ Ошибка выполнения задачи '%s': %v", taskName, err)
}

// HandleTelegramError logs telegram API errors
func (eh *ErrorHandler) HandleTelegramError(err error, chatID int64) {
	if err == nil {
		return
	}

	logger.L.Error("telegram error", "chat_id", chatID, "error", err)
}

// HandleWhitelistError logs and formats whitelist operation errors
func (eh *ErrorHandler) HandleWhitelistError(err error, operation string) string {
	if err == nil {
		return ""
	}

	logger.L.Error("whitelist error", "operation", operation, "error", err)
	return fmt.Sprintf("❌ Ошибка %s: %v", operation, err)
}

// IsRetryableError determines if an error is retryable
func (eh *ErrorHandler) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "rate_limit") ||
		strings.Contains(errStr, "context deadline exceeded")
}

// GetRetryDelay returns appropriate delay for retryable errors
func (eh *ErrorHandler) GetRetryDelay(err error) int {
	if err == nil {
		return 0
	}

	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "rate_limit"):
		return 60 // 1 minute for rate limits
	case strings.Contains(errStr, "timeout"):
		return 30 // 30 seconds for timeouts
	default:
		return 10 // 10 seconds for other retryable errors
	}
}

// Global error handler instance
var DefaultErrorHandler = NewErrorHandler()
