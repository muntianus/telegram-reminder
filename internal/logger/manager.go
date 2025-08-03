package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	loggerManager *LoggerManager
	once          sync.Once
)

// LoggerManager manages loggers for different modules
type LoggerManager struct {
	loggers map[string]*StructuredLogger
	config  GlobalLogConfig
	mu      sync.RWMutex
}

// GlobalLogConfig defines global logging configuration
type GlobalLogConfig struct {
	DefaultLevel   slog.Level
	Format         LogFormat
	EnableColors   bool
	ModuleLevels   map[string]slog.Level
	TelegramToken  string
	TelegramChat   int64
	EnableTelegram bool
}

// GetLoggerManager returns the singleton logger manager
func GetLoggerManager() *LoggerManager {
	once.Do(func() {
		loggerManager = NewLoggerManager(LoadLogConfig())
	})
	return loggerManager
}

// LoadLogConfig loads logging configuration from environment
func LoadLogConfig() GlobalLogConfig {
	config := GlobalLogConfig{
		DefaultLevel:   parseLevel(os.Getenv("LOG_LEVEL")),
		Format:         parseFormat(os.Getenv("LOG_FORMAT")),
		EnableColors:   parseBool(os.Getenv("LOG_COLORS"), true),
		ModuleLevels:   parseModuleLevels(os.Getenv("LOG_MODULE_LEVELS")),
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		TelegramChat:   parseInt64(os.Getenv("LOG_TELEGRAM_CHAT")),
		EnableTelegram: parseBool(os.Getenv("LOG_TELEGRAM_ENABLE"), false),
	}

	return config
}

// NewLoggerManager creates a new logger manager
func NewLoggerManager(config GlobalLogConfig) *LoggerManager {
	return &LoggerManager{
		loggers: make(map[string]*StructuredLogger),
		config:  config,
	}
}

// GetLogger returns a structured logger for the specified module
func (lm *LoggerManager) GetLogger(module string) *StructuredLogger {
	lm.mu.RLock()
	if logger, exists := lm.loggers[module]; exists {
		lm.mu.RUnlock()
		return logger
	}
	lm.mu.RUnlock()

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Double-check pattern
	if logger, exists := lm.loggers[module]; exists {
		return logger
	}

	// Determine level for this module
	level := lm.config.DefaultLevel
	if moduleLevel, exists := lm.config.ModuleLevels[module]; exists {
		level = moduleLevel
	}

	loggerConfig := LoggerConfig{
		Level:        level,
		Format:       lm.config.Format,
		EnableColors: lm.config.EnableColors,
		Module:       module,
	}

	logger := NewStructuredLogger(module, loggerConfig)
	lm.loggers[module] = logger

	return logger
}

// SetModuleLevel dynamically sets the log level for a specific module
func (lm *LoggerManager) SetModuleLevel(module string, level slog.Level) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.config.ModuleLevels == nil {
		lm.config.ModuleLevels = make(map[string]slog.Level)
	}
	lm.config.ModuleLevels[module] = level

	// Recreate logger with new level
	delete(lm.loggers, module)
}

// EnableTelegramLogging enables telegram logging for critical events
func (lm *LoggerManager) EnableTelegramLogging() {
	if lm.config.TelegramToken != "" && lm.config.TelegramChat != 0 {
		EnableTelegramLogging(lm.config.TelegramToken, lm.config.TelegramChat, slog.LevelWarn)
	}
}

// GetModuleLoggers returns all active module loggers
func (lm *LoggerManager) GetModuleLoggers() map[string]*StructuredLogger {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	result := make(map[string]*StructuredLogger)
	for k, v := range lm.loggers {
		result[k] = v
	}
	return result
}

// Helper functions for parsing configuration

func parseFormat(format string) LogFormat {
	switch strings.ToLower(format) {
	case "json":
		return FormatJSON
	case "pretty":
		return FormatPretty
	default:
		return FormatText
	}
}

func parseBool(value string, defaultValue bool) bool {
	switch strings.ToLower(value) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultValue
	}
}

func parseInt64(value string) int64 {
	if value == "" {
		return 0
	}
	// Simple conversion - in production, use strconv.ParseInt with error handling
	var result int64
	for _, r := range value {
		if r >= '0' && r <= '9' {
			result = result*10 + int64(r-'0')
		}
	}
	return result
}

func parseModuleLevels(config string) map[string]slog.Level {
	result := make(map[string]slog.Level)
	if config == "" {
		return result
	}

	// Format: "module1=debug,module2=info,module3=warn"
	pairs := strings.Split(config, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), "=")
		if len(parts) == 2 {
			module := strings.TrimSpace(parts[0])
			level := parseLevel(strings.TrimSpace(parts[1]))
			result[module] = level
		}
	}

	return result
}

// Module-specific logger getters

// GetBotLogger returns logger for bot operations
func GetBotLogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("bot")
}

// GetAPILogger returns logger for API calls
func GetAPILogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("api")
}

// GetTaskLogger returns logger for task execution
func GetTaskLogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("task")
}

// GetTelegramLogger returns logger for telegram operations
func GetTelegramLogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("telegram")
}

// GetSecurityLogger returns logger for security events
func GetSecurityLogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("security")
}

// GetDigestLogger returns logger for digest operations
func GetDigestLogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("digest")
}

// GetHandlerLogger returns logger for request handlers
func GetHandlerLogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("handler")
}

// GetOpenAILogger returns logger for OpenAI operations
func GetOpenAILogger() *StructuredLogger {
	return GetLoggerManager().GetLogger("openai")
}
