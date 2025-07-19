package bot

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// ChatCompleter abstracts the OpenAI client method used by chatCompletion.
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// ChatCompletion sends messages to OpenAI and returns the reply text using the current model.
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
func SystemCompletion(ctx context.Context, client ChatCompleter, prompt, model string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}}
	return ChatCompletion(ctx, client, msgs, model)
}

// UserCompletion generates a reply to a user's message using OpenAI.
func UserCompletion(ctx context.Context, client ChatCompleter, message, model string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: message}}
	return ChatCompletion(ctx, client, msgs, model)
}
