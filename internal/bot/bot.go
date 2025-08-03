package bot

import (
	"context"
	"fmt"
	"html"
	"os"
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
func EnhancedSystemCompletion(ctx context.Context, client ChatCompleter, prompt string, model string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}}
	return ChatCompletion(ctx, client, msgs, model)
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

// OpenAITimeout defines how long the bot waits for a response from OpenAI.
// The previous value of 3 minutes was occasionally insufficient for
// complex prompts. Increasing the timeout helps prevent premature
// cancellation of requests.
const OpenAITimeout = 5 * time.Minute
const BlockchainTimeout = 10 * time.Second

const TelegramMessageLimit = 4096

const Version = "0.1.0"

// RuntimeConfig holds runtime configuration for the bot
type RuntimeConfig struct {
	CurrentModel    string
	MaxTokens       int
	ServiceTier     openai.ServiceTier
	ReasoningEffort string
	EnableWebSearch bool
	ToolChoice      string
	BasePrompt      string
}

var runtimeConfig = RuntimeConfig{
	CurrentModel:    "gpt-4.1",
	MaxTokens:       600,
	EnableWebSearch: true,
	ToolChoice:      "auto",
}

// getRuntimeConfig returns a deep copy of the current runtime configuration
func getRuntimeConfig() RuntimeConfig {
	ModelMu.RLock()
	defer ModelMu.RUnlock()
	// Create a deep copy to prevent race conditions
	return RuntimeConfig{
		CurrentModel:    runtimeConfig.CurrentModel,
		MaxTokens:       runtimeConfig.MaxTokens,
		ServiceTier:     runtimeConfig.ServiceTier,
		ReasoningEffort: runtimeConfig.ReasoningEffort,
		EnableWebSearch: runtimeConfig.EnableWebSearch,
		ToolChoice:      runtimeConfig.ToolChoice,
		BasePrompt:      runtimeConfig.BasePrompt,
	}
}

// updateRuntimeConfig updates runtime configuration safely
func updateRuntimeConfig(updateFunc func(*RuntimeConfig)) {
	ModelMu.Lock()
	defer ModelMu.Unlock()
	updateFunc(&runtimeConfig)
}

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
		return fmt.Sprintf("❌ Модель %s недоступна\n💡 Попробуйте /model gpt-4.1", model)

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
		return fmt.Sprintf("❌ Ошибка OpenAI: %s\n💡 Попробуйте позже или используйте /model gpt-4.1", errStr)
	}
}

var baseCommands = []string{
	"/chat <сообщение> – задать боту вопрос",
	"/search <запрос> – выполнить поиск через OpenAI",
	"/ping – проверка состояния",
	"/start – добавить текущий чат в рассылку",
	"/whitelist – показать список подключённых чатов",
	"/remove <id> – убрать чат из списка",
	"/model [имя] – показать или сменить модель (по умолчанию gpt-4.1)",
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
		sb.WriteString(html.EscapeString(cmd))
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
	ModelMu sync.RWMutex
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

var (
	LoadedTasks []Task
	TasksMu     sync.RWMutex
)

// applyTemplate replaces placeholders in the prompt with runtime values.
func applyTemplate(prompt, model string) string {
	vars := map[string]string{
		"base_prompt":  runtimeConfig.BasePrompt,
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
			ModelMu.RLock()
			model := runtimeConfig.CurrentModel
			ModelMu.RUnlock()
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

// createTaskJob creates a job function for a scheduled task
func createTaskJob(task Task, client ChatCompleter, b *tb.Bot, chatID int64) func() {
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()

		model := getRuntimeConfig().CurrentModel
		if task.Model != "" {
			model = task.Model
		}

		taskLogger := logger.GetTaskLogger()
		openaiLogger := logger.GetOpenAILogger()
		
		op := taskLogger.Operation("task_execution")
		op.WithContext("task_name", task.Name)
		op.WithContext("model", model)
		op.WithContext("chat_id", chatID)
		
		op.Step("preparing_prompt")
		prompt := applyTemplate(task.Prompt, model)
		
		op.Step("calling_openai")
		startTime := time.Now()
		resp, err := SystemCompletion(ctx, client, prompt, model)
		duration := time.Since(startTime)
		
		openaiLogger.APICall("openai", "system_completion", err == nil, duration, err)
		
		if err != nil {
			op.Failure("Task execution failed", err)
			DefaultErrorHandler.HandleTaskError(err, task.Name, model)
			return
		}
		
		op.WithContext("response_length", len(resp))
		op.Step("broadcasting_result")
		
		taskLogger.TaskExecution(task.Name, true, time.Since(startTime), nil)
		op.Success("Task completed successfully")
		
		broadcastTaskResult(b, chatID, resp)
	}
}

// broadcastTaskResult sends task result to specified chat or all whitelisted chats
func broadcastTaskResult(b *tb.Bot, chatID int64, text string) {
	if chatID != 0 {
		if err := sendLong(b, tb.ChatID(chatID), text); err != nil {
			DefaultErrorHandler.HandleTelegramError(err, chatID)
		}
		return
	}

	ids, err := LoadWhitelist()
	if err != nil {
		logger.L.Error("load whitelist", "err", err)
		return
	}

	logger.L.Debug("broadcast task result", "recipients", len(ids))
	for _, id := range ids {
		if err := sendLong(b, tb.ChatID(id), text); err != nil {
			DefaultErrorHandler.HandleTelegramError(err, id)
		}
	}
}

// scheduleTask schedules a single task in the scheduler
func scheduleTask(s *gocron.Scheduler, task Task, job func()) error {
	var j *gocron.Job
	var err error

	switch {
	case task.Cron != "":
		logger.L.Debug("schedule cron", "name", task.Name, "cron", task.Cron)
		j, err = s.Cron(task.Cron).Do(job)
	default:
		timeStr := task.Time
		if timeStr == "" {
			timeStr = "00:00"
		}
		logger.L.Debug("schedule daily", "name", task.Name, "schedule_time", timeStr)
		j, err = s.Every(1).Day().At(timeStr).Do(job)
	}

	if err != nil {
		return err
	}

	// Register event listeners for monitoring (debug logging removed to prevent spam)
	j.RegisterEventListeners(
		gocron.WhenJobReturnsError(func(jobName string, err error) { logger.L.Error("job error", "job", task.Name, "err", err) }),
	)
	j.Tag(task.Name)

	return nil
}

// ScheduleDailyMessages sets up the daily lunch idea and brief messages.
func ScheduleDailyMessages(s *gocron.Scheduler, client ChatCompleter, b *tb.Bot, chatID int64) {
	tasks, err := LoadTasks()
	if err != nil {
		logger.L.Error("load tasks", "err", err)
		return
	}

	logger.L.Debug("loaded tasks", "count", len(tasks))

	TasksMu.Lock()
	LoadedTasks = tasks
	TasksMu.Unlock()

	for _, task := range tasks {
		job := createTaskJob(task, client, b, chatID)
		if err := scheduleTask(s, task, job); err != nil {
			logger.L.Error("schedule job", "task", task.Name, "err", err)
		}
	}
}

// SendStartupMessage notifies the chat that the bot is running.
func SendStartupMessage(b *tb.Bot, chatID int64, msg string) {
	logger.L.Debug("send startup message", "chat_id", chatID)
	if chatID != 0 {
		if err := sendLong(b, tb.ChatID(chatID), msg); err != nil {
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
		if err := sendLong(b, tb.ChatID(id), msg); err != nil {
			logger.L.Error("telegram send", "err", err)
		}
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
