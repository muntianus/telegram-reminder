package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"telegram-reminder/internal/logger"

	tb "gopkg.in/telebot.v3"
)

// --- HANDLER FUNCTIONS ---

func handlePing(c tb.Context) error {
	handlerLogger := logger.GetHandlerLogger()
	handlerLogger.UserAction(c.Chat().ID, "ping", nil)
	return c.Send("pong")
}

func handleStart(c tb.Context) error {
	handlerLogger := logger.GetHandlerLogger()
	securityLogger := logger.GetSecurityLogger()

	op := handlerLogger.Operation("user_start")
	op.WithContext("user_id", c.Chat().ID)

	handlerLogger.UserAction(c.Chat().ID, "start", nil)

	// Use enhanced chat management for better group support
	if err := AddChatToWhitelist(c.Chat()); err != nil {
		securityLogger.SecurityEvent("whitelist_add_failed", c.Chat().ID, map[string]interface{}{
			"error": err.Error(),
		})
		op.Failure("Failed to add user to whitelist", err)
		return c.Send("Ошибка активации")
	}

	securityLogger.SecurityEvent("user_activated", c.Chat().ID, map[string]interface{}{
		"action": "whitelist_added",
	})
	op.Success("User successfully activated")
	// Provide different messages based on chat type
	chatType := getChatTypeString(c.Chat().Type)
	switch chatType {
	case "group", "supergroup":
		return c.Send(fmt.Sprintf("🎉 Бот активирован для группы \"%s\"!\n📢 Теперь все участники будут получать дайджесты", getChatTitle(c.Chat())))
	case "private":
		return c.Send("🤖 Бот активирован! Вы будете получать ежедневные дайджесты")
	default:
		return c.Send("✅ Бот активирован")
	}
}

func handleWhitelist(c tb.Context) error {
	logger.L.Debug("command whitelist", "chat", c.Chat().ID)

	// Use enhanced chat formatting
	chatList := FormatChatList()
	if strings.Contains(chatList, "пуст") {
		return c.Send("📭 Список активных чатов пуст")
	}

	return c.Send(chatList, &tb.SendOptions{ParseMode: tb.ModeHTML})
}

func handleGroups(c tb.Context) error {
	logger.L.Debug("command groups", "chat", c.Chat().ID)

	wlMu.RLock()
	var groupChats []*ChatInfo
	for _, chat := range chatRegistry {
		if chat.Active && (chat.Type == "group" || chat.Type == "supergroup") {
			groupChats = append(groupChats, chat)
		}
	}
	wlMu.RUnlock()

	if len(groupChats) == 0 {
		return c.Send("👥 Нет активных групповых чатов")
	}

	var result strings.Builder
	result.WriteString("👥 Групповые чаты:\n\n")

	for _, chat := range groupChats {
		icon := "👥"
		if chat.Type == "supergroup" {
			icon = "🏢"
		}

		result.WriteString(fmt.Sprintf("%s <b>%s</b>\n", icon, chat.Title))
		result.WriteString(fmt.Sprintf("   ID: <code>%d</code>\n", chat.ID))
		result.WriteString(fmt.Sprintf("   Тип: %s\n", getChatTypeRussian(chat.Type)))
		if chat.Username != "" {
			result.WriteString(fmt.Sprintf("   @%s\n", chat.Username))
		}
		result.WriteString(fmt.Sprintf("   Добавлен: %s\n\n", chat.AddedAt.Format("02.01.2006 15:04")))
	}

	result.WriteString("📝 <i>Чтобы добавить новую группу, напишите /start в нужной группе</i>")

	return c.Send(result.String(), &tb.SendOptions{ParseMode: tb.ModeHTML})
}

func handleStats(c tb.Context) error {
	logger.L.Debug("command stats", "chat", c.Chat().ID)

	stats := GetChatStats()

	var result strings.Builder
	result.WriteString("📈 <b>Статистика чатов:</b>\n\n")
	result.WriteString(fmt.Sprintf("📊 Всего чатов: <b>%d</b>\n", stats["total"]))
	result.WriteString(fmt.Sprintf("✅ Активных: <b>%d</b>\n\n", stats["active"]))
	result.WriteString("📁 <b>По типам:</b>\n")
	result.WriteString(fmt.Sprintf("👤 Личных: <b>%d</b>\n", stats["private"]))
	result.WriteString(fmt.Sprintf("👥 Групп: <b>%d</b>\n", stats["group"]))
	result.WriteString(fmt.Sprintf("🏢 Супергрупп: <b>%d</b>\n", stats["supergroup"]))
	result.WriteString(fmt.Sprintf("📢 Каналов: <b>%d</b>\n", stats["channel"]))

	return c.Send(result.String(), &tb.SendOptions{ParseMode: tb.ModeHTML})
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
		handlerLogger := logger.GetHandlerLogger()
		openaiLogger := logger.GetOpenAILogger()

		op := handlerLogger.Operation("chat_completion")
		op.WithContext("user_id", c.Chat().ID)

		q := sanitizeInput(c.Message().Payload)
		if err := validateChatMessage(q); err != nil {
			handlerLogger.Debug("Invalid chat message", "error", err, "user_id", c.Chat().ID)
			op.Failure("Chat message validation failed", err)
			return c.Send("Message too long or invalid")
		}
		if q == "" {
			op.Failure("Empty chat message", nil)
			return c.Send("Usage: /chat <message>")
		}

		op.WithContext("query_length", len(q))
		op.Step("validating_input")

		handlerLogger.UserAction(c.Chat().ID, "chat", map[string]interface{}{
			"query_length": len(q),
			"model":        getRuntimeConfig().CurrentModel,
		})

		ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
		defer cancel()

		op.Step("calling_openai_api")
		startTime := time.Now()

		resp, err := UserCompletion(ctx, client, q, getRuntimeConfig().CurrentModel)
		duration := time.Since(startTime)

		openaiLogger.APICall("openai", "chat_completion", err == nil, duration, err)

		if err != nil {
			op.Failure("OpenAI API call failed", err)
			return c.Send(DefaultErrorHandler.HandleOpenAIError(err, getRuntimeConfig().CurrentModel))
		}

		op.WithContext("response_length", len(resp))
		op.Success("Chat completion successful", "response_length", len(resp))

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
			return c.Send("🔍 Ошибка поиска. Попробуйте позже.")
		}
		if strings.TrimSpace(result) == "" {
			return c.Send("🤔 Поиск не дал результатов. Попробуйте другой запрос.")
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
