package main

import (
	"context"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

type fakeDigestCommandsClient struct{ text string }

func (f fakeDigestCommandsClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{
		{Message: openai.ChatCompletionMessage{Content: f.text}},
	}}, nil
}

type digestCommandsCtx struct {
	tb.Context
	called bool
	msg    interface{}
}

func (d *digestCommandsCtx) Send(what interface{}, opts ...interface{}) error {
	d.called = true
	d.msg = what
	return nil
}

func TestCryptoDigestCommand(t *testing.T) {
	client := fakeDigestCommandsClient{text: "🔥 КРИПТО-ДАЙДЖЕСТ ЗА СЕГОДНЯ"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/crypto", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.CryptoDigestPrompt, "gpt-4o")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCommandsCtx{}
	if err := bot.Trigger("/crypto", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "🔥 КРИПТО-ДАЙДЖЕСТ ЗА СЕГОДНЯ" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestTechDigestCommand(t *testing.T) {
	client := fakeDigestCommandsClient{text: "💻 ТЕХ-ДАЙДЖЕСТ НА СЕГОДНЯ"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/tech", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.TechDigestPrompt, "gpt-4o")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCommandsCtx{}
	if err := bot.Trigger("/tech", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "💻 ТЕХ-ДАЙДЖЕСТ НА СЕГОДНЯ" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestRealEstateDigestCommand(t *testing.T) {
	client := fakeDigestCommandsClient{text: "🏠 НЕДВИЖИМОСТЬ: ДАЙДЖЕСТ НА СЕГОДНЯ"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/realestate", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.RealEstateDigestPrompt, "gpt-4o")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCommandsCtx{}
	if err := bot.Trigger("/realestate", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "🏠 НЕДВИЖИМОСТЬ: ДАЙДЖЕСТ НА СЕГОДНЯ" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestBusinessDigestCommand(t *testing.T) {
	client := fakeDigestCommandsClient{text: "💼 БИЗНЕС-ДАЙДЖЕСТ НА СЕГОДНЯ"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/business", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.BusinessDigestPrompt, "gpt-4o")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCommandsCtx{}
	if err := bot.Trigger("/business", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "💼 БИЗНЕС-ДАЙДЖЕСТ НА СЕГОДНЯ" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestInvestmentDigestCommand(t *testing.T) {
	client := fakeDigestCommandsClient{text: "💰 ИНВЕСТИЦИОННЫЙ ДАЙДЖЕСТ НА СЕГОДНЯ"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/investment", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.InvestmentDigestPrompt, "gpt-4o")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCtx{}
	if err := bot.Trigger("/investment", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "💰 ИНВЕСТИЦИОННЫЙ ДАЙДЖЕСТ НА СЕГОДНЯ" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestStartupDigestCommand(t *testing.T) {
	client := fakeDigestClient{text: "🚀 СТАРТАП-ДАЙДЖЕСТ НА СЕГОДНЯ"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/startup", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.StartupDigestPrompt, "gpt-4o")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCtx{}
	if err := bot.Trigger("/startup", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "🚀 СТАРТАП-ДАЙДЖЕСТ НА СЕГОДНЯ" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}

func TestGlobalDigestCommand(t *testing.T) {
	client := fakeDigestClient{text: "🌍 ГЛОБАЛЬНЫЙ ДАЙДЖЕСТ НА СЕГОДНЯ"}
	bot, err := tb.NewBot(tb.Settings{Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}

	bot.Handle("/global", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), botpkg.OpenAITimeout)
		defer cancel()

		text, err := botpkg.SystemCompletion(ctx, client, botpkg.GlobalDigestPrompt, "gpt-4o")
		if err != nil {
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	ctx := &digestCtx{}
	if err := bot.Trigger("/global", ctx); err != nil {
		t.Fatalf("trigger: %v", err)
	}
	if !ctx.called {
		t.Fatal("send not called")
	}
	if ctx.msg != "🌍 ГЛОБАЛЬНЫЙ ДАЙДЖЕСТ НА СЕГОДНЯ" {
		t.Errorf("unexpected message: %v", ctx.msg)
	}
}
