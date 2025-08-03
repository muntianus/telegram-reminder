package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu sync.Mutex
	L  *slog.Logger
)

func init() {
	// Initialize with simple text handler first
	L = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: parseLevel(os.Getenv("LOG_LEVEL"))}))
}

// SetLogger replaces the global logger. Useful in tests.
func SetLogger(l *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	L = l
}

// GetLogger returns the current logger.
func GetLogger() *slog.Logger {
	mu.Lock()
	defer mu.Unlock()
	return L
}

// parseLevel parses string level to slog.Level
func parseLevel(l string) slog.Level {
	switch strings.ToLower(l) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "":
		return slog.LevelError
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

// EnableTelegramLogging adds a Telegram handler to the global logger.
func EnableTelegramLogging(token string, chatID int64, level slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	th := NewTelegramHandler(token, chatID, level)
	L = slog.New(newMulti(L.Handler(), th))
}

// Legacy compatibility functions

// Info logs an info message (compatibility)
func Info(msg string, args ...interface{}) {
	L.Info(msg, args...)
}

// Debug logs a debug message (compatibility)
func Debug(msg string, args ...interface{}) {
	L.Debug(msg, args...)
}

// Warn logs a warning message (compatibility)
func Warn(msg string, args ...interface{}) {
	L.Warn(msg, args...)
}

// Error logs an error message (compatibility)
func Error(msg string, args ...interface{}) {
	L.Error(msg, args...)
}
