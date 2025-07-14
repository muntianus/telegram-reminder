package main

import (
	"context"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
	botpkg "telegram-reminder/internal/bot"
)

type recordClient struct{ prompts []string }

func (r *recordClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	p := ""
	if len(req.Messages) > 0 {
		p = req.Messages[0].Content
	}
	r.prompts = append(r.prompts, p)
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{
		{Message: openai.ChatCompletionMessage{Content: "resp"}},
	}}, nil
}

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
	botpkg.TasksMu.Lock()
	botpkg.LoadedTasks = []botpkg.Task{{Name: "foo", Prompt: "path:{chart_path}"}}
	botpkg.TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	client := &recordClient{}
	t.Setenv("CHART_PATH", "one")
	botpkg.RegisterTaskCommands(b, client)

	ctx1 := &recordCtx{}
	if err := b.Trigger("/foo", ctx1); err != nil {
		t.Fatalf("trigger1: %v", err)
	}
	if client.prompts[0] != "path:one" {
		t.Errorf("first prompt %q", client.prompts[0])
	}

	t.Setenv("CHART_PATH", "two")
	ctx2 := &recordCtx{}
	if err := b.Trigger("/foo", ctx2); err != nil {
		t.Fatalf("trigger2: %v", err)
	}
	if client.prompts[1] != "path:two" {
		t.Errorf("second prompt %q", client.prompts[1])
	}
}

func TestScheduleDailyMessagesTemplate(t *testing.T) {
	t.Setenv("TASKS_JSON", `[{"name":"foo","time":"00:00","prompt":"val:{chart_path}"}]`)

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	s := gocron.NewScheduler(time.UTC)
	client := &recordClient{}

	t.Setenv("CHART_PATH", "one")
	botpkg.ScheduleDailyMessages(s, client, b, 0)
	s.StartAsync()
	s.RunAll()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	if len(client.prompts) == 0 || client.prompts[0] != "val:one" {
		t.Fatalf("first run prompt %v", client.prompts)
	}

	t.Setenv("CHART_PATH", "two")
	s.StartAsync()
	s.RunAll()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	if len(client.prompts) < 2 || client.prompts[1] != "val:two" {
		t.Fatalf("second run prompt %v", client.prompts)
	}
}
