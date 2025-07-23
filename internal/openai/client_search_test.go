package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch_CallsWebSearchTool(t *testing.T) {
	var reqBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		reqBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"1","created_at":0,"error":{},"incomplete_details":{},"instructions":{},"metadata":{},"model":"gpt-4o-mini","object":"response","output":[{"content":[{"type":"output_text","text":"ok"}],"type":"message"}],"parallel_tool_calls":false,"temperature":0,"tool_choice":{},"tools":[]}`))
	}))
	defer srv.Close()

	c := New(Config{APIKey: "sk-test", ModelSearch: "gpt-4o-mini", BaseURL: srv.URL})
	_, err := c.Search(context.Background(), "hello")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	var data struct {
		Model string `json:"model"`
		Tools []struct {
			Type string `json:"type"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(reqBody, &data); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if data.Model != "gpt-4o-mini" {
		t.Errorf("model %s", data.Model)
	}
	if len(data.Tools) == 0 || data.Tools[0].Type == "" {
		t.Fatalf("no tool in request")
	}
	if data.Tools[0].Type != "web_search_preview" && data.Tools[0].Type != "web_search" {
		t.Errorf("unexpected tool %s", data.Tools[0].Type)
	}
}
