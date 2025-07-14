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
	Init("info")
}

// Init configures the global logger with the given level.
func Init(level string) {
	mu.Lock()
	defer mu.Unlock()
	L = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: parseLevel(level)}))
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

func parseLevel(l string) slog.Level {
	switch strings.ToLower(l) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
