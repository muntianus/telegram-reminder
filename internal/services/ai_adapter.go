package services

import (
	"context"
	"strings"

	"telegram-reminder/internal/logger"
	
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIAdapter adapts the existing OpenAI client to our service interface
type OpenAIAdapter struct {
	client ChatCompleter
}

// ChatCompleter defines the interface for OpenAI chat completion
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// NewOpenAIAdapter creates a new OpenAI adapter
func NewOpenAIAdapter(client ChatCompleter) *OpenAIAdapter {
	return &OpenAIAdapter{client: client}
}

// EnhancedSystemCompletion implements the digest generation using the existing OpenAI helper
func (a *OpenAIAdapter) EnhancedSystemCompletion(ctx context.Context, prompt, model string) (string, error) {
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: prompt},
	}

	return a.chatCompletion(ctx, msgs, model)
}

// chatCompletion performs the actual chat completion
func (a *OpenAIAdapter) chatCompletion(ctx context.Context, msgs []openai.ChatCompletionMessage, model string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: msgs,
	}

	// Apply runtime configuration
	config := getRuntimeConfig()
	
	if config.ServiceTier != "" {
		req.ServiceTier = config.ServiceTier
	}
	if config.ReasoningEffort != "" {
		req.ReasoningEffort = config.ReasoningEffort
	}
	if config.EnableWebSearch && supportsWebSearch(model) {
		req.Tools = []openai.Tool{getWebSearchTool()}
	}
	if config.ToolChoice != "" {
		req.ToolChoice = config.ToolChoice
		if config.ToolChoice == "none" {
			req.Tools = nil
		}
	}

	// Set token limits based on model
	if strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o1") {
		req.MaxCompletionTokens = config.MaxTokens
	} else if strings.HasPrefix(model, "gpt-4o") {
		req.Temperature = 0.9
		req.MaxTokens = config.MaxTokens
	} else {
		req.Temperature = 0.9
		req.MaxTokens = config.MaxTokens
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		logger.L.Warn("empty openai response choices")
		return "", nil
	}

	msg := resp.Choices[0].Message
	
	// Handle tool calls for web search
	if config.EnableWebSearch && len(msg.ToolCalls) > 0 {
		// This would integrate with the existing web search functionality
		// For now, just return the content
		logger.L.Debug("tool calls detected", "count", len(msg.ToolCalls))
	}

	out := strings.TrimSpace(msg.Content)
	logger.L.Debug("openai result", "length", len(out), "preview", truncateString(out, 200))
	
	if len(out) == 0 {
		logger.L.Warn("empty openai response", "msg_content", msg.Content, "msg_role", msg.Role)
	}
	
	return out, nil
}

// RuntimeConfig represents the runtime configuration
type RuntimeConfig struct {
	CurrentModel     string
	MaxTokens        int
	ServiceTier      openai.ServiceTier
	ReasoningEffort  string
	EnableWebSearch  bool
	ToolChoice       string
}

// getRuntimeConfig returns the current runtime configuration
// This would be injected via DI in a real implementation
func getRuntimeConfig() RuntimeConfig {
	// Placeholder - this would come from the actual config service
	return RuntimeConfig{
		CurrentModel:    "gpt-4.1",
		MaxTokens:       600,
		EnableWebSearch: true,
		ToolChoice:      "auto",
	}
}

// supportsWebSearch checks if a model supports web search
func supportsWebSearch(model string) bool {
	supportedModels := []string{
		"gpt-4o", "gpt-4o-2024-05-13", "gpt-4o-2024-08-06", "gpt-4o-2024-11-20",
		"chatgpt-4o-latest", "gpt-4o-mini", "gpt-4o-mini-2024-07-18",
		"gpt-4-turbo", "gpt-4-turbo-2024-04-09", "gpt-4-0125-preview",
		"gpt-4-1106-preview", "gpt-4-turbo-preview", "gpt-4-vision-preview",
		"gpt-4", "gpt-4.1", "gpt-4.1-2025-04-14", "gpt-4.1-mini",
		"gpt-4.1-mini-2025-04-14", "gpt-4.1-nano", "gpt-4.1-nano-2025-04-14",
	}
	
	for _, supported := range supportedModels {
		if model == supported {
			return true
		}
	}
	return false
}

// getWebSearchTool returns the web search tool definition
func getWebSearchTool() openai.Tool {
	return openai.Tool{
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
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}