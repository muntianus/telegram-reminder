package main

import "testing"

func TestLoadConfigSuccess(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "tok")
	t.Setenv("CHAT_ID", "42")
	t.Setenv("OPENAI_API_KEY", "key")
	t.Setenv("OPENAI_MODEL", "m")
	tok, id, k, model, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "tok" || id != 42 || k != "key" || model != "m" {
		t.Fatalf("got %q %d %q %q", tok, id, k, model)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "")
	t.Setenv("CHAT_ID", "")
	t.Setenv("OPENAI_API_KEY", "key")
	_, _, _, _, err := loadConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigInvalidChatID(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "tok")
	t.Setenv("CHAT_ID", "bad")
	t.Setenv("OPENAI_API_KEY", "key")
	_, _, _, _, err := loadConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadConfigDefaultModel(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "tok")
	t.Setenv("CHAT_ID", "3")
	t.Setenv("OPENAI_API_KEY", "key")
	t.Setenv("OPENAI_MODEL", "")
	_, _, _, model, err := loadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model != "gpt-4o" {
		t.Fatalf("got model %q", model)
	}
}
