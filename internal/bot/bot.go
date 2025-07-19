package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"telegram-reminder/internal/config"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
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
const BlockchainTimeout = 10 * time.Second

const Version = "0.1.0"

const CommandsList = `/chat <сообщение> – задать боту вопрос
/ping – проверка состояния
/start – добавить текущий чат в рассылку
/whitelist – показать список подключённых чатов
/remove <id> – убрать чат из списка
/model [имя] – показать или сменить модель (по умолчанию gpt-4o)
/lunch – немедленно запросить идеи на обед
/brief – немедленно запросить вечерний дайджест
/tasks – вывести текущее расписание задач
/task [имя] – список задач или запуск выбранной
/blockchain – метрики сети биткоина`

var StartupMessage = fmt.Sprintf("Billion Roadmap %s\n\n%s", Version, CommandsList)

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
	BasePrompt   string
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

// applyTemplate replaces placeholders in the prompt with runtime values.
func applyTemplate(prompt string) string {
	vars := map[string]string{
		"base_prompt":  BasePrompt,
		"date":         time.Now().Format("2006-01-02"),
		"exchange_api": os.Getenv("EXCHANGE_API"),
		"chart_path":   os.Getenv("CHART_PATH"),
	}
	for k, v := range vars {
		prompt = strings.ReplaceAll(prompt, "{"+k+"}", v)
	}
	return prompt
}

// RegisterTaskCommands creates bot handlers for all named tasks.
func RegisterTaskCommands(b *tb.Bot, client ChatCompleter) {
	TasksMu.RLock()
	tasks := append([]Task(nil), LoadedTasks...)
	TasksMu.RUnlock()
	for _, t := range tasks {
		if t.Name == "" {
			continue
		}
		tcopy := t
		cmd := "/" + t.Name
		b.Handle(cmd, func(c tb.Context) error {
			ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
			defer cancel()

			prompt := applyTemplate(tcopy.Prompt)
			text, err := SystemCompletion(ctx, client, prompt)
			if err != nil {
				logger.L.Error("openai error", "err", err)
				return c.Send("OpenAI error")
			}
			return c.Send(text)
		})
	}
}

// scheduleDailyMessages sets up the daily lunch idea and brief messages.
func ScheduleDailyMessages(s *gocron.Scheduler, client ChatCompleter, b *tb.Bot, chatID int64) {
	tasks, err := LoadTasks()
	if err != nil {
		logger.L.Error("load tasks", "err", err)
		return
	}

	TasksMu.Lock()
	LoadedTasks = tasks
	TasksMu.Unlock()

	for _, t := range tasks {
		tcopy := t
		job := func() {
			ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
			defer cancel()

			log.Printf("running task: %s", tcopy.Name)
			prompt := applyTemplate(tcopy.Prompt)
			text, err := SystemCompletion(ctx, client, prompt)
			if err != nil {
				logger.L.Error("openai error", "err", err)
				return
			}
			if chatID != 0 {
				if _, err := b.Send(tb.ChatID(chatID), text); err != nil {
					log.Printf("telegram send error: %v", err)
				} else {
					log.Printf("sent to chat_id: %d", chatID)
				}
				return
			}
			ids, err := LoadWhitelist()
			if err != nil {
				logger.L.Error("load whitelist", "err", err)
				return
			}
			for _, id := range ids {
				if _, err := b.Send(tb.ChatID(id), text); err != nil {
					log.Printf("telegram send error: %v", err)
				} else {
					log.Printf("sent to chat_id: %d", id)
				}
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
			logger.L.Error("schedule job", "err", jerr)
		}
	}
}

// SendStartupMessage notifies the chat that the bot is running.
func SendStartupMessage(b MessageSender, chatID int64) {
	if chatID != 0 {
		if _, err := b.Send(tb.ChatID(chatID), StartupMessage); err != nil {
			logger.L.Error("telegram send", "err", err)
		}
		return
	}
	ids, err := LoadWhitelist()
	if err != nil {
		logger.L.Error("load whitelist", "err", err)
		return
	}
	for _, id := range ids {
		if _, err := b.Send(tb.ChatID(id), StartupMessage); err != nil {
			logger.L.Error("telegram send", "err", err)
		}
	}
}

// --- HANDLER FUNCTIONS ---

func handlePing(c tb.Context) error {
	return c.Send("pong")
}

func handleStart(c tb.Context) error {
	if err := AddIDToWhitelist(c.Chat().ID); err != nil {
		log.Printf("whitelist add: %v", err)
	}
	return c.Send("Бот активирован")
}

func handleWhitelist(c tb.Context) error {
	ids, err := LoadWhitelist()
	if err != nil {
		log.Printf("load whitelist: %v", err)
		return c.Send("whitelist error")
	}
	if len(ids) == 0 {
		return c.Send("Whitelist is empty")
	}
	return c.Send(FormatWhitelist(ids))
}

func handleRemove(c tb.Context) error {
	payload := strings.TrimSpace(c.Message().Payload)
	if payload == "" {
		return c.Send("Usage: /remove <id>")
	}
	id, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return c.Send("Bad ID")
	}
	if err := RemoveIDFromWhitelist(id); err != nil {
		log.Printf("remove id: %v", err)
		return c.Send("remove error")
	}
	return c.Send("Removed")
}

