package main

import (
	"context"
	"log"
	"os"
	"time"


	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

func main() {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN env var is required")
	}
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY env var is required")
	}


	pref := tb.Settings{
		Token:   telegramToken,
		Poller:  &tb.LongPoller{Timeout: 10 * time.Second},
		Verbose: true,
	}

	bot, err := tb.NewBot(pref)

	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	log.Printf("\u2705 Bot up as @%s (ID: %d), listening...", bot.Me.Username, bot.Me.ID)

	client := openai.NewClient(openaiKey)

	bot.Handle("/start", func(c tb.Context) error {
		msg := "Hello! Send me any message and I'll ask ChatGPT to reply.\n" +
			"Use /task <instruction> to forward a specific command."
		return c.Send(msg)
	})

	bot.Handle("/help", func(c tb.Context) error {
		return c.Send("Available commands:\n" +
			"/start - show welcome message\n" +
			"/task <text> - send instruction to ChatGPT\n" +
			"/ping - check bot responsiveness")
	})

	bot.Handle("/ping", func(c tb.Context) error {
		return c.Send("pong")
	})

	bot.Handle("/task", func(c tb.Context) error {
		prompt := c.Message().Payload
		if prompt == "" {
			return c.Send("usage: /task <your instruction>")
		}

		resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{{
				Role:    "user",
				Content: prompt,
			}},
		})
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("failed to get response from ChatGPT")
		}

		if len(resp.Choices) > 0 {
			return c.Send(resp.Choices[0].Message.Content)
		}
		return nil
	})

	bot.Handle(tb.OnText, func(c tb.Context) error {
		prompt := c.Text()
		if prompt == "" {
			return nil
		}

		resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{{
				Role:    "user",
				Content: prompt,
			}},
		})
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("failed to get response from ChatGPT")
		}

		if len(resp.Choices) > 0 {
			return c.Send(resp.Choices[0].Message.Content)
		}
		return nil
	})

	bot.Start()
}
