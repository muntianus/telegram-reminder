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

// EnhancedSystemCompletion combines web search results with OpenAI completions
func EnhancedSystemCompletion(ctx context.Context, client *openai.Client, prompt string, model string) (string, error) {
	// Просто используем обычный SystemCompletion без веб-поиска
	return SystemCompletion(ctx, client, prompt, model)
}

// Prompt templates
const (
	DailyBriefPrompt = `
Ты — Telegram-бот для ежедневного дайджеста. Говоришь кратко, дерзко, панибратски.

📅 ВАЖНО: Используй веб-поиск для получения актуальных новостей.

Заполни блоки:
⚡ Микродействие (одно простое действие на сегодня)
🧠 Тема дня (мини‑инсайт/мысль на основе сегодняшних событий)
💰 Что залутать (актив/идея на основе сегодняшних трендов)
🏞️ Земля на присмотр (лоты в южном Подмосковье: Бутово, Щербинка, Подольск, Воскресенск)
🪙 Альт дня (актуальная монета на основе сегодняшних движений, линк CoinGecko)
🚀 Пушка с ProductHunt (сегодняшние топовые проекты)

🔍 ВЕБ-ПОИСК: Найди актуальные новости по темам:
- Криптовалюты и DeFi
- Технологии и стартапы
- Недвижимость и инвестиции
- Бизнес-тренды

ВАЖНО: Используй веб-поиск для получения актуальных новостей и включай ссылки на источники.

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

var baseCommands = []string{
	"/chat <сообщение> – задать боту вопрос",
	"/ping – проверка состояния",
	"/start – добавить текущий чат в рассылку",
	"/whitelist – показать список подключённых чатов",
	"/remove <id> – убрать чат из списка",
	"/model [имя] – показать или сменить модель (по умолчанию gpt-4o)",
	"/lunch – немедленно запросить идеи на обед",
	"/brief – немедленно запросить вечерний дайджест",
	"/crypto – криптовалютный дайджест",
	"/tech – технологический дайджест",
	"/realestate – дайджест недвижимости",
	"/business – бизнес-дайджест",
	"/investment – инвестиционный дайджест",
	"/startup – стартап-дайджест",
	"/global – глобальный дайджест",
	"/tasks – вывести текущее расписание задач",
	"/task [имя] – список задач или запуск выбранной",
	"/blockchain – метрики сети биткоина",
}

func buildCommandsList(tasks []Task) string {
	var sb strings.Builder
	for _, cmd := range baseCommands {
		sb.WriteString(cmd)
		sb.WriteByte('\n')
	}
	if len(tasks) > 0 {
		sb.WriteString("\nКоманды задач (автоматически создаются из tasks.yml):\n")
		names := []string{}
		for _, t := range tasks {
			if t.Name != "" {
				names = append(names, "/"+t.Name)
			}
		}
		sb.WriteString(strings.Join(names, ", "))
	}
	return sb.String()
}

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
	Model  string `json:"model,omitempty" yaml:"model,omitempty"`
}

var (
	CurrentModel = "gpt-4o" // Модель по умолчанию с веб-поиском
	ModelMu      sync.RWMutex
	BasePrompt   string
	// SupportedModels contains all OpenAI model identifiers that support web search and tools
	SupportedModels = []string{
		// Models with web search and tools support
		"gpt-4o",
		"gpt-4o-2024-05-13",
		"gpt-4o-2024-08-06",
		"gpt-4o-2024-11-20",
		"chatgpt-4o-latest",
		"gpt-4o-mini",
		"gpt-4o-mini-2024-07-18",
		"gpt-4-turbo",
		"gpt-4-turbo-2024-04-09",
		"gpt-4-0125-preview",
		"gpt-4-1106-preview",
		"gpt-4-turbo-preview",
		"gpt-4-vision-preview",
		"gpt-4",
		"gpt-4.1",
		"gpt-4.1-2025-04-14",
		"gpt-4.1-mini",
		"gpt-4.1-mini-2025-04-14",
		"gpt-4.1-nano",
		"gpt-4.1-nano-2025-04-14",
		"gpt-4.5-preview",
		"gpt-4.5-preview-2025-02-27",
		"o1-mini",
		"o1-mini-2024-09-12",
		"o1-preview",
		"o1-preview-2024-09-12",
		"o1",
		"o1-2024-12-17",
		"o3",
		"o3-2025-04-16",
		"o3-mini",
		"o3-mini-2025-01-31",
		"o4-mini",
		"o4-mini-2025-04-16",
	}
)

// IsSupportedModel returns true if the given model is in the SupportedModels list.
func IsSupportedModel(m string) bool {
	for _, sm := range SupportedModels {
		if sm == m {
			return true
		}
	}
	return false
}

var (
	LoadedTasks []Task
	TasksMu     sync.RWMutex
)

// applyTemplate replaces placeholders in the prompt with runtime values.
func applyTemplate(prompt, model string) string {
	vars := map[string]string{
		"base_prompt":  BasePrompt,
		"date":         time.Now().Format("2006-01-02"),
		"exchange_api": os.Getenv("EXCHANGE_API"),
		"chart_path":   os.Getenv("CHART_PATH"),
		"model":        model,
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
			model := CurrentModel
			if tcopy.Model != "" {
				model = tcopy.Model
			}
			prompt := applyTemplate(tcopy.Prompt, model)
			resp, err := SystemCompletion(ctx, client, prompt, model)
			if err != nil {
				logger.L.Error("openai error", "task", tcopy.Name, "model", model, "err", err)
				return c.Send(formatOpenAIError(err, model))
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

			model := CurrentModel
			if tcopy.Model != "" {
				model = tcopy.Model
			}

			log.Printf("running task: %s", tcopy.Name)
			prompt := applyTemplate(tcopy.Prompt, model)
			resp, err := SystemCompletion(ctx, client, prompt, model)
			if err != nil {
				logger.L.Error("openai error", "scheduled_task", tcopy.Name, "model", model, "err", err)
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
func SendStartupMessage(b *tb.Bot, chatID int64, msg string) {
	if chatID != 0 {
		if _, err := b.Send(tb.ChatID(chatID), msg); err != nil {
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
		if _, err := b.Send(tb.ChatID(id), msg); err != nil {
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
		model := CurrentModel
		if t.Model != "" {
			model = t.Model
		}
		prompt := applyTemplate(t.Prompt, model)
		resp, err := SystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "task", t.Name, "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
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
		if !IsSupportedModel(payload) {
			return c.Send("unsupported model")
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
		model := CurrentModel
		prompt := applyTemplate(LunchIdeaPrompt, model)
		resp, err := SystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "command", "lunch", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

func handleBrief(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := CurrentModel
		prompt := applyTemplate(DailyBriefPrompt, model)
		resp, err := SystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "command", "brief", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
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
		model := CurrentModel
		prompt := applyTemplate(CryptoDigestPrompt, model)
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "digest", "crypto", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

func handleTechDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := CurrentModel
		prompt := applyTemplate(TechDigestPrompt, model)
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "digest", "tech", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

func handleRealEstateDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := CurrentModel
		prompt := applyTemplate(RealEstateDigestPrompt, model)
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "digest", "realestate", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

func handleBusinessDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := CurrentModel
		prompt := applyTemplate(BusinessDigestPrompt, model)
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "digest", "business", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

func handleInvestmentDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := CurrentModel
		prompt := applyTemplate(InvestmentDigestPrompt, model)
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "digest", "investment", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

func handleStartupDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := CurrentModel
		prompt := applyTemplate(StartupDigestPrompt, model)
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "digest", "startup", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

func handleGlobalDigest(client *openai.Client) func(tb.Context) error {
	return func(c tb.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := CurrentModel
		prompt := applyTemplate(GlobalDigestPrompt, model)
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "digest", "global", "model", model, "err", err)
			return c.Send(formatOpenAIError(err, model))
		}
		return c.Send(resp)
	}
}

// Run initializes and starts the Telegram bot.
func Run(cfg config.Config) error {
	b, err := New(cfg)
	if err != nil {
		return err
	}
	return b.Start()
}
