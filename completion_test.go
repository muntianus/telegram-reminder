package main

import (
	"context"
	"errors"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

type mockAI struct {
	received openai.ChatCompletionRequest
	resp     openai.ChatCompletionResponse
	err      error
}

func (m *mockAI) CreateChatCompletion(_ context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	m.received = req
	return m.resp, m.err
}

func TestChatCompletionSuccess(t *testing.T) {
	m := &mockAI{resp: openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{
		{Message: openai.ChatCompletionMessage{Content: "  hi "}},
	}}}
	got, err := chatCompletion(m, "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hi" {
		t.Errorf("got %q", got)
	}
	if m.received.Messages[0].Content != "prompt" {
		t.Errorf("prompt not forwarded")
	}
}

func TestChatCompletionNoChoices(t *testing.T) {
	m := &mockAI{}
	got, err := chatCompletion(m, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestChatCompletionError(t *testing.T) {
	m := &mockAI{err: errors.New("boom")}
	_, err := chatCompletion(m, "test")
	if err == nil {
		t.Fatal("expected error")
	}
}
