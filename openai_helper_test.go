package main

import (
	"context"
	"errors"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
)

// errorClient implements ChatCompleter and always returns an error
type errorClient struct{}

func (e errorClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{}, errors.New("API error")
}

// emptyResponseClient implements ChatCompleter and returns empty responses
type emptyResponseClient struct{}

func (e emptyResponseClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{}}, nil
}

// whitespaceClient implements ChatCompleter and returns responses with whitespace
type whitespaceClient struct{}

func (w whitespaceClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{Message: openai.ChatCompletionMessage{Content: "  response with spaces  \n"}},
		},
	}, nil
}

func TestChatCompletionWithError(t *testing.T) {
	client := errorClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	_, err := botpkg.ChatCompletion(ctx, client, msgs, "gpt-4o")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "API error" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatCompletionEmptyResponse(t *testing.T) {
	client := emptyResponseClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	resp, err := botpkg.ChatCompletion(ctx, client, msgs, "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response, got: %q", resp)
	}
}

func TestChatCompletionWhitespaceTrimming(t *testing.T) {
	client := whitespaceClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	resp, err := botpkg.ChatCompletion(ctx, client, msgs, "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "response with spaces"
	if resp != expected {
		t.Errorf("expected %q, got %q", expected, resp)
	}
}

func TestSystemCompletionWithError(t *testing.T) {
	client := errorClient{}
	ctx := context.Background()

	_, err := botpkg.SystemCompletion(ctx, client, "test prompt", "gpt-4o")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "API error" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUserCompletionWithError(t *testing.T) {
	client := errorClient{}
	ctx := context.Background()

	_, err := botpkg.UserCompletion(ctx, client, "test message", "gpt-4o")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "API error" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChatCompletionEmptyMessages(t *testing.T) {
	client := emptyResponseClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{}

	resp, err := botpkg.ChatCompletion(ctx, client, msgs, "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response for empty messages, got: %q", resp)
	}
}

func TestSystemCompletionEmptyPrompt(t *testing.T) {
	client := emptyResponseClient{}
	ctx := context.Background()

	resp, err := botpkg.SystemCompletion(ctx, client, "", "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response for empty prompt, got: %q", resp)
	}
}

func TestUserCompletionEmptyMessage(t *testing.T) {
	client := emptyResponseClient{}
	ctx := context.Background()

	resp, err := botpkg.UserCompletion(ctx, client, "", "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response for empty message, got: %q", resp)
	}
}
