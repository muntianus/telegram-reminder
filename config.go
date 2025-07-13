package main

import (
	"fmt"
	"os"
	"strconv"
)

// loadConfig reads environment variables and validates them.
// TELEGRAM_TOKEN, CHAT_ID and OPENAI_API_KEY are required.
// OPENAI_MODEL is optional.
func loadConfig() (telegramToken string, chatID int64, openaiKey string, openaiModel string, err error) {
	telegramToken = os.Getenv("TELEGRAM_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")
	openaiKey = os.Getenv("OPENAI_API_KEY")
	openaiModel = os.Getenv("OPENAI_MODEL")

	if telegramToken == "" || chatIDStr == "" || openaiKey == "" {
		err = fmt.Errorf("missing required env vars")
		return
	}

	chatID, err = strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		err = fmt.Errorf("invalid CHAT_ID: %w", err)
		return
	}

	return
}
