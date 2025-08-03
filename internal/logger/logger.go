package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	mu sync.Mutex
	L  *slog.Logger
)

func init() {
	handlerOpts := &slog.HandlerOptions{
		Level: parseLevel(os.Getenv("LOG_LEVEL")),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				t := a.Value.Time()
				return slog.String("time", t.Format(time.DateTime))
			case slog.LevelKey:
				return slog.String("level", strings.ToUpper(a.Value.String()))
			default:
				// Preserve Unicode characters in string values without escaping
				if a.Value.Kind() == slog.KindString {
					// Return the string value as-is to prevent Unicode escaping
					return a
				}
				return a
			}
		},
	}
	// Use custom handler that preserves Unicode without escaping
	L = slog.New(NewUnicodeTextHandler(os.Stderr, handlerOpts))
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

// UnicodeTextHandler is a text handler that preserves Unicode characters without escaping
type UnicodeTextHandler struct {
	w    io.Writer
	opts *slog.HandlerOptions
}

// NewUnicodeTextHandler creates a new Unicode-preserving text handler
func NewUnicodeTextHandler(w io.Writer, opts *slog.HandlerOptions) *UnicodeTextHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &UnicodeTextHandler{w: w, opts: opts}
}

// Enabled reports whether the handler handles records at the given level
func (h *UnicodeTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle formats and writes a log record
func (h *UnicodeTextHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)
	
	// Apply ReplaceAttr if set
	replaceAttr := h.opts.ReplaceAttr
	if replaceAttr == nil {
		replaceAttr = func(groups []string, a slog.Attr) slog.Attr { return a }
	}
	
	// Format time
	timeAttr := replaceAttr(nil, slog.Time(slog.TimeKey, r.Time))
	buf = fmt.Appendf(buf, "%s ", timeAttr.Value.String())
	
	// Format level
	levelAttr := replaceAttr(nil, slog.Any(slog.LevelKey, r.Level))
	buf = fmt.Appendf(buf, "[%s] ", levelAttr.Value.String())
	
	// Format message
	buf = fmt.Appendf(buf, "%s", r.Message)
	
	// Format attributes
	r.Attrs(func(a slog.Attr) bool {
		a = replaceAttr(nil, a)
		buf = append(buf, ' ')
		buf = h.appendAttr(buf, a)
		return true
	})
	
	buf = append(buf, '\n')
	_, err := h.w.Write(buf)
	return err
}

// appendAttr appends a single attribute to the buffer, preserving Unicode
func (h *UnicodeTextHandler) appendAttr(buf []byte, a slog.Attr) []byte {
	if a.Equal(slog.Attr{}) {
		return buf
	}
	
	buf = append(buf, a.Key...)
	buf = append(buf, '=')
	
	switch a.Value.Kind() {
	case slog.KindString:
		// Preserve Unicode characters without escaping
		s := a.Value.String()
		if needsQuoting(s) {
			buf = append(buf, '"')
			buf = append(buf, s...)
			buf = append(buf, '"')
		} else {
			buf = append(buf, s...)
		}
	case slog.KindInt64:
		buf = strconv.AppendInt(buf, a.Value.Int64(), 10)
	case slog.KindUint64:
		buf = strconv.AppendUint(buf, a.Value.Uint64(), 10)
	case slog.KindFloat64:
		buf = strconv.AppendFloat(buf, a.Value.Float64(), 'g', -1, 64)
	case slog.KindBool:
		buf = strconv.AppendBool(buf, a.Value.Bool())
	default:
		// For other types, use the string representation
		s := a.Value.String()
		if needsQuoting(s) {
			buf = append(buf, '"')
			buf = append(buf, s...)
			buf = append(buf, '"')
		} else {
			buf = append(buf, s...)
		}
	}
	
	return buf
}

// needsQuoting returns true if the string needs to be quoted
func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			return true
		}
		if r < 32 || r == '"' || r == '\\' {
			return true
		}
		i += size
	}
	return false
}

// WithAttrs returns a new handler with the given attributes
func (h *UnicodeTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, return the same handler
	// In a full implementation, you'd store these attrs and include them in Handle
	return h
}

// WithGroup returns a new handler with the given group name
func (h *UnicodeTextHandler) WithGroup(name string) slog.Handler {
	// For simplicity, return the same handler
	// In a full implementation, you'd track the group name
	return h
}
