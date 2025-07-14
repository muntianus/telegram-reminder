// config_test.go содержит тесты функции загрузки конфигурации.
package main

import "testing"

// TestLoadConfigSuccess проверяет успешное чтение конфигурации.
func TestLoadConfigSuccess(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "99")
	t.Setenv("OPENAI_API_KEY", "key")
	t.Setenv("OPENAI_MODEL", "model")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TelegramToken != "token" || cfg.ChatID != 99 || cfg.OpenAIKey != "key" || cfg.OpenAIModel != "model" {
		t.Fatalf("unexpected values: %+v", cfg)
	}
}

// TestLoadConfigMissing проверяет отсутствие обязательных переменных.
func TestLoadConfigMissing(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "")
	t.Setenv("CHAT_ID", "99")
	t.Setenv("OPENAI_API_KEY", "key")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

// TestLoadConfigBadChatID проверяет обработку неверного идентификатора чата.
func TestLoadConfigBadChatID(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "bad")
	t.Setenv("OPENAI_API_KEY", "key")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}
