package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"telegram-reminder/internal/logger"

	openai "github.com/sashabaranov/go-openai"
)

var webSearchTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        "web_search",
		Description: "Search the web for a query and return top results",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query text",
				},
			},
			"required": []string{"query"},
		},
	},
}

// searchFunc performs a web search and returns plain text results. It can be
// overridden in tests.
var searchFunc = defaultWebSearch

type searchEntry struct {
	result   string
	ts       time.Time
	accessed time.Time
}

var (
	searchCache     = map[string]searchEntry{}
	searchCacheTTL  = 10 * time.Minute
	searchCacheSize = 100 // Maximum cache entries
	searchMu        sync.RWMutex
	searchAPIFunc   = realSearchAPI
)

func realSearchAPI(ctx context.Context, query string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}
	return ResponsesCompletion(ctx, apiKey, query, getRuntimeConfig().CurrentModel)
}

func normalizeQuery(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return q
	}
	return strings.Join(strings.Fields(q), " ")
}

func getCachedSearch(query string) (string, bool) {
	searchMu.Lock()
	defer searchMu.Unlock()
	e, ok := searchCache[query]
	if !ok || time.Since(e.ts) > searchCacheTTL {
		return "", false
	}
	// Update access time for LRU
	e.accessed = time.Now()
	searchCache[query] = e
	return e.result, true
}

func setCachedSearch(query, result string) {
	searchMu.Lock()
	defer searchMu.Unlock()

	now := time.Now()
	entry := searchEntry{
		result:   result,
		ts:       now,
		accessed: now,
	}

	// Clean expired entries first
	cleanExpiredSearchCache()

	// If cache is at capacity, remove LRU entry
	if len(searchCache) >= searchCacheSize {
		removeLRUSearchEntry()
	}

	searchCache[query] = entry
}

// cleanExpiredSearchCache removes expired entries from cache
func cleanExpiredSearchCache() {
	now := time.Now()
	for query, entry := range searchCache {
		if now.Sub(entry.ts) > searchCacheTTL {
			delete(searchCache, query)
		}
	}
}

// removeLRUSearchEntry removes the least recently used entry
func removeLRUSearchEntry() {
	if len(searchCache) == 0 {
		return
	}

	var oldestQuery string
	var oldestTime time.Time = time.Now()

	for query, entry := range searchCache {
		if entry.accessed.Before(oldestTime) {
			oldestTime = entry.accessed
			oldestQuery = query
		}
	}

	if oldestQuery != "" {
		delete(searchCache, oldestQuery)
	}
}

func supportsWebSearch(model string) bool {
	for _, m := range SupportedModels {
		if model == m {
			return true
		}
	}
	return false
}

// defaultWebSearch performs a search using the OpenAI search API and returns the
// plain text result.
func defaultWebSearch(ctx context.Context, query string) (string, error) {
	logger.L.Debug("web search", "query", query)
	q := normalizeQuery(query)
	if res, ok := getCachedSearch(q); ok {
		logger.L.Debug("web search cache hit", "query", q)
		return res, nil
	}
	res, err := searchAPIFunc(ctx, q)
	if err != nil {
		return "", err
	}
	setCachedSearch(q, res)
	return res, nil
}

// StreamChatCompletion sends messages to OpenAI using the streaming API and
// returns a channel with incremental text parts as they are produced.
func StreamChatCompletion(ctx context.Context, client StreamCompleter, msgs []openai.ChatCompletionMessage, model string) (<-chan string, error) {
	logger.L.Debug("chat completion stream", "model", model, "messages", len(msgs))
	outCh := make(chan string)
	if len(msgs) == 0 {
		close(outCh)
		return outCh, nil
	}
	for _, m := range msgs {
		if strings.TrimSpace(m.Content) == "" {
			close(outCh)
			return outCh, nil
		}
	}
	timeMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("Current datetime: %s", time.Now().Format(time.RFC3339)),
	}
	msgs = append(msgs, timeMsg)

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: msgs,
		Stream:   true,
	}
	if getRuntimeConfig().ServiceTier != "" {
		req.ServiceTier = getRuntimeConfig().ServiceTier
	}
	if getRuntimeConfig().ReasoningEffort != "" {
		req.ReasoningEffort = getRuntimeConfig().ReasoningEffort
	}
	if getRuntimeConfig().EnableWebSearch && supportsWebSearch(model) {
		req.Tools = []openai.Tool{webSearchTool}
	}
	if getRuntimeConfig().ToolChoice != "" {
		req.ToolChoice = getRuntimeConfig().ToolChoice
		if getRuntimeConfig().ToolChoice == "none" {
			req.Tools = nil
		}
	}
	if strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o1") {
		req.MaxCompletionTokens = getRuntimeConfig().MaxTokens
	} else if strings.HasPrefix(model, "gpt-4o") {
		req.Temperature = 0.9
		req.MaxTokens = getRuntimeConfig().MaxTokens
	} else {
		req.Temperature = 0.9
		req.MaxTokens = getRuntimeConfig().MaxTokens
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		logger.L.Debug("openai stream error", "err", err)
		close(outCh)
		return outCh, err
	}

	go func() {
		defer func() {
			if err := stream.Close(); err != nil {
				logger.L.Debug("stream close error", "err", err)
			}
			close(outCh)
		}()
		for {
			select {
			case <-ctx.Done():
				logger.L.Debug("stream context cancelled", "err", ctx.Err())
				return
			default:
				resp, err := stream.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return
					}
					logger.L.Debug("stream recv error", "err", err)
					return
				}
				if len(resp.Choices) == 0 {
					continue
				}
				delta := resp.Choices[0].Delta.Content
				if strings.TrimSpace(delta) == "" {
					continue
				}

				// Try to send with timeout to prevent blocking
				select {
				case outCh <- delta:
				case <-ctx.Done():
					logger.L.Debug("stream send cancelled", "err", ctx.Err())
					return
				case <-time.After(5 * time.Second):
					logger.L.Debug("stream send timeout")
					return
				}
			}
		}
	}()

	return outCh, nil
}