func handleTasks(c tb.Context) error {
	TasksMu.RLock()
	tasks := append([]Task(nil), LoadedTasks...)
	TasksMu.RUnlock()
	return c.Send(FormatTasks(tasks))
}

func handleTask(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		name := strings.TrimSpace(c.Message().Payload)
		TasksMu.RLock()
		tasks := append([]Task(nil), LoadedTasks...)
		TasksMu.RUnlock()
		if name == "" {
			return c.Send(FormatTaskNames(tasks))
		}
		t, ok := FindTask(tasks, name)
		if !ok {
			return c.Send("unknown task")
		}
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(t.Prompt)
		text, err := SystemCompletion(ctx, client, prompt)
		if err != nil {
			logger.L.Error("openai error", "err", err)
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	}
}

func handleModel() func(tb.Context) error {
	return func(c tb.Context) error {
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
	}
}

func handleLunch(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		text, err := SystemCompletion(ctx, client, LunchIdeaPrompt)
		if err != nil {
			logger.L.Error("openai error", "err", err)
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	}
}

func handleBrief(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		text, err := SystemCompletion(ctx, client, DailyBriefPrompt)
		if err != nil {
			logger.L.Error("openai error", "err", err)
			return c.Send("OpenAI error")
		}
		return c.Send(text)
	}
}

func handleBlockchain(apiURL string) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), BlockchainTimeout)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			logger.L.Error("blockchain req", "err", err)
			return c.Send("blockchain error")
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.L.Error("blockchain call", "err", err)
			return c.Send("blockchain error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logger.L.Error("blockchain status", "status", resp.Status)
			return c.Send("blockchain error")
		}
		var st struct {
			MarketPriceUSD float64 `json:"market_price_usd"`
			NTx            int64   `json:"n_tx"`
			HashRate       float64 `json:"hash_rate"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
			logger.L.Error("blockchain decode", "err", err)
			return c.Send("blockchain error")
		}
		msg := fmt.Sprintf("BTC price: $%.2f\nTransactions: %d\nHash rate: %.2f", st.MarketPriceUSD, st.NTx, st.HashRate)
		return c.Send(msg)
	}
}

func handleChat(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		text, err := UserCompletion(ctx, client, q)
		if err != nil {
			logger.L.Error("openai error", "err", err)
			return c.Send("OpenAI error")
		}
		_, err = c.Bot().Send(c.Sender(), text)
		return err
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

	if cfg.ChatID != 0 {
		if err := AddIDToWhitelist(cfg.ChatID); err != nil {
			log.Printf("whitelist add: %v", err)
		}
	}

	oaCfg := openai.DefaultConfig(cfg.OpenAIKey)
	oaCfg.HTTPClient = &http.Client{Timeout: OpenAITimeout}
	client := openai.NewClientWithConfig(oaCfg)

	moscowTZ, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return fmt.Errorf("failed to load timezone: %w", err)
	}

	scheduler := gocron.NewScheduler(moscowTZ)
	ScheduleDailyMessages(scheduler, client, b, cfg.ChatID)
	RegisterTaskCommands(b, client)

	log.Println("Scheduler started. Sending briefs…")
	scheduler.StartAsync()

	SendStartupMessage(b, cfg.ChatID)

	b.Handle("/ping", handlePing)
	b.Handle("/start", handleStart)
	b.Handle("/whitelist", handleWhitelist)
	b.Handle("/remove", handleRemove)
	b.Handle("/tasks", handleTasks)
	b.Handle("/task", handleTask(client))
	b.Handle("/model", handleModel())
	b.Handle("/lunch", handleLunch(client))
	b.Handle("/brief", handleBrief(client))
	b.Handle("/blockchain", handleBlockchain(cfg.BlockchainAPI))
	b.Handle("/chat", handleChat(client))

	b.Start()
	return nil
}
