package bot

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
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
	if cfg.OpenAIModel != "" {
		CurrentModel = cfg.OpenAIModel
	}
	EnableWebSearch = cfg.EnableWebSearch
	SearchProviderURL = cfg.SearchProviderURL

	tele, err := tb.NewBot(tb.Settings{Token: cfg.TelegramToken})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	oaCfg := openai.DefaultConfig(cfg.OpenAIKey)
	oaCfg.HTTPClient = &http.Client{Timeout: OpenAITimeout}
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
	log.Printf("Authorized as %s", b.TeleBot.Me.Username)

	if b.Config.LogChatID != 0 {
		logger.EnableTelegramLogging(b.Config.TelegramToken, b.Config.LogChatID, slog.LevelDebug)
	}

	if b.Config.ChatID != 0 {
		if err := AddIDToWhitelist(b.Config.ChatID); err != nil {
			log.Printf("whitelist add: %v", err)
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
	b.TeleBot.Handle("/crypto", handleCryptoDigest(b.Client))
	b.TeleBot.Handle("/tech", handleTechDigest(b.Client))
	b.TeleBot.Handle("/realestate", handleRealEstateDigest(b.Client))
	b.TeleBot.Handle("/business", handleBusinessDigest(b.Client))
	b.TeleBot.Handle("/investment", handleInvestmentDigest(b.Client))
	b.TeleBot.Handle("/startup", handleStartupDigest(b.Client))
	b.TeleBot.Handle("/global", handleGlobalDigest(b.Client))
	b.TeleBot.Handle("/blockchain", handleBlockchain(b.Config.BlockchainAPI))
	b.TeleBot.Handle("/chat", handleChat(b.Client))

	b.TeleBot.Start()
	return nil
}
