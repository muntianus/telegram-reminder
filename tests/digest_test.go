package main

import (
	"context"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type fakeDigestClient struct{ text string }

func (f fakeDigestClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{
		{Message: openai.ChatCompletionMessage{Content: f.text}},
	}}, nil
}

type digestCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

func (d *digestCtx) Send(what interface{}, opts ...interface{}) error {
	d.called = true
	d.msg = what
	return nil
}

func TestLunchCommand(t *testing.T) {
	client := fakeDigestClient{text: "idea"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/lunch", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.LunchIdeaPrompt, "gpt-4.1")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCtx{}
	if err := bot.Trigger("/lunch", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "idea" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestBriefCommand(t *testing.T) {
	client := fakeDigestClient{text: "brief"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/brief", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.DailyBriefPrompt, "gpt-4.1")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCtx{}
	if err := bot.Trigger("/brief", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "brief" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}
