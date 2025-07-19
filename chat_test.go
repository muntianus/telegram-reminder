package main

import (
	"context"
	"strings"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

// fakeChatClient implements bot.ChatCompleter and returns a fixed reply.
type fakeChatClient struct{ text string }

func (f fakeChatClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: f.text}}}}, nil
}

// chatCtx captures messages sent by the handler.
type chatCtx struct {
	tb.Context
	msg    *tb.Message
	called bool
	sent   interface{}
}

func (c *chatCtx) Message() *tb.Message { return c.msg }

func (c *chatCtx) Send(what interface{}, opts ...interface{}) error {
	c.called = true
	c.sent = what
	return nil
}

func TestChatCommand(t *testing.T) {
	client := fakeChatClient{text: "hi"}

	b, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	b.Handle("/chat", func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()
		text, err := botpkg.UserCompletion(ctx, client, "test message", "o3")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &chatCtx{msg: &tb.Message{Payload: "hello"}}
	if err := b.Trigger("/chat", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.sent != "hi" {
		t.Errorf("unexpected reply: %v", ctx.sent)
	}

	emptyCtx := &chatCtx{msg: &tb.Message{Payload: ""}}
	if err := b.Trigger("/chat", emptyCtx); err != nil {
		t.Fatalf("trigger empty: %v", err)
	}
	if emptyCtx.sent != "Usage: /chat <message>" {
		t.Errorf("unexpected empty reply: %v", emptyCtx.sent)
	}
}
