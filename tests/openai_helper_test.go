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

// captureClient stores the last request passed to CreateChatCompletion
type captureClient struct {
	req openai.ChatCompletionRequest
}

func (c *captureClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	c.req = req
	return openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "ok"}}}}, nil
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

func TestWebSearchToolAdded(t *testing.T) {
	orig := botpkg.EnableWebSearch
	botpkg.EnableWebSearch = true
	defer func() { botpkg.EnableWebSearch = orig }()
	client := &captureClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	if _, err := botpkg.ChatCompletion(ctx, client, msgs, "gpt-4o"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.req.Tools) == 0 {
		t.Fatalf("web search tool not added")
	}
	if client.req.Tools[0].Type != openai.ToolTypeFunction {
		t.Fatalf("unexpected tool type: %v", client.req.Tools[0].Type)
	}
	if client.req.Tools[0].Function == nil || client.req.Tools[0].Function.Name != "web_search" {
		t.Fatalf("missing web_search function definition")
	}
}

func TestWebSearchToolNotAddedForUnsupportedModel(t *testing.T) {
	orig := botpkg.EnableWebSearch
	botpkg.EnableWebSearch = true
	defer func() { botpkg.EnableWebSearch = orig }()
	client := &captureClient{}
	ctx := context.Background()
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}}

	if _, err := botpkg.ChatCompletion(ctx, client, msgs, "gpt-3.5"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.req.Tools) != 0 {
		t.Fatalf("unexpected tools: %+v", client.req.Tools)
	}
}

func TestEnhancedSystemCompletionUsesWebSearch(t *testing.T) {
	orig := botpkg.EnableWebSearch
	botpkg.EnableWebSearch = true
	defer func() { botpkg.EnableWebSearch = orig }()
	client := &captureClient{}
	ctx := context.Background()
	_, err := botpkg.EnhancedSystemCompletion(ctx, client, "prompt", "gpt-4o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.req.Tools) == 0 {
		t.Fatalf("tools not added in enhanced completion")
	}
}
