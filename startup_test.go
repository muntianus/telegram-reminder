package main

import (
	"testing"

	tb "gopkg.in/telebot.v3"
	"telegram-reminder/internal/bot"
)

type fakeBot struct {
	recipient tb.Recipient
	message   interface{}
	called    bool
}

func (f *fakeBot) Send(recipient tb.Recipient, what interface{}, opts ...interface{}) (*tb.Message, error) {
	f.recipient = recipient
	f.message = what
	f.called = true
	return nil, nil
}

func TestSendStartupMessage(t *testing.T) {
	fb := &fakeBot{}
	bot.SendStartupMessage(fb, 42)
	if !fb.called {
		t.Fatal("send not called")
	}
	if id, ok := fb.recipient.(tb.ChatID); !ok || int64(id) != 42 {
		t.Errorf("wrong recipient: %v", fb.recipient)
	}
	if fb.message != bot.StartupMessage {
		t.Errorf("unexpected message: %v", fb.message)
	}
}
