package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
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

var (
	currentModel = "gpt-4o"
	modelMu      sync.RWMutex
)

// chatCompletion sends messages to OpenAI and returns the reply text using the current model.
func chatCompletion(client *openai.Client, msgs []openai.ChatCompletionMessage) (string, error) {
	modelMu.RLock()
	m := currentModel
	modelMu.RUnlock()

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:       m,
		Messages:    msgs,
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

func systemCompletion(client *openai.Client, prompt string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}}
	return chatCompletion(client, msgs)
}

func userCompletion(client *openai.Client, message string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: message}}
	return chatCompletion(client, msgs)
}

func main() {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if telegramToken == "" || chatIDStr == "" || openaiKey == "" {
		log.Fatal("Set TELEGRAM_TOKEN, CHAT_ID, OPENAI_API_KEY env vars")
	}

	if envModel := os.Getenv("OPENAI_MODEL"); envModel != "" {
		currentModel = envModel
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

	scheduler.Every(1).Day().At("13:00").Do(func() {
		text, err := systemCompletion(client, lunchIdeaPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := bot.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	})

	scheduler.Every(1).Day().At("20:00").Do(func() {
		text, err := systemCompletion(client, dailyBriefPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := bot.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	})

	log.Println("Scheduler started. Sending briefs…")
	scheduler.StartAsync()

	bot.Handle("/model", func(c tb.Context) error {
		payload := strings.TrimSpace(c.Message().Payload)
		if payload == "" {
			modelMu.RLock()
			cur := currentModel
			modelMu.RUnlock()
			return c.Send(fmt.Sprintf("Current model: %s", cur))
		}
		modelMu.Lock()
		currentModel = payload
		modelMu.Unlock()
		return c.Send(fmt.Sprintf("Model set to %s", payload))
	})

	bot.Handle("/chat", func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		text, err := userCompletion(client, q)
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("OpenAI error")
		}
		_, err = c.Bot().Send(c.Sender(), text)
		return err
	})

	bot.Start()
}
