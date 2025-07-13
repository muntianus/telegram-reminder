package main

import (
	"fmt"
	"os"
	"strconv"
)

// loadConfig reads environment variables and returns the configuration values.
// TELEGRAM_TOKEN, CHAT_ID and OPENAI_API_KEY must be set. OPENAI_MODEL is
// optional and defaults to "gpt-4o". If CHAT_ID cannot be parsed as int64 or
// required variables are missing, an error is returned.
func loadConfig() (telegramToken string, chatID int64, openaiKey string, openaiModel string, err error) {
	telegramToken = os.Getenv("TELEGRAM_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")
	openaiKey = os.Getenv("OPENAI_API_KEY")
	openaiModel = os.Getenv("OPENAI_MODEL")
	if openaiModel == "" {
		openaiModel = "gpt-4o"
	}

	if telegramToken == "" || chatIDStr == "" || openaiKey == "" {
		err = fmt.Errorf("missing required environment variables")
		return
	}

	chatID, convErr := strconv.ParseInt(chatIDStr, 10, 64)
	if convErr != nil {
		err = fmt.Errorf("invalid CHAT_ID: %w", convErr)
		return
	}
	return
}
