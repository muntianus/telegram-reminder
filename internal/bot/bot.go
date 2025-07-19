package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"telegram-reminder/internal/config"
	"telegram-reminder/internal/logger"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

// Prompt templates
const (
	DailyBriefPrompt = `
Ты — Telegram-бот для ежедневного дайджеста. Говоришь кратко, дерзко, панибратски.

📅 ВАЖНО: Анализируй информацию ТОЛЬКО за сегодняшний день.

Заполни блоки:
⚡ Микродействие (одно простое действие на сегодня)
🧠 Тема дня (мини‑инсайт/мысль на основе сегодняшних событий)
💰 Что залутать (актив/идея на основе сегодняшних трендов)
🏞️ Земля на присмотр (лоты в южном Подмосковье: Бутово, Щербинка, Подольск, Воскресенск)
🪙 Альт дня (актуальная монета на основе сегодняшних движений, линк CoinGecko)
🚀 Пушка с ProductHunt (сегодняшние топовые проекты)

🔍 ИНТЕРНЕТ-АНАЛИЗ: Используй актуальную информацию из интернета по темам:
- Криптовалюты и DeFi
- Технологии и стартапы
- Недвижимость и инвестиции
- Бизнес-тренды

Форматируй одним сообщением для Telegram, без лишней воды.
`

	LunchIdeaPrompt = `
🚀 БИЗНЕС-ИДЕЯ НА СЕГОДНЯ

Подавай одну бизнес‑идею на основе сегодняшних трендов и событий.
Примерный план из 4‑5 пунктов со ссылками на релевантные ресурсы.
Стиль панибратский, минимум воды.
Используй актуальную информацию из интернета.

Форматируй для Telegram с эмодзи и четкой структурой.
`
)

const OpenAITimeout = 40 * time.Second
const BlockchainTimeout = 10 * time.Second

const Version = "0.1.0"

// formatOpenAIError форматирует ошибку OpenAI для пользователя
func formatOpenAIError(err error, model string) string {
	errStr := err.Error()

	// Определяем тип ошибки по содержимому
	switch {
	case strings.Contains(errStr, "insufficient_quota"):
		return "❌ Недостаточно кредитов на аккаунте OpenAI\n💡 Пополните баланс на platform.openai.com"

	case strings.Contains(errStr, "invalid_api_key"):
		return "❌ Неверный API ключ OpenAI\n💡 Проверьте OPENAI_API_KEY в настройках"

	case strings.Contains(errStr, "model_not_found"):
		return fmt.Sprintf("❌ Модель %s недоступна\n💡 Попробуйте /model gpt-4o", model)

	case strings.Contains(errStr, "rate_limit"):
		return "⏳ Превышен лимит запросов\n💡 Подождите немного и попробуйте снова"

	case strings.Contains(errStr, "timeout"):
		return "⏰ Превышено время ожидания\n💡 Попробуйте позже или используйте другую модель"

	case strings.Contains(errStr, "context deadline exceeded"):
		return "⏰ Превышено время ожидания\n💡 Попробуйте позже или используйте другую модель"

	case strings.Contains(errStr, "network"):
		return "🌐 Проблемы с сетью\n💡 Проверьте подключение к интернету"

	case strings.Contains(errStr, "unauthorized"):
		return "🔐 Ошибка авторизации\n💡 Проверьте API ключ OpenAI"

	default:
		return fmt.Sprintf("❌ Ошибка OpenAI: %s\n💡 Попробуйте позже или используйте /model gpt-4o", errStr)
	}
}

const CommandsList = `/chat <сообщение> – задать боту вопрос
/ping – проверка состояния
/start – добавить текущий чат в рассылку
/whitelist – показать список подключённых чатов
/remove <id> – убрать чат из списка
/model [имя] – показать или сменить модель (по умолчанию o3)
/lunch – немедленно запросить идеи на обед
/brief – немедленно запросить вечерний дайджест
/crypto – криптовалютный дайджест
/tech – технологический дайджест
/realestate – дайджест недвижимости
/business – бизнес-дайджест
/investment – инвестиционный дайджест
/startup – стартап-дайджест
/global – глобальный дайджест
/tasks – вывести текущее расписание задач
/task [имя] – список задач или запуск выбранной
/blockchain – метрики сети биткоина

Команды задач (автоматически создаются из tasks.yml):
/land_price, /micro_noon, /crypto_am, /gis_lots, /micro_pm, /mvp, /crypto_pm, /biz_idea, /bri_digest, /micro_night`

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

// Удалить неиспользуемый интерфейс MessageSender

var (
	CurrentModel = "o3"
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
func RegisterTaskCommands(b *tb.Bot, client *openai.Client) {
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
			resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
			if err != nil {
				logger.L.Error("openai error", "task", tcopy.Name, "model", CurrentModel, "err", err)
				return c.Send(formatOpenAIError(err, CurrentModel))
			}
			return c.Send(resp)
		})
	}
}

// scheduleDailyMessages sets up the daily lunch idea and brief messages.
func ScheduleDailyMessages(s *gocron.Scheduler, client *openai.Client, b *tb.Bot, chatID int64) {
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
			resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
			if err != nil {
				logger.L.Error("openai error", "scheduled_task", tcopy.Name, "model", CurrentModel, "err", err)
				// Для запланированных задач не отправляем сообщение пользователю, только логируем
				return
			}
			text := resp
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
func SendStartupMessage(b *tb.Bot, chatID int64) {
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

func handleTask(client *openai.Client) func(tb.Context) error {
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
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "task", t.Name, "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
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

func handleLunch(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(LunchIdeaPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "command", "lunch", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
	}
}

func handleBrief(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(DailyBriefPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "command", "brief", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
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
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logger.L.Error("failed to close response body", "err", err)
			}
		}()
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

func handleChat(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		resp, err := UserCompletion(ctx, client, q, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "command", "chat", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		_, err = c.Bot().Send(c.Sender(), resp)
		return err
	}
}

// Обработчики для новых команд дайджестов
func handleCryptoDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(CryptoDigestPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "digest", "crypto", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
	}
}

func handleTechDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(TechDigestPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "digest", "tech", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
	}
}

func handleRealEstateDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(RealEstateDigestPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "digest", "realestate", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
	}
}

func handleBusinessDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(BusinessDigestPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "digest", "business", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
	}
}

func handleInvestmentDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(InvestmentDigestPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "digest", "investment", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
	}
}

func handleStartupDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(StartupDigestPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "digest", "startup", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
	}
}

func handleGlobalDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		prompt := applyTemplate(GlobalDigestPrompt)
		resp, err := SystemCompletion(ctx, client, prompt, CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "digest", "global", "model", CurrentModel, "err", err)
			return c.Send(formatOpenAIError(err, CurrentModel))
		}
		return c.Send(resp)
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
	b.Handle("/crypto", handleCryptoDigest(client))
	b.Handle("/tech", handleTechDigest(client))
	b.Handle("/realestate", handleRealEstateDigest(client))
	b.Handle("/business", handleBusinessDigest(client))
	b.Handle("/investment", handleInvestmentDigest(client))
	b.Handle("/startup", handleStartupDigest(client))
	b.Handle("/global", handleGlobalDigest(client))
	b.Handle("/blockchain", handleBlockchain(cfg.BlockchainAPI))
	b.Handle("/chat", handleChat(client))

	b.Start()
	return nil
}
