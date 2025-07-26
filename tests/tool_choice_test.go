package main

import (
	"context"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	botpkg "telegram-reminder/internal/bot"
)

type recordClient struct {
	req  openai.ChatCompletionRequest
	resp openai.ChatCompletionResponse
}

func (r *recordClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	r.req = req
	return r.resp, nil
}

func TestChatCompletionToolChoice(t *testing.T) {
	botpkg.OpenAIToolChoice = "none"
	rc := &recordClient{resp: openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "hi"}}}}}
	_, err := botpkg.ChatCompletion(context.Background(), rc, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "q"}}, "gpt-4o")
	if err != nil {
		t.Fatalf("chat completion error: %v", err)
	}
	if rc.req.ToolChoice != "none" {
		t.Fatalf("expected tool_choice none, got %v", rc.req.ToolChoice)
	}
	if rc.req.Tools != nil {
		t.Fatalf("expected no tools when tool_choice none")
	}
}
