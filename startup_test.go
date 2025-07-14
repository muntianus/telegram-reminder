// startup_test.go тестирует сообщение о запуске бота.
package main

import (
	"testing"

	tb "gopkg.in/telebot.v3"
)

type fakeBot struct {
	recipient tb.Recipient
	message   interface{}
	called    bool
}

// Send сохраняет параметры вызова для проверки.
func (f *fakeBot) Send(recipient tb.Recipient, what interface{}, opts ...interface{}) (*tb.Message, error) {
	f.recipient = recipient
	f.message = what
	f.called = true
	return nil, nil
}

// TestSendStartupMessage проверяет отправку приветствия при старте.
func TestSendStartupMessage(t *testing.T) {
	fb := &fakeBot{}
	sendStartupMessage(fb, 42)
	if !fb.called {
		t.Fatal("send not called")
	}
	if id, ok := fb.recipient.(tb.ChatID); !ok || int64(id) != 42 {
		t.Errorf("wrong recipient: %v", fb.recipient)
	}
	if fb.message != startupMessage {
		t.Errorf("unexpected message: %v", fb.message)
	}
}
