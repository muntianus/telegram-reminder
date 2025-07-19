package config

import (
	"fmt"
	"os"
	"strconv"
)

// Environment variable names
const (
	EnvTelegramToken = "TELEGRAM_TOKEN"
	EnvChatID        = "CHAT_ID"
	EnvOpenAIKey     = "OPENAI_API_KEY"
	EnvOpenAIModel   = "OPENAI_MODEL"
	EnvBlockchainAPI = "BLOCKCHAIN_API"
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

	telegramToken := os.Getenv(EnvTelegramToken)
	chatIDStr := os.Getenv(EnvChatID)
	openaiKey := os.Getenv(EnvOpenAIKey)
	openaiModel := os.Getenv(EnvOpenAIModel)
	blockchainAPI := os.Getenv(EnvBlockchainAPI)

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
