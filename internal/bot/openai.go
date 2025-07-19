package bot

import (
	"context"
	"strings"

	"sync"

	openai "github.com/sashabaranov/go-openai"
	tb "gopkg.in/telebot.v3"
)

// CurrentModel — текущая модель OpenAI для генерации ответов.
// ModelMu — мьютекс для потокобезопасного доступа к CurrentModel.
// SupportedModels — список поддерживаемых моделей OpenAI (для справки и тестов).
var (
	CurrentModel string = "gpt-4o"
	ModelMu      sync.RWMutex
)

var SupportedModels = []string{"gpt-4o", "gpt-3.5-turbo"}

// ChatCompleter абстрагирует клиента OpenAI для unit-тестов и реального использования.
type ChatCompleter interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// MessageSender реализуется типами, которые могут отправлять сообщения в Telegram.
type MessageSender interface {
	Send(recipient tb.Recipient, what interface{}, opts ...interface{}) (*tb.Message, error)
}

// ChatCompletion отправляет сообщения в OpenAI и возвращает ответ, используя текущую модель.
func ChatCompletion(ctx context.Context, client ChatCompleter, msgs []openai.ChatCompletionMessage) (string, error) {
	ModelMu.RLock()
	m := CurrentModel
	ModelMu.RUnlock()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       m,
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

// SystemCompletion генерирует ответ на системный prompt через OpenAI.
func SystemCompletion(ctx context.Context, client ChatCompleter, prompt string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleSystem, Content: prompt}}
	return ChatCompletion(ctx, client, msgs)
}

// UserCompletion генерирует ответ на пользовательское сообщение через OpenAI.
func UserCompletion(ctx context.Context, client ChatCompleter, message string) (string, error) {
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: message}}
	return ChatCompletion(ctx, client, msgs)
}
