package bot

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

// ChatCompleter defines the interface for OpenAI chat completion
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// StreamCompleter defines the interface for OpenAI streaming chat completion
type StreamCompleter interface {
	CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error)
}

// AIClient combines all AI-related operations
type AIClient interface {
	ChatCompleter
	StreamCompleter
}

// Ensure that openai.Client implements our interfaces
var _ ChatCompleter = (*openai.Client)(nil)
var _ StreamCompleter = (*openai.Client)(nil)
var _ AIClient = (*openai.Client)(nil)
