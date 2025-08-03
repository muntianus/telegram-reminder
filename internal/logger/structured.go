package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// LoggerConfig defines configuration for structured logging
type LoggerConfig struct {
	Level        slog.Level
	Format       LogFormat
	EnableColors bool
	Module       string
}

// LogFormat defines the output format for logs
type LogFormat string

const (
	FormatJSON   LogFormat = "json"
	FormatText   LogFormat = "text"
	FormatPretty LogFormat = "pretty"
)

// StructuredLogger provides context-aware logging with better formatting
type StructuredLogger struct {
	*slog.Logger
	module string
}

// NewStructuredLogger creates a new structured logger for a specific module
func NewStructuredLogger(module string, config LoggerConfig) *StructuredLogger {
	var handler slog.Handler

	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: config.Level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
				}
				return a
			},
		})
	case FormatPretty:
		handler = NewPrettyHandler(config.Level, config.EnableColors)
	default:
		handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: config.Level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.String("time", a.Value.Time().Format("15:04:05"))
				}
				return a
			},
		})
	}

	baseLogger := slog.New(handler)
	
	// Add module context to all logs
	if module != "" {
		baseLogger = baseLogger.With("module", module)
	}

	return &StructuredLogger{
		Logger: baseLogger,
		module: module,
	}
}

// Operation logging methods
func (l *StructuredLogger) Operation(operation string) *OperationLogger {
	return &OperationLogger{
		Logger:    l.Logger,
		operation: operation,
		startTime: time.Now(),
	}
}

// HTTP request logging
func (l *StructuredLogger) HTTPRequest(method, url string, statusCode int, duration time.Duration) {
	level := slog.LevelInfo
	if statusCode >= 400 {
		level = slog.LevelWarn
	}
	if statusCode >= 500 {
		level = slog.LevelError
	}

	l.Log(context.Background(), level, "HTTP request",
		"method", method,
		"url", url,
		"status_code", statusCode,
		"duration", duration,
	)
}

// API call logging
func (l *StructuredLogger) APICall(service, endpoint string, success bool, duration time.Duration, err error) {
	if success {
		l.Info("API call successful",
			"service", service,
			"endpoint", endpoint,
			"duration", duration,
		)
	} else {
		l.Error("API call failed",
			"service", service,
			"endpoint", endpoint,
			"duration", duration,
			"error", err,
		)
	}
}

// Task execution logging
func (l *StructuredLogger) TaskExecution(taskName string, success bool, duration time.Duration, err error) {
	if success {
		l.Info("Task completed successfully",
			"task", taskName,
			"duration", duration,
		)
	} else {
		l.Error("Task execution failed",
			"task", taskName,
			"duration", duration,
			"error", err,
		)
	}
}

// User action logging
func (l *StructuredLogger) UserAction(userID int64, action string, payload interface{}) {
	l.Info("User action",
		"user_id", userID,
		"action", action,
		"payload", payload,
	)
}

// Security event logging
func (l *StructuredLogger) SecurityEvent(event string, userID int64, details map[string]interface{}) {
	attrs := []interface{}{"event", event, "user_id", userID}
	for k, v := range details {
		attrs = append(attrs, k, v)
	}
	l.Warn("Security event", attrs...)
}

// Performance logging
func (l *StructuredLogger) Performance(operation string, duration time.Duration, metrics map[string]interface{}) {
	attrs := []interface{}{"operation", operation, "duration", duration}
	for k, v := range metrics {
		attrs = append(attrs, k, v)
	}
	l.Debug("Performance metrics", attrs...)
}

// OperationLogger tracks the lifecycle of an operation
type OperationLogger struct {
	*slog.Logger
	operation string
	startTime time.Time
	context   map[string]interface{}
}

// WithContext adds context to the operation
func (ol *OperationLogger) WithContext(key string, value interface{}) *OperationLogger {
	if ol.context == nil {
		ol.context = make(map[string]interface{})
	}
	ol.context[key] = value
	return ol
}

