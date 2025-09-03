package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"telegram-reminder/internal/logger"

	openai "github.com/sashabaranov/go-openai"
)

// ChatCompletionSearchProvider implements SearchProvider using OpenAI Chat Completion API
type ChatCompletionSearchProvider struct {
	client ChatCompleter
}

// NewChatCompletionSearchProvider creates a new search provider using Chat Completion API
func NewChatCompletionSearchProvider(client ChatCompleter) SearchProvider {
	return &ChatCompletionSearchProvider{client: client}
}

// Search performs web search using Chat Completion API with web_search tool
func (p *ChatCompletionSearchProvider) Search(ctx context.Context, query string) (string, error) {
	logger.L.Debug("chat completion search", "query", query)

	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: query,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:    getRuntimeConfig().CurrentModel,
		Messages: msgs,
		Tools:    []openai.Tool{webSearchTool},
	}

	if getRuntimeConfig().ToolChoice != "" {
		req.ToolChoice = getRuntimeConfig().ToolChoice
	}

	// Configure model-specific parameters
	if strings.HasPrefix(req.Model, "o3") || strings.HasPrefix(req.Model, "o1") {
		req.MaxCompletionTokens = getRuntimeConfig().MaxTokens
	} else {
		req.Temperature = 0.9
		req.MaxTokens = getRuntimeConfig().MaxTokens
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("chat completion search failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from chat completion")
	}

	msg := resp.Choices[0].Message

	// Handle tool calls for web search
	if len(msg.ToolCalls) > 0 {
		toolMsgs := make([]openai.ChatCompletionMessage, 0, len(msg.ToolCalls))
		for _, tc := range msg.ToolCalls {
			if tc.Type != openai.ToolTypeFunction || tc.Function.Name != "web_search" {
				continue
			}

			var params struct {
				Query string `json:"query"`
			}

			if err := json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
				continue
			}

			// Use responses API for actual search
			searchResult, err := p.performActualSearch(ctx, params.Query)
			if err != nil {
				searchResult = err.Error()
			}
			if strings.TrimSpace(searchResult) == "" {
				searchResult = "Поиск не дал результатов"
			}

			toolMsgs = append(toolMsgs, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: tc.ID,
				Content:    searchResult,
			})
		}

		if len(toolMsgs) > 0 {
			msgs = append(msgs, msg)
			msgs = append(msgs, toolMsgs...)
			req.Messages = msgs

			resp, err = p.client.CreateChatCompletion(ctx, req)
			if err != nil {
				return "", fmt.Errorf("follow-up completion failed: %w", err)
			}

			if len(resp.Choices) == 0 {
				return "", fmt.Errorf("no follow-up response")
			}

			msg = resp.Choices[0].Message
		}
	}

	return strings.TrimSpace(msg.Content), nil
}

// performActualSearch calls the responses API for web search
func (p *ChatCompletionSearchProvider) performActualSearch(ctx context.Context, query string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	return ResponsesCompletion(ctx, apiKey, query, getRuntimeConfig().CurrentModel)
}

// SupportsModel checks if the model supports web search
func (p *ChatCompletionSearchProvider) SupportsModel(model string) bool {
	return supportsWebSearch(model)
}

// ResponsesSearchProvider implements SearchProvider using OpenAI Responses API
type ResponsesSearchProvider struct {
	apiKey string
}

// NewResponsesSearchProvider creates a new search provider using Responses API
func NewResponsesSearchProvider(apiKey string) SearchProvider {
	return &ResponsesSearchProvider{apiKey: apiKey}
}

// Search performs web search using Responses API
func (p *ResponsesSearchProvider) Search(ctx context.Context, query string) (string, error) {
	logger.L.Debug("responses search", "query", query)

	result, err := ResponsesCompletion(ctx, p.apiKey, query, getRuntimeConfig().CurrentModel)
	if err != nil {
		return "", fmt.Errorf("responses search failed: %w", err)
	}

	return result, nil
}

// SupportsModel checks if the model supports web search
func (p *ResponsesSearchProvider) SupportsModel(model string) bool {
	return supportsWebSearch(model)
}
