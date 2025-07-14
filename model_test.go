package main

import (
	"fmt"
	"strings"
	"testing"

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
			modelMu.RLock()
			cur := currentModel
			modelMu.RUnlock()
			return c.Send(fmt.Sprintf("Current model: %s", cur))
		}
		modelMu.Lock()
		currentModel = payload
		modelMu.Unlock()
		return c.Send(fmt.Sprintf("Model set to %s", payload))
	})

	modelMu.Lock()
	currentModel = "gpt-4o"
	modelMu.Unlock()

	ctx := &modelFakeCtx{msg: &tb.Message{Payload: ""}}
	if err := bot.Trigger("/model", ctx); err != nil {
		t.Fatalf("trigger no arg: %v", err)
	}
	if ctx.sent != "Current model: gpt-4o" {
		t.Errorf("unexpected response: %v", ctx.sent)
	}

	ctx2 := &modelFakeCtx{msg: &tb.Message{Payload: "gpt-3"}}
	if err := bot.Trigger("/model", ctx2); err != nil {
		t.Fatalf("trigger with arg: %v", err)
	}
	if ctx2.sent != "Model set to gpt-3" {
		t.Errorf("unexpected response: %v", ctx2.sent)
	}

	modelMu.RLock()
	got := currentModel
	modelMu.RUnlock()
	if got != "gpt-3" {
		t.Errorf("currentModel not updated: %s", got)
	}

	ctx3 := &modelFakeCtx{msg: &tb.Message{Payload: ""}}
	if err := bot.Trigger("/model", ctx3); err != nil {
		t.Fatalf("trigger query after set: %v", err)
	}
	if ctx3.sent != "Current model: gpt-3" {
		t.Errorf("unexpected response: %v", ctx3.sent)
	}
}
