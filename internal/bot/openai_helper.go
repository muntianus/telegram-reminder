package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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

func normalizeQuery(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return q
	}
	return strings.Join(strings.Fields(q), " ")
}

func supportsWebSearch(model string) bool {
	for _, m := range SupportedModels {
		if model == m {
			return true
		}
	}
	return false
}

// defaultWebSearch performs a search using the configured search service
func defaultWebSearch(ctx context.Context, query string) (string, error) {
	searchService := getSearchService()
	return searchService.Search(ctx, query)
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
		// OpenAI stream error logging removed
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
		// OpenAI error logging removed
		return "", err
	}
	// OpenAI response debug logging removed
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
				res = "Поиск не дал результатов"
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

	// Log successful LLM responses with readable text
	if len(out) > 0 {
		// Create readable preview for logs
		preview := out
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		logger.L.Info("LLM response generated", "model", model, "length", len(out), "preview", preview)
	} else {
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
