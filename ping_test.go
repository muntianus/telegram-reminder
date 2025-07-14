// ping_test.go тестирует обработчик команды /ping.
package main

import (
	"testing"

	tb "gopkg.in/telebot.v3"
)

type fakeCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

// Send сохраняет отправленное сообщение.
func (f *fakeCtx) Send(what interface{}, opts ...interface{}) error {
	f.called = true
	f.msg = what
	return nil
}

// TestPingHandler проверяет ответ на команду /ping.
func TestPingHandler(t *testing.T) {
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}
	bot.Handle("/ping", func(c tb.Context) error {
		return c.Send("pong")
	})
	ctx := &fakeCtx{}
	if err := bot.Trigger("/ping", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "pong" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}
