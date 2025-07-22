package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	botpkg "telegram-reminder/internal/bot"

	openai "github.com/sashabaranov/go-openai"
)

// helper to create openai client backed by test server
func newOAITestClient(handler http.HandlerFunc) (*openai.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := openai.DefaultConfig("test")
	cfg.BaseURL = srv.URL + "/"
	c := srv.Client()
	cfg.HTTPClient = c
	return openai.NewClientWithConfig(cfg), srv
}

func TestExtractSearchQueries(t *testing.T) {
	prompt := "intro\n🔍 ВЕБ-ПОИСК: Найди новости\n- \"query one\"\n- \"query two\"\nПОЛЕЗНЫЕ ССЫЛКИ:"
	got := botpkg.ExtractSearchQueries(prompt)
	if len(got) != 2 || got[0] != "query one" || got[1] != "query two" {
		t.Fatalf("unexpected queries: %v", got)
	}

	if q := botpkg.ExtractSearchQueries("no search section"); len(q) != 0 {
		t.Fatalf("expected none, got %v", q)
	}
}

func TestEnhancedSystemCompletionAddsResults(t *testing.T) {
	var req openai.ChatCompletionRequest
	client, srv := newOAITestClient(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "ok"}}}})
	})
	defer srv.Close()

	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"RelatedTopics": []map[string]any{
				{"Text": "Result text", "FirstURL": "https://example.com"},
			},
		})
	}))
	botpkg.SearchHTTPClient = searchSrv.Client()
	botpkg.SearchBaseURL = searchSrv.URL + "/"
	defer func() {
		botpkg.SearchHTTPClient = http.DefaultClient
		botpkg.SearchBaseURL = "https://duckduckgo.com/"
	}()
	defer searchSrv.Close()

	prompt := "test\n🔍 ВЕБ-ПОИСК: Найди новости\n- \"query\""
	_, err := botpkg.EnhancedSystemCompletion(context.Background(), client, prompt, "o3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := req.Messages[0].Content
	if !strings.Contains(content, "Result text") {
		t.Errorf("search results not appended: %s", content)
	}
}

func TestEnhancedSystemCompletionSearchError(t *testing.T) {
	var req openai.ChatCompletionRequest
	client, srv := newOAITestClient(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "ok"}}}})
	})
	defer srv.Close()

	botpkg.SearchHTTPClient = &http.Client{Transport: searchRoundTripperFunc(func(*http.Request) (*http.Response, error) { return nil, http.ErrHandlerTimeout })}
	botpkg.SearchBaseURL = "https://example.com/"
	defer func() {
		botpkg.SearchHTTPClient = http.DefaultClient
		botpkg.SearchBaseURL = "https://duckduckgo.com/"
	}()

	prompt := "test\n🔍 ВЕБ-ПОИСК: Найди новости\n- \"query\""
	_, err := botpkg.EnhancedSystemCompletion(context.Background(), client, prompt, "o3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(req.Messages[0].Content, "query") && strings.Contains(req.Messages[0].Content, "https://") {
		t.Errorf("search results should not be appended on error")
	}
}

func TestEnhancedSystemCompletionNoResults(t *testing.T) {
	var req openai.ChatCompletionRequest
	client, srv := newOAITestClient(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{Content: "ok"}}}})
	})
	defer srv.Close()

	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"RelatedTopics": []any{}})
	}))
	botpkg.SearchHTTPClient = searchSrv.Client()
	botpkg.SearchBaseURL = searchSrv.URL + "/"
	defer func() {
		botpkg.SearchHTTPClient = http.DefaultClient
		botpkg.SearchBaseURL = "https://duckduckgo.com/"
	}()
	defer searchSrv.Close()

	prompt := "test\n🔍 ВЕБ-ПОИСК: Найди новости\n- \"query\""
	_, err := botpkg.EnhancedSystemCompletion(context.Background(), client, prompt, "o3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(req.Messages[0].Content, "Result text") {
		t.Errorf("expected no appended results")
	}
}

type searchRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f searchRoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
