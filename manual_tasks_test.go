package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type taskCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

func (t *taskCtx) Send(what interface{}, opts ...interface{}) error {
	t.called = true
	t.msg = what
	return nil
}

func TestRegisterTaskCommands(t *testing.T) {
	// Create mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "resp"
				}
			}]
		}`))
	}))
	defer srv.Close()

	// Create real client with mock server
	cfg := openai.DefaultConfig("test-key")
	cfg.BaseURL = srv.URL
	client := openai.NewClientWithConfig(cfg)

	botpkg.TasksMu.Lock()
	botpkg.LoadedTasks = []botpkg.Task{{Name: "foo", Prompt: "p"}}
	botpkg.TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	botpkg.RegisterTaskCommands(b, client)

	ctx := &taskCtx{}
	if err := b.Trigger("/foo", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "resp" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}
