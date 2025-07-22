package config

import (
	"fmt"
	"os"
	"strconv"
)

// Environment variable names
const (
	EnvTelegramToken     = "TELEGRAM_TOKEN"
	EnvChatID            = "CHAT_ID"
	EnvLogChatID         = "LOG_CHAT_ID"
	EnvOpenAIKey         = "OPENAI_API_KEY"
	EnvOpenAIModel       = "OPENAI_MODEL"
	EnvBlockchainAPI     = "BLOCKCHAIN_API"
	EnvEnableWebSearch   = "ENABLE_WEB_SEARCH"
	EnvSearchProviderURL = "SEARCH_PROVIDER_URL"
)

const DefaultBlockchainAPI = "https://api.blockchain.info/stats"
const DefaultSearchProviderURL = "https://api.duckduckgo.com"

// Config holds environment configuration values.
type Config struct {
	TelegramToken     string
	ChatID            int64
	LogChatID         int64
	OpenAIKey         string
	OpenAIModel       string
	BlockchainAPI     string
	EnableWebSearch   bool
	SearchProviderURL string
}

// Load reads environment variables and validates them.
func Load() (Config, error) {
	var cfg Config

	telegramToken := os.Getenv(EnvTelegramToken)
	chatIDStr := os.Getenv(EnvChatID)
	logChatIDStr := os.Getenv(EnvLogChatID)
	openaiKey := os.Getenv(EnvOpenAIKey)
	openaiModel := os.Getenv(EnvOpenAIModel)
	blockchainAPI := os.Getenv(EnvBlockchainAPI)
	enableWebSearchStr := os.Getenv(EnvEnableWebSearch)
	searchProviderURL := os.Getenv(EnvSearchProviderURL)

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

	var logChatID int64
	if logChatIDStr != "" {
		var err error
		logChatID, err = strconv.ParseInt(logChatIDStr, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("invalid LOG_CHAT_ID: %w", err)
		}
	}

	enableWebSearch := false
	if enableWebSearchStr != "" {
		var err error
		enableWebSearch, err = strconv.ParseBool(enableWebSearchStr)
		if err != nil {
			return cfg, fmt.Errorf("invalid ENABLE_WEB_SEARCH: %w", err)
		}
	}

	if blockchainAPI == "" {
		blockchainAPI = DefaultBlockchainAPI
	}
	if searchProviderURL == "" {
		searchProviderURL = DefaultSearchProviderURL
	}

	cfg = Config{
		TelegramToken:     telegramToken,
		ChatID:            chatID,
		LogChatID:         logChatID,
		OpenAIKey:         openaiKey,
		OpenAIModel:       openaiModel,
		BlockchainAPI:     blockchainAPI,
		EnableWebSearch:   enableWebSearch,
		SearchProviderURL: searchProviderURL,
	}

	return cfg, nil
}
