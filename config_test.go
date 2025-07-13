package main

import "testing"

func TestLoadConfigSuccess(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "99")
	t.Setenv("OPENAI_API_KEY", "key")
	t.Setenv("OPENAI_MODEL", "model")

	tok, chatID, key, model, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "token" || chatID != 99 || key != "key" || model != "model" {
		t.Fatalf("unexpected values: %v %v %v %v", tok, chatID, key, model)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "")
	t.Setenv("CHAT_ID", "99")
	t.Setenv("OPENAI_API_KEY", "key")

	_, _, _, _, err := loadConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigBadChatID(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "token")
	t.Setenv("CHAT_ID", "bad")
	t.Setenv("OPENAI_API_KEY", "key")

	_, _, _, _, err := loadConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}
