package bot

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
	"telegram-reminder/internal/config"
)

// Prompt templates
const (
	DailyBriefPrompt = `
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

	LunchIdeaPrompt = `
Подавай одну бизнес‑идею + примерный план из 4‑5 пунктов (коротко) + ссылки на релевантные ресурсы/репо/доки. Стиль панибратский, минимум воды.
`
)

const OpenAITimeout = 40 * time.Second
const StartupMessage = "джарвис в сети, обновление произошло успешно"

// ChatCompleter abstracts the OpenAI client method used by chatCompletion.
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// MessageSender is implemented by types that can send Telegram messages.
type MessageSender interface {
	Send(recipient tb.Recipient, what interface{}, opts ...interface{}) (*tb.Message, error)
}

var (
	CurrentModel    = "gpt-4o"
	ModelMu         sync.RWMutex
	SupportedModels = []string{
		openai.GPT4o,
		openai.GPT4Turbo,
		openai.GPT3Dot5Turbo,
	}
)

// chatCompletion sends messages to OpenAI and returns the reply text using the current model.
func ChatCompletion(ctx context.Context, client ChatCompleter, msgs []openai.ChatCompletionMessage) (string, error) {
	ModelMu.RLock()
	m := CurrentModel
	ModelMu.RUnlock()

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
func SystemCompletion(ctx context.Context, client ChatCompleter, prompt string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}}
	return ChatCompletion(ctx, client, msgs)
}

// userCompletion generates a reply to a user's message using OpenAI.
func UserCompletion(ctx context.Context, client ChatCompleter, message string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: message}}
	return ChatCompletion(ctx, client, msgs)
}

// scheduleDailyMessages sets up the daily lunch idea and brief messages.
func ScheduleDailyMessages(s *gocron.Scheduler, client ChatCompleter, b *tb.Bot, chatID int64) {
	lunchTime := os.Getenv("LUNCH_TIME")
	if lunchTime == "" {
		lunchTime = "13:00"
	}
	briefTime := os.Getenv("BRIEF_TIME")
	if briefTime == "" {
		briefTime = "20:00"
	}

	if _, err := s.Every(1).Day().At(lunchTime).Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()

		text, err := SystemCompletion(ctx, client, LunchIdeaPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := b.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	}); err != nil {
		log.Printf("schedule job: %v", err)
	}

	if _, err := s.Every(1).Day().At(briefTime).Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()

		text, err := SystemCompletion(ctx, client, DailyBriefPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return
		}
		if _, err := b.Send(tb.ChatID(chatID), text); err != nil {
			log.Printf("telegram send error: %v", err)
		}
	}); err != nil {
		log.Printf("schedule job: %v", err)
	}
}

// SendStartupMessage notifies the chat that the bot is running.
func SendStartupMessage(b MessageSender, chatID int64) {
	if _, err := b.Send(tb.ChatID(chatID), StartupMessage); err != nil {
		log.Printf("telegram send error: %v", err)
	}
}

// Run initializes and starts the Telegram bot.
func Run(cfg config.Config) error {
	if cfg.OpenAIModel != "" {
		CurrentModel = cfg.OpenAIModel
	}

	b, err := tb.NewBot(tb.Settings{Token: cfg.TelegramToken})
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	log.Printf("Authorized as %s", b.Me.Username)

	oaCfg := openai.DefaultConfig(cfg.OpenAIKey)
	oaCfg.HTTPClient = &http.Client{Timeout: OpenAITimeout}
	client := openai.NewClientWithConfig(oaCfg)

	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return fmt.Errorf("failed to load timezone: %w", err)
	}

	scheduler := gocron.NewScheduler(moscowTZ)
	ScheduleDailyMessages(scheduler, client, b, cfg.ChatID)

	log.Println("Scheduler started. Sending briefs…")
	scheduler.StartAsync()

	SendStartupMessage(b, cfg.ChatID)

	b.Handle("/ping", func(c tb.Context) error {
		return c.Send("pong")
	})

	b.Handle("/model", func(c tb.Context) error {
		payload := strings.TrimSpace(c.Message().Payload)
		if payload == "" {
			ModelMu.RLock()
			cur := CurrentModel
			ModelMu.RUnlock()
			return c.Send(fmt.Sprintf(
				"Current model: %s\nSupported: %s",
				cur, strings.Join(SupportedModels, ", "),
			))
		}
		ModelMu.Lock()
		CurrentModel = payload
		ModelMu.Unlock()
		return c.Send(fmt.Sprintf("Model set to %s", payload))
	})

	b.Handle("/lunch", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()

		text, err := SystemCompletion(ctx, client, LunchIdeaPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	b.Handle("/brief", func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()

		text, err := SystemCompletion(ctx, client, DailyBriefPrompt)
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	})

	b.Handle("/chat", func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()

		text, err := UserCompletion(ctx, client, q)
		if err != nil {
			log.Printf("openai error: %v", err)
			return c.Send("OpenAI error")
		}
		_, err = c.Bot().Send(c.Sender(), text)
		return err
	})

	b.Start()
	return nil
}
