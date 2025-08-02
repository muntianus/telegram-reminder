package handlers

import (
	"context"

	"telegram-reminder/internal/domain"
	"telegram-reminder/internal/logger"
	"telegram-reminder/internal/services"

	tb "gopkg.in/telebot.v3"
)

// DigestHandler handles digest-related commands
type DigestHandler struct {
	digestService *services.DigestService
	errorHandler  ErrorHandler
}

// NewDigestHandler creates a new digest handler
func NewDigestHandler(digestService *services.DigestService, errorHandler ErrorHandler) *DigestHandler {
	return &DigestHandler{
		digestService: digestService,
		errorHandler:  errorHandler,
	}
}

// ErrorHandler defines interface for error handling
type ErrorHandler interface {
	HandleOpenAIError(err error, model string) string
}

// TelegramSender defines interface for sending Telegram messages
type TelegramSender interface {
	Send(what interface{}, opts ...interface{}) error
}

// HandleDigest creates a generic digest handler for any digest type
func (h *DigestHandler) HandleDigest(digestType domain.DigestType) func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("digest command", "type", digestType, "chat", c.Chat().ID)

		// Get current model from runtime config
		model := getCurrentModel()

		// Generate digest
		req := services.DigestRequest{
			Type:   digestType,
			Model:  model,
			ChatID: c.Chat().ID,
		}

		resp, err := h.digestService.GenerateDigest(context.Background(), req)
		if err != nil {
			logger.L.Error("digest generation error", "type", digestType, "model", model, "error", err)
			return c.Send(h.errorHandler.HandleOpenAIError(err, model))
		}

		if resp.Content == "" {
			logger.L.Warn("empty digest content", "type", digestType)
			return c.Send("❌ Получен пустой дайджест")
		}

		return h.replyLong(c, resp.Content)
	}
}

// RegisterDigestHandlers registers all digest handlers with the bot
func (h *DigestHandler) RegisterDigestHandlers(bot TelegramBot) {
	configs := domain.GetDigestConfigs()

	for digestType, config := range configs {
		handler := h.HandleDigest(digestType)
		bot.Handle("/"+config.CommandName, handler)
		logger.L.Debug("registered digest handler", "command", config.CommandName, "type", digestType)
	}
}

// TelegramBot defines the interface for bot registration
type TelegramBot interface {
	Handle(endpoint interface{}, handler interface{}, middlewares ...tb.MiddlewareFunc)
}

// replyLong sends long messages, splitting if necessary
func (h *DigestHandler) replyLong(c tb.Context, text string) error {
	if text == "" {
		logger.L.Warn("empty text in replyLong")
		return c.Send("❌ Получен пустой ответ")
	}

	const telegramMessageLimit = 4096
	runes := []rune(text)

	for len(runes) > 0 {
		end := telegramMessageLimit
		if len(runes) < end {
			end = len(runes)
		}

		if err := c.Send(string(runes[:end]), tb.ModeHTML); err != nil {
			return err
		}

		runes = runes[end:]
	}

	return nil
}

// getCurrentModel returns the current AI model from runtime config
// This is a placeholder - actual implementation would get this from config service
func getCurrentModel() string {
	// TODO: Implement proper config service injection
	return "gpt-4.1"
}
