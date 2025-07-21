package main

import (
	"testing"

	"telegram-reminder/internal/bot"

	tb "gopkg.in/telebot.v3"
)

func TestSendStartupMessage(t *testing.T) {
	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	// Test with specific chat ID
	bot.SendStartupMessage(b, 42, "test")

	// The test passes if no error occurs
	// In offline mode, the bot won't actually send messages
}
