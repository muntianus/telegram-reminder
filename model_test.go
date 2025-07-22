package main

import (
	"fmt"
	"strings"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	tb "gopkg.in/telebot.v3"
)

type modelFakeCtx struct {
	tb.Context
	msg  *tb.Message
	sent interface{}
}

func (f *modelFakeCtx) Message() *tb.Message { return f.msg }

func (f *modelFakeCtx) Send(what interface{}, opts ...interface{}) error {
	f.sent = what
	return nil
}

func TestModelCommand(t *testing.T) {
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/model", func(c tb.Context) error {
		payload := strings.TrimSpace(c.Message().Payload)
		if payload == "" {
			botpkg.ModelMu.RLock()
			cur := botpkg.CurrentModel
			botpkg.ModelMu.RUnlock()
			return c.Send(fmt.Sprintf(
				"Current model: %s\nSupported: %s",
				cur, strings.Join(botpkg.SupportedModels, ", "),
			))
		}
		if !botpkg.IsSupportedModel(payload) {
			return c.Send("unsupported model")
		}
		botpkg.ModelMu.Lock()
		botpkg.CurrentModel = payload
		botpkg.ModelMu.Unlock()
		return c.Send(fmt.Sprintf("Model set to %s", payload))
	})

	botpkg.ModelMu.Lock()
	botpkg.CurrentModel = "o3"
	botpkg.ModelMu.Unlock()

	ctx := &modelFakeCtx{msg: &tb.Message{Payload: ""}}
	if err := bot.Trigger("/model", ctx); err != nil {
		t.Fatalf("trigger no arg: %v", err)
	}
	if ctx.sent != fmt.Sprintf("Current model: o3\nSupported: %s", strings.Join(botpkg.SupportedModels, ", ")) {
		t.Errorf("unexpected response: %v", ctx.sent)
	}

	ctx2 := &modelFakeCtx{msg: &tb.Message{Payload: "gpt-4o"}}
	if err := bot.Trigger("/model", ctx2); err != nil {
		t.Fatalf("trigger with arg: %v", err)
	}
	if ctx2.sent != "Model set to gpt-4o" {
		t.Errorf("unexpected response: %v", ctx2.sent)
	}

	botpkg.ModelMu.RLock()
	got := botpkg.CurrentModel
	botpkg.ModelMu.RUnlock()
	if got != "gpt-4o" {
		t.Errorf("currentModel not updated: %s", got)
	}

	ctx3 := &modelFakeCtx{msg: &tb.Message{Payload: ""}}
	if err := bot.Trigger("/model", ctx3); err != nil {
		t.Fatalf("trigger query after set: %v", err)
	}
	if ctx3.sent != fmt.Sprintf("Current model: gpt-4o\nSupported: %s", strings.Join(botpkg.SupportedModels, ", ")) {
		t.Errorf("unexpected response: %v", ctx3.sent)
	}

	invalid := &modelFakeCtx{msg: &tb.Message{Payload: "gpt-3"}}
	if err := bot.Trigger("/model", invalid); err != nil {
		t.Fatalf("trigger invalid: %v", err)
	}
	if invalid.sent != "unsupported model" {
		t.Errorf("unexpected invalid response: %v", invalid.sent)
	}

	botpkg.ModelMu.RLock()
	after := botpkg.CurrentModel
	botpkg.ModelMu.RUnlock()
	if after != "gpt-4o" {
		t.Errorf("model changed on invalid input: %s", after)
	}
}
