// Package main demonstrates the new structured logging system
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"telegram-reminder/internal/logger"
)

func main() {
	// Set up demo environment
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "pretty")
	os.Setenv("LOG_COLORS", "true")
	os.Setenv("LOG_MODULE_LEVELS", "demo=debug,api=info")

	fmt.Println("=== Демонстрация структурированного логирования ===")

	// 1. Basic module logging
	fmt.Println("1. Базовое модульное логирование:")
	demoBasicLogging()

	// 2. Operation logging
	fmt.Println("\n2. Операционное логирование:")
	demoOperationLogging()

	// 3. Specialized logging methods
	fmt.Println("\n3. Специализированные методы:")
	demoSpecializedLogging()

	// 4. Performance logging
	fmt.Println("\n4. Логирование производительности:")
	demoPerformanceLogging()

	// 5. Legacy compatibility
	fmt.Println("\n5. Совместимость со старой системой:")
	demoLegacyCompatibility()

	fmt.Println("\n=== Демонстрация завершена ===")
}

func demoBasicLogging() {
	// Get module-specific loggers
	botLogger := logger.GetBotLogger()
	apiLogger := logger.GetAPILogger()

	botLogger.Debug("Bot initialization started", "version", "1.0.0")
	botLogger.Info("Bot service started", "port", 8080, "env", "development")

	apiLogger.Info("API endpoint registered", "path", "/api/chat", "method", "POST")
	apiLogger.Warn("Rate limit approaching", "current_requests", 950, "limit", 1000)
}

func demoOperationLogging() {
	handlerLogger := logger.GetHandlerLogger()

	// Create an operation tracker
	op := handlerLogger.Operation("user_chat_request")
	op.WithContext("user_id", 123456789)
	op.WithContext("session_id", "sess_abc123")

	// Log operation steps
	op.Step("validating_input", "message_length", 25)
	time.Sleep(10 * time.Millisecond) // Simulate work

	op.Step("processing_request", "model", "gpt-4")
	time.Sleep(50 * time.Millisecond) // Simulate API call

	// Simulate successful completion
	op.Success("Chat request processed successfully",
		"response_length", 150,
		"tokens_used", 75)
}

func demoSpecializedLogging() {
	openaiLogger := logger.GetOpenAILogger()
	handlerLogger := logger.GetHandlerLogger()
	securityLogger := logger.GetSecurityLogger()
	taskLogger := logger.GetTaskLogger()

	// API call logging
	duration := 1250 * time.Millisecond
	openaiLogger.APICall("openai", "chat_completion", true, duration, nil)

	// User action logging
	handlerLogger.UserAction(123456789, "send_message", map[string]interface{}{
		"message_type":      "text",
		"length":            45,
		"contains_mentions": false,
	})

	// Security event logging
	securityLogger.SecurityEvent("unusual_activity", 123456789, map[string]interface{}{
		"event_type":    "rapid_requests",
		"request_count": 15,
		"time_window":   "1m",
	})

	// Task execution logging
	taskLogger.TaskExecution("daily_digest", true, 2*time.Second, nil)

	// HTTP request logging
	handlerLogger.HTTPRequest("POST", "/api/chat", 200, 150*time.Millisecond)
}

func demoPerformanceLogging() {
	botLogger := logger.GetBotLogger()

	// Performance metrics
	botLogger.Performance("database_query", 45*time.Millisecond, map[string]interface{}{
		"query_type":    "SELECT",
		"table":         "users",
		"rows_returned": 1,
		"cache_hit":     false,
	})

	botLogger.Performance("redis_operation", 5*time.Millisecond, map[string]interface{}{
		"operation":   "GET",
		"key_pattern": "user:*",
		"cache_hit":   true,
	})
}

func demoLegacyCompatibility() {
	// Old style logging (still works)
	logger.L.Debug("Legacy debug message", "key", "value")
	logger.L.Info("Legacy info message", "user_id", 123)

	// New convenience functions
	logger.Info("New style info", "module", "demo")
	logger.Debug("New style debug", "timestamp", time.Now())
}

// Demonstration of error handling with logging
func demoErrorScenario() {
	handlerLogger := logger.GetHandlerLogger()

	op := handlerLogger.Operation("failed_operation")
	op.WithContext("user_id", 999999)

	op.Step("attempting_dangerous_operation")

	// Simulate an error
	err := errors.New("simulated network timeout")
	op.Failure("Operation failed due to network error", err)
}
