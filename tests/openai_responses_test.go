package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	botpkg "telegram-reminder/internal/bot"
)

func TestResponsesCompletion(t *testing.T) {
	orig := botpkg.EnableWebSearch
	botpkg.EnableWebSearch = true
	defer func() { botpkg.EnableWebSearch = orig }()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req botpkg.ResponseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Model != "gpt-4.1" {
			t.Fatalf("model: %s", req.Model)
		}
		if !req.Stream {
			t.Fatalf("stream not enabled")
		}
		if len(req.Tools) == 0 || req.Tools[0].Type != "web_search" {
			t.Fatalf("unexpected tools: %+v", req.Tools)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("no flusher")
		}
		chunks := []string{
			`{"delta":{"output_text":"hello"}}`,
			`{"delta":{"output_text":" world"}}`,
		}
		for _, c := range chunks {
			_, _ = w.Write([]byte("event: thread.message.delta\n"))
			_, _ = w.Write([]byte("data: " + c + "\n\n"))
			flusher.Flush()
		}
		_, _ = w.Write([]byte("event: done\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	botpkg.ResponsesEndpoint = srv.URL + "/v1/responses"
	ctx := context.Background()
	out, err := botpkg.ResponsesCompletion(ctx, "test-key", "hi", "gpt-4.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Fatalf("unexpected output: %s", out)
	}
}
