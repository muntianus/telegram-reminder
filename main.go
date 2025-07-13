package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

// Prompt templates
const (
	dailyBriefPrompt = `
Ты говоришь кратко, дерзко, панибратски.
Заполни блоки:
⚡ Микродействие (одно простое действие на сегодня)
🧠 Тема дня (мини‑инсайт/мысль)
💰 Что залутать (актив/идея)
🏞️ Земля на присмотр (лоты в южном Подмосковье: Бутово, Щербинка, Подольск, Воскресенск), дай 1‑2 лота со ссылками.
🪙 Альт дня (актуальная монета, линк CoinGecko)
🚀 Пушка с ProductHunt (ссылка)
Форматируй одним сообщением, без лишней воды.
`

	lunchIdeaPrompt = `
Подавай одну бизнес‑идею + примерный план из 4‑5 пунктов (коротко) + ссылки на релевантные ресурсы/репо/доки. Стиль панибратский, минимум воды.
`
)

// ChatCompleter abstracts the OpenAI client method used by chatCompletion.
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// chatCompletion sends a prompt to OpenAI and returns the reply text.
func chatCompletion(client ChatCompleter, prompt string) (string, error) {
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{

		Model:       "gpt-4o",
		Messages:    []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}},
		Temperature: 0.9,
		MaxTokens:   600,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// setupScheduler configures daily jobs on the provided scheduler.
func setupScheduler(s *gocron.Scheduler, client ChatCompleter, bot *tb.Bot, chatID int64) {
	s.Every(1).Day().At("13:00").Do(func() {
		text, err := chatCompletion(client, lunchIdeaPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := bot.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	})

	s.Every(1).Day().At("20:00").Do(func() {
		text, err := chatCompletion(client, dailyBriefPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := bot.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	})
}

func main() {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if telegramToken == "" || chatIDStr == "" || openaiKey == "" {
		log.Fatal("Set TELEGRAM_TOKEN, CHAT_ID, OPENAI_API_KEY env vars")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("invalid CHAT_ID: %v", err)
	}

	bot, err := tb.NewBot(tb.Settings{Token: telegramToken})
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	client := openai.NewClient(openaiKey)

	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}

	scheduler := gocron.NewScheduler(moscowTZ)
	setupScheduler(scheduler, client, bot, chatID)

	log.Println("Scheduler started. Sending briefs…")
	scheduler.StartBlocking()
}
