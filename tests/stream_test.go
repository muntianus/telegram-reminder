package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
)

func newStreamClient(responses []string) (*openai.Client, *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, data := range responses {
			fmt.Fprintf(w, "data: %s\n\n", data)
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
	})
	srv := httptest.NewServer(handler)
	cfg := openai.DefaultConfig("test")
	cfg.BaseURL = srv.URL + "/v1"
	cfg.HTTPClient = srv.Client()
	return openai.NewClientWithConfig(cfg), srv
}

func TestStreamChatCompletion(t *testing.T) {
	chunks := []string{
		`{"choices":[{"delta":{"content":"Hello "}}]}`,
		`{"choices":[{"delta":{"content":"world"}}]}`,
	}
	client, srv := newStreamClient(chunks)
	defer srv.Close()

	msg := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "hi"}
	ch, err := botpkg.StreamChatCompletion(context.Background(), client, []openai.ChatCompletionMessage{msg}, "gpt-4o")
	if err != nil {
		t.Fatalf("stream error: %v", err)
	}
	var out strings.Builder
	for part := range ch {
		out.WriteString(part)
	}
	if out.String() != "Hello world" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestStreamChatCompletionError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	cfg := openai.DefaultConfig("test")
	cfg.BaseURL = srv.URL + "/v1"
	cfg.HTTPClient = srv.Client()
	client := openai.NewClientWithConfig(cfg)

	msg := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: "hi"}
	_, err := botpkg.StreamChatCompletion(context.Background(), client, []openai.ChatCompletionMessage{msg}, "gpt-4o")
	if err == nil {
		t.Fatal("expected error")
	}
}
