package main

import (
	"testing"

	"telegram-reminder/internal/config"
)

func TestLoadConfigSuccess(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "99")
	t.Setenv("OPENAI_API_KEY", "key")
	t.Setenv("OPENAI_MODEL", "model")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramToken != "token" || cfg.ChatID != 99 || cfg.OpenAIKey != "key" || cfg.OpenAIModel != "model" {
		t.Fatalf("unexpected values: %+v", cfg)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "")
	t.Setenv("OPENAI_API_KEY", "key")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigBadChatID(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "bad")
	t.Setenv("OPENAI_API_KEY", "key")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigNoChatID(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "")
	t.Setenv("OPENAI_API_KEY", "key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ChatID != 0 {
		t.Fatalf("unexpected chat id: %d", cfg.ChatID)
	}
}
