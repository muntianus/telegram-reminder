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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req botpkg.ResponseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Model != "gpt-4.1" {
			t.Fatalf("model: %s", req.Model)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output_text":"ok"}`))
	}))
	defer srv.Close()

	botpkg.ResponsesEndpoint = srv.URL + "/v1/responses"
	ctx := context.Background()
	out, err := botpkg.ResponsesCompletion(ctx, "test-key", "hi", "gpt-4.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("unexpected output: %s", out)
	}
}
