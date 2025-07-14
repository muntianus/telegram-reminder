package main

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
	botpkg "telegram-reminder/internal/bot"
)

type fakeTaskClient struct{ text string }

func (f fakeTaskClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: f.text}}}}, nil
}

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
	botpkg.TasksMu.Lock()
	botpkg.LoadedTasks = []botpkg.Task{{Name: "foo", Prompt: "p"}}
	botpkg.TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	client := fakeTaskClient{text: "resp"}
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
