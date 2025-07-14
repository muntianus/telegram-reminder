// completion_test.go проверяет работу chatCompletion с поддельным сервером.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

// newTestClient создаёт клиента OpenAI поверх тестового сервера.
func newTestClient(handler http.HandlerFunc) (*openai.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := openai.DefaultConfig("test")
	cfg.BaseURL = srv.URL + "/"
	c := srv.Client()
	c.Timeout = openAITimeout
	cfg.HTTPClient = c
	return openai.NewClientWithConfig(cfg), srv
}

// TestChatCompletionSuccess проверяет успешный ответ от OpenAI.
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
	got, err := chatCompletion(context.Background(), client, []openai.ChatCompletionMessage{msg})
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

// TestChatCompletionNoChoices проверяет обработку пустого ответа.
func TestChatCompletionNoChoices(t *testing.T) {
	client, srv := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(openai.ChatCompletionResponse{}); err != nil {
			t.Fatalf("encode: %v", err)
		}
	})
	defer srv.Close()

	got, err := chatCompletion(context.Background(), client, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestChatCompletionError проверяет обработку ошибки запроса.
func TestChatCompletionError(t *testing.T) {
	cfg := openai.DefaultConfig("test")
	cfg.HTTPClient = &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("boom")
	})}
	client := openai.NewClientWithConfig(cfg)

	_, err := chatCompletion(context.Background(), client, []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "test"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

// roundTripperFunc реализует http.RoundTripper через функцию.
type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip вызывает вложенную функцию для выполнения запроса.
func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
