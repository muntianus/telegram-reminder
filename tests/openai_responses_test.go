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
		if len(req.Tools) == 0 || req.Tools[0].Type != "web_search" {
			t.Fatalf("unexpected tools: %+v", req.Tools)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"ok"}]}]}`))
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

func TestResponsesCompletionDelta(t *testing.T) {
	orig := botpkg.EnableWebSearch
	botpkg.EnableWebSearch = false
	defer func() { botpkg.EnableWebSearch = orig }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req botpkg.ResponseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"foo bar"}]}]}`))
	}))
	defer srv.Close()

	botpkg.ResponsesEndpoint = srv.URL + "/v1/responses"
	ctx := context.Background()
	out, err := botpkg.ResponsesCompletion(ctx, "key", "hi", "gpt-4.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "foo bar" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestGetDeleteCancelAndList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/responses/123", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"get"}]}]}`))
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","object":"response","deleted":true}`))
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})
	mux.HandleFunc("/v1/responses/123/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"cancel"}]}]}`))
	})
	mux.HandleFunc("/v1/responses/123/input_items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"msg1","type":"message","role":"user","content":[{"type":"input_text","text":"hi"}]}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	botpkg.ResponsesEndpoint = srv.URL + "/v1/responses"
	ctx := context.Background()
	out, err := botpkg.GetResponse(ctx, "k", "123")
	if err != nil || out != "get" {
		t.Fatalf("get: %v %s", err, out)
	}
	if err := botpkg.DeleteResponse(ctx, "k", "123"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	out, err = botpkg.CancelResponse(ctx, "k", "123")
	if err != nil || out != "cancel" {
		t.Fatalf("cancel: %v %s", err, out)
	}
	items, err := botpkg.ListInputItems(ctx, "k", "123")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 || len(items[0].Content) == 0 || items[0].Content[0].Text != "hi" {
		t.Fatalf("unexpected items: %+v", items)
	}
}
