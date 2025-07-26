package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

// SearchProviderURL is the template URL for performing web searches. It is set
// from configuration at startup.
var SearchProviderURL string

// searchFunc performs a web search and returns plain text results. It can be
// overridden in tests.
var searchFunc = defaultWebSearch

func supportsWebSearch(model string) bool {
	for _, m := range SupportedModels {
		if model == m {
			return true
		}
	}
	return false
}

// defaultWebSearch queries SearchProviderURL using GET and returns the body as a
// string. It limits the response to 2000 bytes.
func defaultWebSearch(ctx context.Context, query string) (string, error) {
	logger.L.Debug("web search", "query", query)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}
	return ResponsesCompletion(ctx, apiKey, query, CurrentModel)
}

// ChatCompleter abstracts the OpenAI client method used by chatCompletion.
// This interface allows for easier testing and mocking of OpenAI API calls.
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
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
	if EnableWebSearch && supportsWebSearch(model) {
		req.Tools = []openai.Tool{webSearchTool}
	}

	// Configure parameters based on model type
	if strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o1") {
		// o3/o1 models have fixed parameters: temperature=1, top_p=1, n=1
		// presence_penalty and frequency_penalty are fixed at 0
		req.MaxCompletionTokens = 600
	} else if strings.HasPrefix(model, "gpt-4o") {
		// GPT-4o models support web search (handled automatically by OpenAI)
		req.Temperature = 0.9
		req.MaxTokens = 600
	} else {
		// Standard models support custom parameters
		req.Temperature = 0.9
		req.MaxTokens = 600
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
	if EnableWebSearch && len(msg.ToolCalls) > 0 {
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
	logger.L.Debug("openai result", "length", len(out))
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
