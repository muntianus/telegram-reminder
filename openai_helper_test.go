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

// captureClient records the request sent to CreateChatCompletion
// so tests can inspect the parameters.
type captureClient struct{ received openai.ChatCompletionRequest }

func (c *captureClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	c.received = req
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{
		{Message: openai.ChatCompletionMessage{Content: "ok"}},
	}}, nil
}

func TestChatCompletionWithError(t *testing.T) {
	client := errorClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	_, err := botpkg.ChatCompletion(ctx, client, msgs, "o3")
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

	resp, err := botpkg.ChatCompletion(ctx, client, msgs, "o3")
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

	resp, err := botpkg.ChatCompletion(ctx, client, msgs, "o3")
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

	_, err := botpkg.SystemCompletion(ctx, client, "test prompt", "o3")
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

	_, err := botpkg.UserCompletion(ctx, client, "test message", "o3")
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

	resp, err := botpkg.ChatCompletion(ctx, client, msgs, "o3")
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

	resp, err := botpkg.SystemCompletion(ctx, client, "", "o3")
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

	resp, err := botpkg.UserCompletion(ctx, client, "", "o3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "" {
		t.Errorf("expected empty response for empty message, got: %q", resp)
	}
}

func TestChatCompletionAddsWebSearchTool(t *testing.T) {
	c := &captureClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	if _, err := botpkg.ChatCompletion(ctx, c, msgs, "gpt-4o"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.received.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(c.received.Tools))
	}
	if c.received.Tools[0].Type != openai.ToolType("web_search") {
		t.Errorf("expected web_search tool, got %v", c.received.Tools[0].Type)
	}
}

func TestChatCompletionNoWebSearchForUnsupportedModel(t *testing.T) {
	c := &captureClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	if _, err := botpkg.ChatCompletion(ctx, c, msgs, "gpt-3.5-turbo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.received.Tools) != 0 {
		t.Fatalf("expected no tools, got %d", len(c.received.Tools))
	}
}
