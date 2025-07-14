package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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

const openAITimeout = 40 * time.Second
const startupMessage = "джарвис в сети, обновление произошло успешно"

// ChatCompleter abstracts the OpenAI client method used by chatCompletion.
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// MessageSender is implemented by types that can send Telegram messages.
type MessageSender interface {
	Send(recipient tb.Recipient, what interface{}, opts ...interface{}) (*tb.Message, error)
}

var (
	currentModel = "gpt-4o"
	modelMu      sync.RWMutex
)

// chatCompletion sends messages to OpenAI and returns the reply text using the current model.
func chatCompletion(ctx context.Context, client ChatCompleter, msgs []openai.ChatCompletionMessage) (string, error) {
	modelMu.RLock()
	m := currentModel
	modelMu.RUnlock()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
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

// systemCompletion generates a reply to a system-level prompt using OpenAI.
func systemCompletion(ctx context.Context, client ChatCompleter, prompt string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}}
	return chatCompletion(ctx, client, msgs)
}

// userCompletion generates a reply to a user's message using OpenAI.
func userCompletion(ctx context.Context, client ChatCompleter, message string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: message}}
	return chatCompletion(ctx, client, msgs)
}

// scheduleDailyMessages sets up the daily lunch idea and brief messages.
func scheduleDailyMessages(s *gocron.Scheduler, client ChatCompleter, bot *tb.Bot, chatID int64) {
	lunchTime := os.Getenv("LUNCH_TIME")
	if lunchTime == "" {
		lunchTime = "13:00"
	}
	briefTime := os.Getenv("BRIEF_TIME")
	if briefTime == "" {
		briefTime = "20:00"
	}

	if _, err := s.Every(1).Day().At(lunchTime).Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), openAITimeout)
		defer cancel()

		text, err := systemCompletion(ctx, client, lunchIdeaPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := bot.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	}); err != nil {
		log.Printf("schedule job: %v", err)
	}

	if _, err := s.Every(1).Day().At(briefTime).Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), openAITimeout)
		defer cancel()

		text, err := systemCompletion(ctx, client, dailyBriefPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := bot.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	}); err != nil {
		log.Printf("schedule job: %v", err)
	}
}

// sendStartupMessage notifies the chat that the bot is running.
func sendStartupMessage(bot MessageSender, chatID int64) {
	if _, err := bot.Send(tb.ChatID(chatID), startupMessage); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.OpenAIModel != "" {
		currentModel = cfg.OpenAIModel
	}

	bot, err := tb.NewBot(tb.Settings{Token: cfg.TelegramToken})
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}
	log.Printf("Authorized as %s", bot.Me.Username)

	oaCfg := openai.DefaultConfig(cfg.OpenAIKey)
	oaCfg.HTTPClient = &http.Client{Timeout: openAITimeout}
	client := openai.NewClientWithConfig(oaCfg)

	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}

	scheduler := gocron.NewScheduler(moscowTZ)
	scheduleDailyMessages(scheduler, client, bot, cfg.ChatID)

	log.Println("Scheduler started. Sending briefs…")
	scheduler.StartAsync()

	sendStartupMessage(bot, cfg.ChatID)

	bot.Handle("/ping", func(c tb.Context) error {
		return c.Send("pong")
	})

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

	bot.Handle("/lunch", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), openAITimeout)
		defer cancel()

		text, err := systemCompletion(ctx, client, lunchIdeaPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	bot.Handle("/brief", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), openAITimeout)
		defer cancel()

		text, err := systemCompletion(ctx, client, dailyBriefPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	bot.Handle("/chat", func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), openAITimeout)
		defer cancel()

		text, err := userCompletion(ctx, client, q)
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("OpenAI error")
		}
		_, err = c.Bot().Send(c.Sender(), text)
		return err
	})

	bot.Start()
}
