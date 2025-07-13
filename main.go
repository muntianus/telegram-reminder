package main

import (
	"context"
	"log"
	"os"

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

	bot, err := tb.NewBot(tb.Settings{
		Token: telegramToken,
	})
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	client := openai.NewClient(openaiKey)

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
