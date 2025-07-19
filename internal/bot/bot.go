package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

// WebSearchResult represents a single search result
type WebSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearch performs a web search using DuckDuckGo API
func WebSearch(query string) ([]WebSearchResult, error) {
	// Using DuckDuckGo Instant Answer API (free, no API key required)
	baseURL := "https://api.duckduckgo.com/"
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("no_html", "1")
	params.Add("skip_disambig", "1")

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("web search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Abstract      string `json:"Abstract"`
		AbstractURL   string `json:"AbstractURL"`
		AbstractText  string `json:"AbstractText"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
		Results []struct {
			Title string `json:"Title"`
			URL   string `json:"URL"`
			Text  string `json:"Text"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	var results []WebSearchResult

	// Add main result if available
	if result.Abstract != "" {
		results = append(results, WebSearchResult{
			Title:   "Main Result",
			URL:     result.AbstractURL,
			Snippet: result.Abstract,
		})
	}

	// Add specific results
	for _, res := range result.Results {
		results = append(results, WebSearchResult{
			Title:   res.Title,
			URL:     res.URL,
			Snippet: res.Text,
		})
	}

	// Add related topics if no specific results
	if len(result.Results) == 0 {
		for i, topic := range result.RelatedTopics {
			if i >= 5 { // Limit to 5 results
				break
			}
			results = append(results, WebSearchResult{
				Title:   topic.Text,
				URL:     topic.FirstURL,
				Snippet: topic.Text,
			})
		}
	}

	return results, nil
}

// EnhancedSystemCompletion performs a web search and then generates a response
func EnhancedSystemCompletion(ctx context.Context, client *openai.Client, prompt string, model string) (string, error) {
	// Extract search queries from prompt
	searchQueries := extractSearchQueries(prompt)

	var webContext string
	if len(searchQueries) > 0 {
		webContext = "🔍 АКТУАЛЬНАЯ ИНФОРМАЦИЯ ИЗ ИНТЕРНЕТА:\n\n"

		for _, query := range searchQueries {
			results, err := WebSearch(query)
			if err != nil {
				logger.L.Error("web search failed", "query", query, "err", err)
				continue
			}

			webContext += fmt.Sprintf("📊 Поиск: %s\n", query)
			for i, result := range results {
				if i >= 3 { // Limit to 3 results per query
					break
				}
				// Format with more specific source information
				webContext += fmt.Sprintf("• %s\n  %s\n  Источник: %s\n\n",
					result.Title, result.Snippet, result.URL)
			}
		}
	}

	// Combine web context with original prompt
	enhancedPrompt := prompt
	if webContext != "" {
		enhancedPrompt = webContext + "\n" + prompt
	}

	return SystemCompletion(ctx, client, enhancedPrompt, model)
}

// extractSearchQueries extracts search queries from prompt with more specific queries for today's news
func extractSearchQueries(prompt string) []string {
	var queries []string

	// Extract queries based on prompt type with more specific searches for today's news
	if strings.Contains(prompt, "криптовалют") || strings.Contains(prompt, "crypto") {
		queries = append(queries,
			"bitcoin news today",
			"cryptocurrency news today",
			"crypto market news today",
			"defi news today",
			"ethereum news today",
			"altcoin news today",
			"crypto regulation news today",
			"crypto exchange news today")
	}

	if strings.Contains(prompt, "технолог") || strings.Contains(prompt, "tech") {
		queries = append(queries,
			"AI news today",
			"startup news today",
			"tech company news today",
			"product hunt today",
			"new AI models today",
			"tech IPO news today",
			"artificial intelligence news today",
			"tech funding news today",
			"software news today",
			"tech acquisitions today")
	}

	if strings.Contains(prompt, "недвижимость") || strings.Contains(prompt, "real estate") {
		queries = append(queries,
			"Москва недвижимость новости сегодня",
			"Подмосковье недвижимость новости сегодня",
			"ГИС-Торги новости сегодня",
			"недвижимость новости Москва сегодня",
			"цены на недвижимость новости сегодня",
			"земельные участки новости сегодня",
			"новостройки Москва новости сегодня",
			"коммерческая недвижимость новости сегодня",
			"ипотека новости сегодня",
			"недвижимость аналитика сегодня")
	}

	if strings.Contains(prompt, "бизнес") || strings.Contains(prompt, "business") {
		queries = append(queries,
			"business news today",
			"startup news today",
			"IPO news today",
			"venture capital news today",
			"company earnings news today",
			"mergers acquisitions news today",
			"business trends today",
			"entrepreneurship news today",
			"business technology news today",
			"market news today")
	}

	if strings.Contains(prompt, "инвестиции") || strings.Contains(prompt, "investment") {
		queries = append(queries,
			"stock market news today",
			"investment news today",
			"market analysis today",
			"financial news today",
			"stock prices news today",
			"market trends today",
			"investment opportunities today",
			"portfolio news today",
			"trading news today",
			"wealth management news today")
	}

	if strings.Contains(prompt, "стартап") || strings.Contains(prompt, "startup") {
		queries = append(queries,
			"startup news today",
			"startup funding news today",
			"new startups launched today",
			"venture capital news today",
			"startup acquisitions news today",
			"startup IPO news today",
			"startup ecosystem news today",
			"startup technology news today",
			"startup trends today",
			"startup success stories today")
	}

	if strings.Contains(prompt, "глобаль") || strings.Contains(prompt, "global") {
		queries = append(queries,
			"world news today",
			"global economy news today",
			"international news today",
			"geopolitical news today",
			"world markets news today",
			"global trade news today",
			"world politics news today",
			"international relations news today",
			"global business news today",
			"world events today")
	}

	return queries
}

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

🔍 ИНТЕРНЕТ-АНАЛИЗ: Используй СВЕЖУЮ ИНФОРМАЦИЮ ИЗ СТАТЕЙ ЗА ПОСЛЕДНИЕ 24 ЧАСА по темам:
- Криптовалюты и DeFi
- Технологии и стартапы
- Недвижимость и инвестиции
- Бизнес-тренды

ВАЖНО: Все ссылки должны вести на СВЕЖИЕ СТАТЬИ ЗА ПОСЛЕДНИЕ 24 ЧАСА, а не на старые данные.

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
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, CurrentModel)
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
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, CurrentModel)
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
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, CurrentModel)
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
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, CurrentModel)
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
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, CurrentModel)
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
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, CurrentModel)
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
		resp, err := EnhancedSystemCompletion(ctx, client, prompt, CurrentModel)
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
