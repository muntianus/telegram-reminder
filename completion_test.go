package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"telegram-reminder/internal/bot"
)

// helper to create client backed by test server
func newTestClient(handler http.HandlerFunc) (*openai.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := openai.DefaultConfig("test")
	cfg.BaseURL = srv.URL + "/"
	c := srv.Client()
	c.Timeout = bot.OpenAITimeout
	cfg.HTTPClient = c
	return openai.NewClientWithConfig(cfg), srv
}

func TestChatCompletionSuccess(t *testing.T) {
	var received openai.ChatCompletionRequest
	client, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if err := json.NewEncoder(w).Encode(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{
			{Message: openai.ChatCompletionMessage{Content: "  hi "}},
		}}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	})
	defer srv.Close()

	msg := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "prompt"}
	got, err := bot.ChatCompletion(context.Background(), client, []openai.ChatCompletionMessage{msg})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hi" {
		t.Errorf("got %q", got)
	}
	if len(received.Messages) == 0 || received.Messages[0].Content != "prompt" {
		t.Errorf("prompt not forwarded")
	}
}

func TestChatCompletionNoChoices(t *testing.T) {
	client, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(openai.ChatCompletionResponse{}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	})
	defer srv.Close()

	got, err := bot.ChatCompletion(context.Background(), client, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestChatCompletionError(t *testing.T) {
	cfg := openai.DefaultConfig("test")
	cfg.HTTPClient = &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("boom")
	})}
	client := openai.NewClientWithConfig(cfg)

	_, err := bot.ChatCompletion(context.Background(), client, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
