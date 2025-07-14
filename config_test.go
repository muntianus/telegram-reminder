package main

import (
	"os"
	"path/filepath"
	"testing"

	"telegram-reminder/internal/config"
)

func TestLoadConfigSuccess(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "99")
	t.Setenv("OPENAI_API_KEY", "key")
	t.Setenv("OPENAI_MODEL", "model")
	t.Setenv("BLOCKCHAIN_API", "http://example.com")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("WHITELIST_FILE", "w.json")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramToken != "token" || cfg.ChatID != 99 || cfg.OpenAIKey != "key" || cfg.OpenAIModel != "model" || cfg.BlockchainAPI != "http://example.com" || cfg.LogLevel != "debug" || cfg.WhitelistFile != "w.json" {
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

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	fn := filepath.Join(dir, "cfg.yml")
	err := os.WriteFile(fn, []byte("log_level: warn\nwhitelist_file: wl.json\nblockchain_api: http://example.com\nopenai_model: modx\n"), 0644)
	if err != nil {
		t.Fatalf("write file: %v", err)
	}
	t.Setenv("CONFIG_FILE", fn)
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("OPENAI_API_KEY", "key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "warn" || cfg.WhitelistFile != "wl.json" || cfg.BlockchainAPI != "http://example.com" || cfg.OpenAIModel != "modx" {
		t.Fatalf("unexpected values: %+v", cfg)
	}
}
