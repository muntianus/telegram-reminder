package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"telegram-reminder/internal/logger"

	tb "gopkg.in/telebot.v3"
)

// --- HANDLER FUNCTIONS ---

func handlePing(c tb.Context) error {
	logger.L.Debug("command ping", "chat", c.Chat().ID)
	return c.Send("pong")
}

func handleStart(c tb.Context) error {
	logger.L.Debug("command start", "chat", c.Chat().ID)
	if err := AddIDToWhitelist(c.Chat().ID); err != nil {
		logger.L.Error("whitelist add", "err", err)
	}
	return c.Send("Бот активирован")
}

func handleWhitelist(c tb.Context) error {
	logger.L.Debug("command whitelist", "chat", c.Chat().ID)
	ids, err := LoadWhitelist()
	if err != nil {
		logger.L.Error("load whitelist", "err", err)
		return c.Send("whitelist error")
	}
	if len(ids) == 0 {
		return c.Send("Whitelist is empty")
	}
	return c.Send(FormatWhitelist(ids))
}

func handleRemove(c tb.Context) error {
	logger.L.Debug("command remove", "chat", c.Chat().ID)
	payload := sanitizeInput(c.Message().Payload)
	if err := validatePayload(payload); err != nil {
		logger.L.Debug("invalid payload", "err", err)
		return c.Send("Usage: /remove <id>")
	}
	if payload == "" {
		return c.Send("Usage: /remove <id>")
	}
	id, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return c.Send("Bad ID")
	}
	if err := RemoveIDFromWhitelist(id); err != nil {
		logger.L.Error("remove id", "err", err)
		return c.Send("remove error")
	}
	return c.Send("Removed")
}

func handleTasks(c tb.Context) error {
	logger.L.Debug("command tasks", "chat", c.Chat().ID)
	TasksMu.RLock()
	tasks := append([]Task(nil), LoadedTasks...)
	TasksMu.RUnlock()
	return c.Send(FormatTasks(tasks))
}

func handleTask(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command task", "chat", c.Chat().ID, "payload", c.Message().Payload)
		name := sanitizeInput(c.Message().Payload)
		if err := validatePayload(name); err != nil {
			logger.L.Debug("invalid task name", "err", err)
			return c.Send("Task name invalid")
		}
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
		model := getRuntimeConfig().CurrentModel
		if t.Model != "" {
			model = t.Model
		}
		prompt := applyTemplate(t.Prompt, model)
		resp, err := SystemCompletion(ctx, client, prompt, model)
		if err != nil {
			return c.Send(DefaultErrorHandler.HandleOpenAIError(err, model))
		}
		return replyLong(c, resp)
	}
}

func handleModel() func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command model", "chat", c.Chat().ID, "payload", c.Message().Payload)
		payload := sanitizeInput(c.Message().Payload)
		if err := validatePayload(payload); err != nil {
			logger.L.Debug("invalid model payload", "err", err)
			return c.Send("Invalid model name")
		}
		if payload == "" {
			cur := getRuntimeConfig().CurrentModel
			return c.Send(fmt.Sprintf(
				"Current model: %s\nSupported: %s",
				cur, strings.Join(SupportedModels, ", "),
			))
		}
		valid := false
		for _, m := range SupportedModels {
			if payload == m {
				valid = true
				break
			}
		}
		if !valid {
			return c.Send(fmt.Sprintf("Unsupported model: %s", payload))
		}
		updateRuntimeConfig(func(cfg *RuntimeConfig) {
			cfg.CurrentModel = payload
		})
		return c.Send(fmt.Sprintf("Model set to %s", payload))
	}
}

func handleLunch(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command lunch", "chat", c.Chat().ID)
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := getRuntimeConfig().CurrentModel
		prompt := applyTemplate(LunchIdeaPrompt, model)
		resp, err := SystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "command", "lunch", "model", model, "err", err)
			return c.Send(DefaultErrorHandler.HandleOpenAIError(err, model))
		}
		return replyLong(c, resp)
	}
}

func handleBrief(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command brief", "chat", c.Chat().ID)
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		model := getRuntimeConfig().CurrentModel
		prompt := applyTemplate(DailyBriefPrompt, model)
		resp, err := SystemCompletion(ctx, client, prompt, model)
		if err != nil {
			logger.L.Error("openai error", "command", "brief", "model", model, "err", err)
			return c.Send(DefaultErrorHandler.HandleOpenAIError(err, model))
		}
		return replyLong(c, resp)
	}
}

func handleBlockchain(apiURL string) func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command blockchain", "chat", c.Chat().ID)
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
		return replyLong(c, msg)
	}
}

func handleChat(client ChatCompleter) func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command chat", "chat", c.Chat().ID)
		q := sanitizeInput(c.Message().Payload)
		if err := validateChatMessage(q); err != nil {
			logger.L.Debug("invalid chat message", "err", err)
			return c.Send("Message too long or invalid")
		}
		if q == "" {
			return c.Send("Usage: /chat <message>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()
		resp, err := UserCompletion(ctx, client, q, getRuntimeConfig().CurrentModel)
		if err != nil {
			logger.L.Error("openai error", "command", "chat", "model", getRuntimeConfig().CurrentModel, "err", err)
			return c.Send(DefaultErrorHandler.HandleOpenAIError(err, getRuntimeConfig().CurrentModel))
		}
		return sendLong(c.Bot(), c.Sender(), resp)
	}
}

func handleSearch() func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command search", "chat", c.Chat().ID)
		q := sanitizeInput(c.Message().Payload)
		if err := validateQuery(q); err != nil {
			logger.L.Debug("invalid search query", "err", err)
			return c.Send("Search query too long, too short, or invalid")
		}
		result, err := OpenAISearch(q)
		if err != nil {
			logger.L.Error("openai search", "err", err)
			return c.Send("search error")
		}
		if strings.TrimSpace(result) == "" {
			return c.Send("no results")
		}
		return replyLong(c, result)
	}
}

// Old digest handlers removed - replaced with new architecture in digest_integration.go

// handleWebDoc sends the web search documentation snippet to the user.
func handleWebDoc() func(tb.Context) error {
	return func(c tb.Context) error {
		logger.L.Debug("command webdoc", "chat", c.Chat().ID)
		webSearchDoc := `🔍 **Веб-поиск через OpenAI**

Команда /search позволяет выполнять поиск в интернете через OpenAI API.

**Использование:**
/search <запрос>

**Пример:**
/search последние новости ИИ

**Особенности:**
- Использует возможности веб-поиска OpenAI
- Возвращает актуальную информацию
- Форматирует результаты для удобного чтения

**Примечание:** Доступность зависит от модели OpenAI.`
		return replyLong(c, webSearchDoc)
	}
}
