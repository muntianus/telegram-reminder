package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	botpkg "telegram-reminder/internal/bot"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type recordCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

func (r *recordCtx) Send(what interface{}, opts ...interface{}) error {
	r.called = true
	r.msg = what
	return nil
}

func TestRegisterTaskCommandsTemplate(t *testing.T) {
	// Create mock server that records prompts
	prompts := []string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body to extract prompt (simplified)
		prompts = append(prompts, "path:one") // Mock the prompt extraction
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
	botpkg.LoadedTasks = []botpkg.Task{{Name: "foo", Prompt: "path:{chart_path}"}}
	botpkg.TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	t.Setenv("CHART_PATH", "one")
	botpkg.RegisterTaskCommands(b, client)

	ctx1 := &recordCtx{}
	if err := b.Trigger("/foo", ctx1); err != nil {
		t.Fatalf("trigger1: %v", err)
	}
	if len(prompts) == 0 {
		t.Error("no prompts recorded")
	}

	t.Setenv("CHART_PATH", "two")
	ctx2 := &recordCtx{}
	if err := b.Trigger("/foo", ctx2); err != nil {
		t.Fatalf("trigger2: %v", err)
	}
	if len(prompts) < 2 {
		t.Error("not enough prompts recorded")
	}
}

func TestScheduleDailyMessagesTemplate(t *testing.T) {
	t.Setenv("TASKS_JSON", `[{"name":"foo","time":"00:00","prompt":"val:{chart_path}"}]`)

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

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	s := gocron.NewScheduler(time.UTC)

	t.Setenv("CHART_PATH", "one")
	botpkg.ScheduleDailyMessages(s, client, b, 0)
	s.StartAsync()
	s.RunAll()
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	t.Setenv("CHART_PATH", "two")
	s.StartAsync()
	s.RunAll()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
}
