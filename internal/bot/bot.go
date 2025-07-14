package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
	yaml "gopkg.in/yaml.v3"
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
const (
	DefaultLunchTime = "13:00"
	DefaultBriefTime = "20:00"
)

// Task represents a scheduled job definition.
type Task struct {
	Name   string `json:"name" yaml:"name"`
	Prompt string `json:"prompt" yaml:"prompt"`
	Time   string `json:"time,omitempty" yaml:"time,omitempty"`
	Cron   string `json:"cron,omitempty" yaml:"cron,omitempty"`
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// LoadTasks reads task configuration from TASKS_FILE or TASKS_JSON. If neither
// is provided, it falls back to the legacy LUNCH_TIME and BRIEF_TIME
// environment variables.
func LoadTasks() ([]Task, error) {
	if fn := os.Getenv("TASKS_FILE"); fn != "" {
		data, err := os.ReadFile(fn)
		if err != nil {
			return nil, err
		}
		tasks := []Task{}
		ext := strings.ToLower(filepath.Ext(fn))
		if ext == ".yaml" || ext == ".yml" {
			if err := yaml.Unmarshal(data, &tasks); err != nil {
				return nil, err
			}
		} else {
			if err := json.Unmarshal(data, &tasks); err != nil {
				return nil, err
			}
		}
		return tasks, nil
	}

	if txt := os.Getenv("TASKS_JSON"); txt != "" {
		tasks := []Task{}
		if err := json.Unmarshal([]byte(txt), &tasks); err != nil {
			return nil, err
		}
		return tasks, nil
	}

	lunchTime := envDefault("LUNCH_TIME", DefaultLunchTime)
	briefTime := envDefault("BRIEF_TIME", DefaultBriefTime)
	return []Task{
		{Name: "lunch", Prompt: LunchIdeaPrompt, Time: lunchTime},
		{Name: "brief", Prompt: DailyBriefPrompt, Time: briefTime},
	}, nil
}

// ChatCompleter abstracts the OpenAI client method used by chatCompletion.
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// MessageSender is implemented by types that can send Telegram messages.
type MessageSender interface {
	Send(recipient tb.Recipient, what interface{}, opts ...interface{}) (*tb.Message, error)
}

var (
	CurrentModel = "gpt-4o"
	ModelMu      sync.RWMutex
	//nolint:staticcheck // list includes deprecated model constants for completeness
	SupportedModels = []string{
		openai.O1Mini,
		openai.O1Mini20240912,
		openai.O1Preview,
		openai.O1Preview20240912,
		openai.O1,
		openai.O120241217,
		openai.O3,
		openai.O320250416,
		openai.O3Mini,
		openai.O3Mini20250131,
		openai.O4Mini,
		openai.O4Mini20250416,
		openai.GPT432K0613,
		openai.GPT432K0314,
		openai.GPT432K,
		openai.GPT40613,
		openai.GPT40314,
		openai.GPT4o,
		openai.GPT4o20240513,
		openai.GPT4o20240806,
		openai.GPT4o20241120,
		openai.GPT4oLatest,
		openai.GPT4oMini,
		openai.GPT4oMini20240718,
		openai.GPT4Turbo,
		openai.GPT4Turbo20240409,
		openai.GPT4Turbo0125,
		openai.GPT4Turbo1106,
		openai.GPT4TurboPreview,
		openai.GPT4VisionPreview,
		openai.GPT4,
		openai.GPT4Dot1,
		openai.GPT4Dot120250414,
		openai.GPT4Dot1Mini,
		openai.GPT4Dot1Mini20250414,
		openai.GPT4Dot1Nano,
		openai.GPT4Dot1Nano20250414,
		openai.GPT4Dot5Preview,
		openai.GPT4Dot5Preview20250227,
		openai.GPT3Dot5Turbo0125,
		openai.GPT3Dot5Turbo1106,
		openai.GPT3Dot5Turbo0613,
		openai.GPT3Dot5Turbo0301,
		openai.GPT3Dot5Turbo16K,
		openai.GPT3Dot5Turbo16K0613,
		openai.GPT3Dot5Turbo,
		openai.GPT3Dot5TurboInstruct,
		openai.GPT3TextDavinci003,
		openai.GPT3TextDavinci002,
		openai.GPT3TextCurie001,
		openai.GPT3TextBabbage001,
		openai.GPT3TextAda001,
		openai.GPT3TextDavinci001,
		openai.GPT3DavinciInstructBeta,
		openai.GPT3Davinci,
		openai.GPT3Davinci002,
		openai.GPT3CurieInstructBeta,
		openai.GPT3Curie,
		openai.GPT3Curie002,
		openai.GPT3Ada,
		openai.GPT3Ada002,
		openai.GPT3Babbage,
		openai.GPT3Babbage002,
		openai.CodexCodeDavinci002,
		openai.CodexCodeCushman001,
		openai.CodexCodeDavinci001,
	}
)

var (
	LoadedTasks []Task
	TasksMu     sync.RWMutex
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

// FormatTasks returns a text summary of tasks with their time or cron expression.
func FormatTasks(tasks []Task) string {
	if len(tasks) == 0 {
		return "no tasks"
	}
	var b strings.Builder
	for i, t := range tasks {
		when := t.Cron
		if when == "" {
			when = t.Time
			if when == "" {
				when = "00:00"
			}
		}
		name := t.Name
		if name == "" {
			name = fmt.Sprintf("task %d", i+1)
		}
		fmt.Fprintf(&b, "%s - %s\n", when, name)
	}
	return strings.TrimSpace(b.String())
}

// scheduleDailyMessages sets up the daily lunch idea and brief messages.
func ScheduleDailyMessages(s *gocron.Scheduler, client ChatCompleter, b *tb.Bot, chatID int64) {
	tasks, err := LoadTasks()
	if err != nil {
		log.Printf("load tasks: %v", err)
		return
	}

	TasksMu.Lock()
	LoadedTasks = tasks
	TasksMu.Unlock()

	for _, t := range tasks {
		prompt := t.Prompt
		job := func() {
			ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
			defer cancel()

			text, err := SystemCompletion(ctx, client, prompt)
			if err != nil {
				log.Printf("openai error: %v", err)
				return
			}
			if _, err := b.Send(tb.ChatID(chatID), text); err != nil {
				log.Printf("telegram send error: %v", err)
			}
		}

		var jerr error
		switch {
		case t.Cron != "":
			_, jerr = s.Cron(t.Cron).Do(job)
		default:
			timeStr := t.Time
			if timeStr == "" {
				timeStr = "00:00"
			}
			_, jerr = s.Every(1).Day().At(timeStr).Do(job)
		}
		if jerr != nil {
			log.Printf("schedule job: %v", jerr)
		}
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

	b.Handle("/tasks", func(c tb.Context) error {
		TasksMu.RLock()
		tasks := append([]Task(nil), LoadedTasks...)
		TasksMu.RUnlock()
		return c.Send(FormatTasks(tasks))
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
