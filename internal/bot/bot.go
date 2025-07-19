package bot

import (
	"context"
	"fmt"
	"net/http"
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

// RegisterTaskCommands creates bot handlers for all named tasks.
// RegisterTaskCommands регистрирует обработчики команд для всех задач с именем.
// Если LoadedTasks не пуст, использует его (для тестов), иначе загружает задачи из конфигурации.
func RegisterTaskCommands(b *tb.Bot, client ChatCompleter) {
	if len(LoadedTasks) == 0 {
		var err error
		LoadedTasks, err = LoadTasks()
		if err != nil {
			logger.L.Error("load tasks", "err", err)
			return
		}
	}
	for _, t := range LoadedTasks {
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

// ScheduleDailyMessages sets up the daily lunch idea and brief messages.
// ScheduleDailyMessages планирует ежедневные задачи (дайджесты, обеды и др.) через gocron.
// Загружает задачи из конфигурации и регистрирует их в планировщике.
func ScheduleDailyMessages(s *gocron.Scheduler, client ChatCompleter, b *tb.Bot, chatID int64) {
	tasks, err := LoadTasks()
	if err != nil {
		logger.L.Error("load tasks", "err", err)
		return
	}
	for _, t := range tasks {
		tcopy := t
		job := func() {
			ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
			defer cancel()

			logger.L.Info("running task", "task", tcopy.Name)
			prompt := applyTemplate(tcopy.Prompt)
			text, err := SystemCompletion(ctx, client, prompt)
			if err != nil {
				logger.L.Error("openai error", "err", err)
				return
			}
			if chatID != 0 {
				if _, err := b.Send(tb.ChatID(chatID), text); err != nil {
					logger.L.Error("telegram send", "err", err)
				} else {
					logger.L.Debug("sent", "chat_id", chatID)
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
					logger.L.Error("telegram send", "err", err, "chat_id", id)
				} else {
					logger.L.Debug("sent", "chat_id", id)
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
// SendStartupMessage отправляет стартовое сообщение в указанный чат или всем из whitelist.
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

// Run initializes and starts the Telegram bot.
// Run — точка входа: инициализирует бота, планировщик, обработчики команд и запускает цикл обработки.
func Run(cfg config.Config) error {
	if cfg.OpenAIModel != "" {
		// This line is no longer needed as OpenAI client is now directly used.
		// openai.CurrentModel = cfg.OpenAIModel
	}

	b, err := tb.NewBot(tb.Settings{Token: cfg.TelegramToken})
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}
	logger.L.Info("authorized", "username", b.Me.Username)

	if cfg.ChatID != 0 {
		if err := AddIDToWhitelist(cfg.ChatID); err != nil {
			logger.L.Error("whitelist add", "err", err)
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

	logger.L.Info("scheduler started")
	scheduler.StartAsync()

	SendStartupMessage(b, cfg.ChatID)

	// В функции Run:
	// RegisterHandlers(b, client, cfg)

	b.Start()
	return nil
}
