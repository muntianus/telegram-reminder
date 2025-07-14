package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds environment configuration values.
type Config struct {
	TelegramToken string
	ChatID        int64
	OpenAIKey     string
	OpenAIModel   string
}

// Load reads environment variables and validates them.
func Load() (Config, error) {
	var cfg Config

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	openaiModel := os.Getenv("OPENAI_MODEL")

	if telegramToken == "" || chatIDStr == "" || openaiKey == "" {
		return cfg, fmt.Errorf("missing required env vars")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return cfg, fmt.Errorf("invalid CHAT_ID: %w", err)
	}

	cfg = Config{
		TelegramToken: telegramToken,
		ChatID:        chatID,
		OpenAIKey:     openaiKey,
		OpenAIModel:   openaiModel,
	}

	return cfg, nil
}
