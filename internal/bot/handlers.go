package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"telegram-reminder/internal/config"
	"telegram-reminder/internal/logger"

	tb "gopkg.in/telebot.v3"
)

// RegisterHandlers регистрирует все основные команды Telegram-бота (ping, start, whitelist, remove, tasks, task, model, lunch, brief, blockchain, chat).
// Использует переданные *tb.Bot, ChatCompleter и config.Config для доступа к API и настройкам.
func RegisterHandlers(b *tb.Bot, client ChatCompleter, cfg config.Config) {
	// /ping — проверка состояния
	b.Handle("/ping", func(c tb.Context) error {
		return c.Send("pong")
	})

	// /start — добавить чат в whitelist
	b.Handle("/start", func(c tb.Context) error {
		if err := AddIDToWhitelist(c.Chat().ID); err != nil {
			logger.L.Error("whitelist add", "err", err)
		}
		return c.Send("Бот активирован")
	})

	// /whitelist — показать список чатов
	b.Handle("/whitelist", func(c tb.Context) error {
		ids, err := LoadWhitelist()
		if err != nil {
			return handleLoadError(c, "whitelist", err)
		}
		if len(ids) == 0 {
			return handleNotFound(c, "Whitelist")
		}
		return c.Send(FormatWhitelist(ids))
	})

	// /remove — удалить чат из whitelist
	b.Handle("/remove", func(c tb.Context) error {
		payload := strings.TrimSpace(c.Message().Payload)
		if payload == "" {
			return sendError(c, "Пожалуйста, укажите ID для удаления: /remove <id>")
		}
		id, err := strconv.ParseInt(payload, 10, 64)
		if err != nil {
			return sendError(c, "Некорректный ID. Пример: /remove 123456789")
		}
		if err := RemoveIDFromWhitelist(id); err != nil {
			return handleLoadError(c, "remove", err)
		}
		logger.L.Info("ID успешно удалён из whitelist", "id", id)
		return c.Send("ID успешно удалён из whitelist.")
	})

	// /tasks — показать расписание задач
	b.Handle("/tasks", func(c tb.Context) error {
		tasks, err := LoadTasks()
		if err != nil {
			return handleLoadError(c, "tasks", err)
		}
		return c.Send(FormatTasks(tasks))
	})

	// /task — показать или выполнить задачу по имени
	b.Handle("/task", func(c tb.Context) error {
		name := strings.TrimSpace(c.Message().Payload)
		tasks, err := LoadTasks()
		if err != nil {
			return handleLoadError(c, "tasks", err)
		}
		if name == "" {
			return c.Send(FormatTaskNames(tasks))
		}
		t, ok := FindTask(tasks, name)
		if !ok {
			return handleNotFound(c, "task")
		}
		ctx, cancel := newTimeoutCtx(OpenAITimeout)
		defer cancel()

		prompt := t.Prompt
		text, err := SystemCompletion(ctx, client, prompt)
		if err != nil {
			return handleOpenAIError(c, err)
		}
		return c.Send(text)
	})

	// /model — показать или сменить модель OpenAI
	b.Handle("/model", func(c tb.Context) error {
		payload := strings.TrimSpace(c.Message().Payload)
		if payload == "" {
			return c.Send("Current model: (not supported)\nSupported: (not supported)")
		}
		return c.Send(fmt.Sprintf("Model set to %s", payload))
	})

	// /lunch — немедленно запросить идею на обед
	b.Handle("/lunch", func(c tb.Context) error {
		ctx, cancel := newTimeoutCtx(OpenAITimeout)
		defer cancel()

		text, err := SystemCompletion(ctx, client, LunchIdeaPrompt)
		if err != nil {
			return handleOpenAIError(c, err)
		}
		return c.Send(text)
	})

	// /brief — немедленно запросить вечерний дайджест
	b.Handle("/brief", func(c tb.Context) error {
		ctx, cancel := newTimeoutCtx(OpenAITimeout)
		defer cancel()

		text, err := SystemCompletion(ctx, client, DailyBriefPrompt)
		if err != nil {
			return handleOpenAIError(c, err)
		}
		return c.Send(text)
	})

	// /blockchain — получить метрики сети биткоина
	b.Handle("/blockchain", func(c tb.Context) error {
		ctx, cancel := newTimeoutCtx(BlockchainTimeout)
		defer cancel()

		apiURL := cfg.BlockchainAPI
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
	})

	// /chat — задать вопрос и получить ответ от OpenAI
	b.Handle("/chat", func(c tb.Context) error {
		q := strings.TrimSpace(c.Message().Payload)
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := newTimeoutCtx(OpenAITimeout)
		defer cancel()

		text, err := UserCompletion(ctx, client, q)
		if err != nil {
			return handleOpenAIError(c, err)
		}
		_, err = c.Bot().Send(c.Sender(), text)
		return err
	})
}

// newTimeoutCtx создает context с таймаутом.
func newTimeoutCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// sendError отправляет стандартное сообщение об ошибке.
func sendError(c tb.Context, msg string) error {
	return c.Send("❗ " + msg)
}

// handleOpenAIError логирует и отправляет ошибку OpenAI.
func handleOpenAIError(c tb.Context, err error) error {
	logger.L.Error("Ошибка OpenAI", "err", err)
	return sendError(c, "Ошибка при обращении к OpenAI. Попробуйте позже или проверьте настройки API-ключа.")
}

// handleLoadError логирует и отправляет ошибку загрузки данных (whitelist, tasks и др.).
func handleLoadError(c tb.Context, what string, err error) error {
	logger.L.Error("Ошибка загрузки "+what, "err", err)
	return sendError(c, "Ошибка при загрузке "+what+": "+err.Error())
}

// handleNotFound отправляет стандартное сообщение о ненайденном ресурсе.
func handleNotFound(c tb.Context, what string) error {
	return sendError(c, what+" не найден(а). Проверьте правильность ввода.")
}
