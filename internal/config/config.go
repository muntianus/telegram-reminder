package config

import (
	"fmt"
	"os"
	"strconv"
)

const DefaultBlockchainAPI = "https://api.blockchain.info/stats"

// Config holds environment configuration values.
type Config struct {
	TelegramToken string
	ChatID        int64
	OpenAIKey     string
	OpenAIModel   string
	BlockchainAPI string
}

// Load reads environment variables and validates them.
func Load() (Config, error) {
	var cfg Config

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	openaiModel := os.Getenv("OPENAI_MODEL")
	blockchainAPI := os.Getenv("BLOCKCHAIN_API")

	if telegramToken == "" || openaiKey == "" {
		return cfg, fmt.Errorf("missing required env vars")
	}

	var chatID int64
	if chatIDStr != "" {
		var err error
		chatID, err = strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("invalid CHAT_ID: %w", err)
		}
	}

	if blockchainAPI == "" {
		blockchainAPI = DefaultBlockchainAPI
	}

	cfg = Config{
		TelegramToken: telegramToken,
		ChatID:        chatID,
		OpenAIKey:     openaiKey,
		OpenAIModel:   openaiModel,
		BlockchainAPI: blockchainAPI,
	}

	return cfg, nil
}
