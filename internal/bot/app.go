package bot

import (
	"fmt"
	"log/slog"
	"time"

	"telegram-reminder/internal/config"
	"telegram-reminder/internal/logger"

	"github.com/go-co-op/gocron"
	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

// Bot encapsulates dependencies of the Telegram bot.
type Bot struct {
	Config    config.Config
	TeleBot   *tb.Bot
	Client    *openai.Client
	Scheduler *gocron.Scheduler
}

// New creates a new Bot instance with initialized dependencies.
func New(cfg config.Config) (*Bot, error) {
	updateRuntimeConfig(func(rc *RuntimeConfig) {
		if cfg.OpenAIModel != "" {
			rc.CurrentModel = cfg.OpenAIModel
		}
		if cfg.OpenAIMaxTokens > 0 {
			rc.MaxTokens = cfg.OpenAIMaxTokens
		}
		rc.EnableWebSearch = cfg.EnableWebSearch
		rc.ToolChoice = cfg.OpenAIToolChoice
		rc.ServiceTier = openai.ServiceTier(cfg.OpenAIServiceTier)
		rc.ReasoningEffort = cfg.OpenAIReasoningEffort
	})

	tele, err := tb.NewBot(tb.Settings{Token: cfg.TelegramToken})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	oaCfg := openai.DefaultConfig(cfg.OpenAIKey)
	oaCfg.HTTPClient = logger.NewHTTPClient(OpenAITimeout)
	client := openai.NewClientWithConfig(oaCfg)

	tz, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	sched := gocron.NewScheduler(tz)

	b := &Bot{
		Config:    cfg,
		TeleBot:   tele,
		Client:    client,
		Scheduler: sched,
	}
	return b, nil
}

// Start registers handlers, schedules tasks and starts the bot.
func (b *Bot) Start() error {
	logger.L.Info("authorized", "user", b.TeleBot.Me.Username)

	if b.Config.LogChatID != 0 {
		logger.EnableTelegramLogging(b.Config.TelegramToken, b.Config.LogChatID, slog.LevelError)
	}

	b.TeleBot.Use(logger.TelebotMiddleware())

	if b.Config.ChatID != 0 {
		if err := AddIDToWhitelist(b.Config.ChatID); err != nil {
			logger.L.Error("whitelist add", "err", err)
		}
	}

	ScheduleDailyMessages(b.Scheduler, b.Client, b.TeleBot, b.Config.ChatID)
	RegisterTaskCommands(b.TeleBot, b.Client)

	b.Scheduler.StartAsync()

	TasksMu.RLock()
	cmds := buildCommandsList(LoadedTasks)
	TasksMu.RUnlock()
	msg := fmt.Sprintf("Billion Roadmap %s\n\n%s", Version, cmds)
	SendStartupMessage(b.TeleBot, b.Config.ChatID, msg)

	b.TeleBot.Handle("/ping", handlePing)
	b.TeleBot.Handle("/start", handleStart)
	b.TeleBot.Handle("/whitelist", handleWhitelist)
	b.TeleBot.Handle("/remove", handleRemove)
	b.TeleBot.Handle("/tasks", handleTasks)
	b.TeleBot.Handle("/task", handleTask(b.Client))
	b.TeleBot.Handle("/model", handleModel())
	b.TeleBot.Handle("/lunch", handleLunch(b.Client))
	b.TeleBot.Handle("/brief", handleBrief(b.Client))
	// Initialize new digest architecture
	digestIntegration, err := NewDigestIntegration(b.Client, DefaultErrorHandler)
	if err != nil {
		logger.L.Error("failed to initialize digest integration", "err", err)
	} else {
		// Replace old digest handlers with new architecture
		digestIntegration.ReplaceDigestHandlers(b.TeleBot, b.Client)
		logger.L.Info("digest handlers replaced with new architecture")
	}
	b.TeleBot.Handle("/blockchain", handleBlockchain(b.Config.BlockchainAPI))
	b.TeleBot.Handle("/chat", handleChat(b.Client))
	b.TeleBot.Handle("/search", handleSearch())
	b.TeleBot.Handle("/webdoc", handleWebDoc())

	b.TeleBot.Start()
	return nil
}
