package main

import (
	"testing"

	"telegram-reminder/internal/config"
)

func TestLoadConfigSuccess(t *testing.T) {
	t.Setenv(config.EnvTelegramToken, "token")
	t.Setenv(config.EnvChatID, "99")
	t.Setenv(config.EnvLogChatID, "100")
	t.Setenv(config.EnvOpenAIKey, "key")
	t.Setenv(config.EnvOpenAIModel, "model")
	t.Setenv(config.EnvBlockchainAPI, "http://example.com")
	t.Setenv(config.EnvEnableWebSearch, "true")
	t.Setenv(config.EnvSearchProviderURL, "http://search")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramToken != "token" || cfg.ChatID != 99 || cfg.LogChatID != 100 || cfg.OpenAIKey != "key" || cfg.OpenAIModel != "model" || cfg.BlockchainAPI != "http://example.com" || !cfg.EnableWebSearch || cfg.SearchProviderURL != "http://search" {
		t.Fatalf("unexpected values: %+v", cfg)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Setenv(config.EnvTelegramToken, "")
	t.Setenv(config.EnvOpenAIKey, "key")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigBadChatID(t *testing.T) {
	t.Setenv(config.EnvTelegramToken, "token")
	t.Setenv(config.EnvChatID, "bad")
	t.Setenv(config.EnvLogChatID, "101")
	t.Setenv(config.EnvOpenAIKey, "key")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigNoChatID(t *testing.T) {
	t.Setenv(config.EnvTelegramToken, "token")
	t.Setenv(config.EnvChatID, "")
	t.Setenv(config.EnvLogChatID, "")
	t.Setenv(config.EnvOpenAIKey, "key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ChatID != 0 {
		t.Fatalf("unexpected chat id: %d", cfg.ChatID)
	}
}

func TestLoadConfigBadLogChatID(t *testing.T) {
	t.Setenv(config.EnvTelegramToken, "token")
	t.Setenv(config.EnvChatID, "123")
	t.Setenv(config.EnvLogChatID, "bad")
	t.Setenv(config.EnvOpenAIKey, "key")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv(config.EnvTelegramToken, "token")
	t.Setenv(config.EnvOpenAIKey, "key")
	t.Setenv(config.EnvEnableWebSearch, "")
	t.Setenv(config.EnvSearchProviderURL, "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.EnableWebSearch {
		t.Fatalf("expected web search enabled by default")
	}
	if cfg.SearchProviderURL == "" {
		t.Fatalf("expected default search provider URL set")
	}
}
