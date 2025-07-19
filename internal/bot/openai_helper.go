package bot

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

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
//   - model: OpenAI model name (e.g., "gpt-4o", "gpt-3.5-turbo")
//
// Returns:
//   - string: The generated response text, trimmed of whitespace
//   - error: Any error that occurred during the API call
func ChatCompletion(ctx context.Context, client ChatCompleter, msgs []openai.ChatCompletionMessage, model string) (string, error) {
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    msgs,
		Temperature: 0.9,
		MaxTokens:   600,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
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