// ChatCompletion sends messages to OpenAI and returns the reply text using the specified model.
//
// Parameters:
//   - ctx: Context for the request with timeout
//   - client: OpenAI client implementing ChatCompleter interface
//   - msgs: Array of chat messages to send
//   - model: OpenAI model name (e.g., "o3", "gpt-3.5-turbo")
//
// Returns:
//   - string: The generated response text, trimmed of whitespace
//   - error: Any error that occurred during the API call
func ChatCompletion(ctx context.Context, client ChatCompleter, msgs []openai.ChatCompletionMessage, model string) (string, error) {
	logger.L.Debug("chat completion", "model", model, "messages", len(msgs))
	if len(msgs) == 0 {
		return "", nil
	}
	for _, m := range msgs {
		if strings.TrimSpace(m.Content) == "" {
			return "", nil
		}
	}
	// Append current date and time as a system message
	timeMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("Current datetime: %s", time.Now().Format(time.RFC3339)),
	}
	msgs = append(msgs, timeMsg)

	// Create base request
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: msgs,
	}
	if getRuntimeConfig().ServiceTier != "" {
		req.ServiceTier = getRuntimeConfig().ServiceTier
	}
	if getRuntimeConfig().ReasoningEffort != "" {
		req.ReasoningEffort = getRuntimeConfig().ReasoningEffort
	}
	if getRuntimeConfig().EnableWebSearch && supportsWebSearch(model) {
		req.Tools = []openai.Tool{webSearchTool}
	}

	if getRuntimeConfig().ToolChoice != "" {
		req.ToolChoice = getRuntimeConfig().ToolChoice
		if getRuntimeConfig().ToolChoice == "none" {
			req.Tools = nil
		}
	}

	// Configure parameters based on model type
	if strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o1") {
		// o3/o1 models have fixed parameters: temperature=1, top_p=1, n=1
		// presence_penalty and frequency_penalty are fixed at 0
		req.MaxCompletionTokens = getRuntimeConfig().MaxTokens
	} else if strings.HasPrefix(model, "gpt-4o") {
		// GPT-4o models support web search (handled automatically by OpenAI)
		req.Temperature = 0.9
		req.MaxTokens = getRuntimeConfig().MaxTokens
	} else {
		// Standard models support custom parameters
		req.Temperature = 0.9
		req.MaxTokens = getRuntimeConfig().MaxTokens
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		logger.L.Debug("openai error", "err", err)
		return "", err
	}
	logger.L.Debug("openai response", "choices", len(resp.Choices))
	if len(resp.Choices) == 0 {
		return "", nil
	}

	msg := resp.Choices[0].Message
	if getRuntimeConfig().EnableWebSearch && len(msg.ToolCalls) > 0 {
		toolMsgs := make([]openai.ChatCompletionMessage, 0, len(msg.ToolCalls))
		for _, tc := range msg.ToolCalls {
			if tc.Type != openai.ToolTypeFunction || tc.Function.Name != "web_search" {
				continue
			}
			var p struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &p); err != nil {
				continue
			}
			res, err := searchFunc(ctx, p.Query)
			if err != nil {
				res = err.Error()
			}
			if strings.TrimSpace(res) == "" {
				res = "no results"
			}
			toolMsgs = append(toolMsgs, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: tc.ID,
				Content:    res,
			})
		}
		if len(toolMsgs) > 0 {
			msgs = append(msgs, msg)
			msgs = append(msgs, toolMsgs...)
			req.Messages = msgs
			resp, err = client.CreateChatCompletion(ctx, req)
			if err != nil {
				return "", err
			}
			if len(resp.Choices) == 0 {
				return "", nil
			}
			msg = resp.Choices[0].Message
		}
	}
	out := strings.TrimSpace(msg.Content)
	logger.L.Debug("openai result", "length", len(out), "preview", logger.Truncate(out, 200))
	if len(out) == 0 {
		logger.L.Warn("empty openai response", "msg_content", msg.Content, "msg_role", msg.Role)
	}
	return out, nil
}

// SystemCompletion generates a reply to a system-level prompt using OpenAI.
// This function is used for tasks that require system-level instructions.
//
// Parameters:
//   - ctx: Context for the request with timeout
//   - client: OpenAI client implementing ChatCompleter interface
//   - prompt: System prompt to send to the model
//   - model: OpenAI model name
//
// Returns:
//   - string: The generated response text
//   - error: Any error that occurred during the API call
func SystemCompletion(ctx context.Context, client ChatCompleter, prompt, model string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}}
	return ChatCompletion(ctx, client, msgs, model)
}

// UserCompletion generates a reply to a user's message using OpenAI.
// This function is used for direct user interactions and chat functionality.
//
// Parameters:
//   - ctx: Context for the request with timeout
//   - client: OpenAI client implementing ChatCompleter interface
//   - message: User message to send to the model
//   - model: OpenAI model name
//
// Returns:
//   - string: The generated response text
//   - error: Any error that occurred during the API call
func UserCompletion(ctx context.Context, client ChatCompleter, message, model string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: message}}
	return ChatCompletion(ctx, client, msgs, model)
}