// Success logs successful completion of the operation
func (ol *OperationLogger) Success(message string, extraAttrs ...interface{}) {
	duration := time.Since(ol.startTime)
	attrs := []interface{}{
		"operation", ol.operation,
		"duration", duration,
		"status", "success",
	}
	
	for k, v := range ol.context {
		attrs = append(attrs, k, v)
	}
	
	attrs = append(attrs, extraAttrs...)
	ol.Info(message, attrs...)
}

// Failure logs failed completion of the operation
func (ol *OperationLogger) Failure(message string, err error, extraAttrs ...interface{}) {
	duration := time.Since(ol.startTime)
	attrs := []interface{}{
		"operation", ol.operation,
		"duration", duration,
		"status", "failure",
		"error", err,
	}
	
	for k, v := range ol.context {
		attrs = append(attrs, k, v)
	}
	
	attrs = append(attrs, extraAttrs...)
	ol.Error(message, attrs...)
}

// Step logs a step within the operation
func (ol *OperationLogger) Step(step string, attrs ...interface{}) {
	stepAttrs := []interface{}{
		"operation", ol.operation,
		"step", step,
	}
	
	for k, v := range ol.context {
		stepAttrs = append(stepAttrs, k, v)
	}
	
	stepAttrs = append(stepAttrs, attrs...)
	ol.Debug("Operation step", stepAttrs...)
}

// PrettyHandler provides human-readable log formatting
type PrettyHandler struct {
	level       slog.Level
	enableColors bool
}

// NewPrettyHandler creates a new pretty-formatted handler
func NewPrettyHandler(level slog.Level, enableColors bool) *PrettyHandler {
	return &PrettyHandler{
		level:       level,
		enableColors: enableColors,
	}
}

func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	var b strings.Builder
	
	// Timestamp
	b.WriteString(r.Time.Format("15:04:05"))
	b.WriteString(" ")
	
	// Level with colors
	levelStr := h.formatLevel(r.Level)
	b.WriteString(levelStr)
	b.WriteString(" ")
	
	// Module (if present)
	var module string
	var attrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "module" {
			module = a.Value.String()
			return true
		}
		attrs = append(attrs, a)
		return true
	})
	
	if module != "" {
		if h.enableColors {
			b.WriteString(fmt.Sprintf("\033[36m[%s]\033[0m ", module))
		} else {
			b.WriteString(fmt.Sprintf("[%s] ", module))
		}
	}
	
	// Message
	b.WriteString(r.Message)
	
	// Attributes
	if len(attrs) > 0 {
		b.WriteString(" {")
		for i, attr := range attrs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(h.formatAttribute(attr))
		}
		b.WriteString("}")
	}
	
	b.WriteString("\n")
	
	_, err := os.Stderr.Write([]byte(b.String()))
	return err
}

func (h *PrettyHandler) formatLevel(level slog.Level) string {
	if !h.enableColors {
		return fmt.Sprintf("[%s]", level.String())
	}
	
	switch level {
	case slog.LevelDebug:
		return "\033[37m[DEBUG]\033[0m" // White
	case slog.LevelInfo:
		return "\033[32m[INFO ]\033[0m" // Green
	case slog.LevelWarn:
		return "\033[33m[WARN ]\033[0m" // Yellow
	case slog.LevelError:
		return "\033[31m[ERROR]\033[0m" // Red
	default:
		return fmt.Sprintf("[%s]", level.String())
	}
}

func (h *PrettyHandler) formatAttribute(attr slog.Attr) string {
	key := attr.Key
	value := h.formatValue(attr.Value)
	
	if h.enableColors {
		return fmt.Sprintf("\033[35m%s\033[0m=\033[37m%s\033[0m", key, value)
	}
	return fmt.Sprintf("%s=%s", key, value)
}

func (h *PrettyHandler) formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindTime:
		return v.Time().Format(time.RFC3339)
	case slog.KindDuration:
		return v.Duration().String()
	case slog.KindString:
		str := v.String()
		if len(str) > 100 {
			return str[:97] + "..."
		}
		return str
	default:
		return fmt.Sprintf("%v", v.Any())
	}
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, we don't implement persistent attributes in pretty handler
	return h
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	// For simplicity, we don't implement groups in pretty handler
	return h
}