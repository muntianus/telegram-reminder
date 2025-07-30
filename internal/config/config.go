package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Environment variable names
const (
	EnvTelegramToken         = "TELEGRAM_TOKEN"
	EnvChatID                = "CHAT_ID"
	EnvLogChatID             = "LOG_CHAT_ID"
	EnvOpenAIKey             = "OPENAI_API_KEY"
	EnvOpenAIModel           = "OPENAI_MODEL"
	EnvBlockchainAPI         = "BLOCKCHAIN_API"
	EnvEnableWebSearch       = "ENABLE_WEB_SEARCH"
	EnvOpenAIMaxTokens       = "OPENAI_MAX_TOKENS"
	EnvOpenAIToolChoice      = "OPENAI_TOOL_CHOICE"
	EnvOpenAIServiceTier     = "OPENAI_SERVICE_TIER"
	EnvOpenAIReasoningEffort = "OPENAI_REASONING_EFFORT"
)

const DefaultBlockchainAPI = "https://api.blockchain.info/stats"

// Config holds environment configuration values.
type Config struct {
	TelegramToken         string
	ChatID                int64
	LogChatID             int64
	OpenAIKey             string
	OpenAIModel           string
	OpenAIMaxTokens       int
	BlockchainAPI         string
	EnableWebSearch       bool
	OpenAIToolChoice      string
	OpenAIServiceTier     string
	OpenAIReasoningEffort string
}

// Load reads environment variables and validates them.
func Load() (Config, error) {
	var cfg Config

	telegramToken := os.Getenv(EnvTelegramToken)
	chatIDStr := os.Getenv(EnvChatID)
	logChatIDStr := os.Getenv(EnvLogChatID)
	openaiKey := os.Getenv(EnvOpenAIKey)
	openaiModel := os.Getenv(EnvOpenAIModel)
	if openaiModel == "" {
		openaiModel = "gpt-4.1"
	}
	blockchainAPI := os.Getenv(EnvBlockchainAPI)
	enableWebSearchStr := os.Getenv(EnvEnableWebSearch)
	maxTokensStr := os.Getenv(EnvOpenAIMaxTokens)
	toolChoice := os.Getenv(EnvOpenAIToolChoice)
	serviceTier := os.Getenv(EnvOpenAIServiceTier)
	reasoningEffort := os.Getenv(EnvOpenAIReasoningEffort)

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

	if blockchainAPI == "" {
		blockchainAPI = DefaultBlockchainAPI
	}

	enableWebSearch := true
	if enableWebSearchStr != "" {
		enableWebSearch = enableWebSearchStr == "1" || strings.ToLower(enableWebSearchStr) == "true"
	}

	if toolChoice == "" {
		toolChoice = "auto"
	}

	maxTokens := 600
	if maxTokensStr != "" {
		if v, err := strconv.Atoi(maxTokensStr); err == nil {
			maxTokens = v
		}
	}

	cfg = Config{
		TelegramToken:         telegramToken,
		ChatID:                chatID,
		LogChatID:             logChatID,
		OpenAIKey:             openaiKey,
		OpenAIModel:           openaiModel,
		OpenAIMaxTokens:       maxTokens,
		BlockchainAPI:         blockchainAPI,
		EnableWebSearch:       enableWebSearch,
		OpenAIToolChoice:      toolChoice,
		OpenAIServiceTier:     serviceTier,
		OpenAIReasoningEffort: reasoningEffort,
	}

	return cfg, nil
}
