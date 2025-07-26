package main

import (
	"context"
	"strings"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type fakeClient struct{ text string }

func (f fakeClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: f.text}}}}, nil
}

type taskCmdCtx struct {
	tb.Context
	msg    *tb.Message
	called bool
	sent   interface{}
}

func (c *taskCmdCtx) Message() *tb.Message { return c.msg }

func (c *taskCmdCtx) Send(what interface{}, opts ...interface{}) error {
	c.called = true
	c.sent = what
	return nil
}

func TestTaskCommand(t *testing.T) {
	botpkg.TasksMu.Lock()
	botpkg.LoadedTasks = []botpkg.Task{{Name: "foo", Prompt: "p"}, {Name: "bar", Prompt: "p2"}}
	botpkg.TasksMu.Unlock()

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	client := fakeClient{text: "resp"}

	b.Handle("/task", func(c tb.Context) error {
		name := strings.TrimSpace(c.Message().Payload)
		botpkg.TasksMu.RLock()
		tasks := append([]botpkg.Task(nil), botpkg.LoadedTasks...)
		botpkg.TasksMu.RUnlock()
		if name == "" {
			return c.Send(botpkg.FormatTaskNames(tasks))
		}
		tsk, ok := botpkg.FindTask(tasks, name)
		if !ok {
			return c.Send("unknown task")
		}
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()
		text, err := botpkg.SystemCompletion(ctx, client, tsk.Prompt, "gpt-4.1")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	listCtx := &taskCmdCtx{msg: &tb.Message{Payload: ""}}
	if err := b.Trigger("/task", listCtx); err != nil {
		t.Fatalf("trigger list: %v", err)
	}
	if listCtx.sent != "foo\nbar" {
		t.Errorf("unexpected list: %v", listCtx.sent)
	}

	runCtx := &taskCmdCtx{msg: &tb.Message{Payload: "foo"}}
	if err := b.Trigger("/task", runCtx); err != nil {
		t.Fatalf("trigger run: %v", err)
	}
	if runCtx.sent != "resp" {
		t.Errorf("unexpected run resp: %v", runCtx.sent)
	}
}
