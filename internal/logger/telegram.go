package logger

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	var attrs []any
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a.Key, a.Value.Any())
		return true
	})
	text := fmt.Sprintf("%s [%s] %s", r.Time.Format(time.RFC3339), r.Level.String(), r.Message)
	if len(attrs) > 0 {
		text += " " + fmt.Sprint(attrs...)
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
