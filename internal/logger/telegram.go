package logger

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"log/slog"
)

// TelegramHandler sends logs to a Telegram chat.
type TelegramHandler struct {
	token  string
	chatID int64
	level  slog.Level
}

// NewTelegramHandler creates a handler that posts log messages to Telegram.
func NewTelegramHandler(token string, chatID int64, level slog.Level) *TelegramHandler {
	return &TelegramHandler{token: token, chatID: chatID, level: level}
}

func (h *TelegramHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *TelegramHandler) Handle(ctx context.Context, r slog.Record) error {
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		// Safely handle different value types
		var value string
		switch a.Value.Kind() {
		case slog.KindTime:
			value = a.Value.Time().Format(time.RFC3339)
		case slog.KindString:
			value = a.Value.String()
		default:
			value = fmt.Sprintf("%v", a.Value.Any())
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", a.Key, value))
		return true
	})
	text := fmt.Sprintf("%s [%s] %s", r.Time.Format(time.RFC3339), r.Level.String(), r.Message)
	if len(attrs) > 0 {
		text += " " + fmt.Sprintf("{%s}", strings.Join(attrs, ", "))
	}
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", h.token)
	values := url.Values{
		"chat_id": {strconv.FormatInt(h.chatID, 10)},
		"text":    {text},
	}
	_, err := http.PostForm(api, values)
	return err
}

func (h *TelegramHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *TelegramHandler) WithGroup(name string) slog.Handler       { return h }
