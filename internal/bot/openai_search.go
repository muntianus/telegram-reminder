package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"telegram-reminder/internal/logger"
)

// markdownToTelegramHTML converts a subset of Markdown to Telegram-compatible HTML.
func markdownToTelegramHTML(input string) string {
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	input = reBold.ReplaceAllString(input, "<b>$1</b>")

	reLink := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	input = reLink.ReplaceAllString(input, `<a href="$2">$1</a>`)

	return input
}

// OpenAISearch performs a web search using the OpenAI responses API and returns
// the result formatted for Telegram HTML.
func OpenAISearch(query string) (string, error) {
	logger.L.Debug("openai search", "query", query)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"model": CurrentModel,
		"input": []map[string]string{{"role": "user", "content": query}},
		"tools": []map[string]any{
			{
				"type": "function",
				"function": map[string]any{
					"name":        "web_search",
					"description": "Search the web for information",
					"parameters": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]string{
								"type":        "string",
								"description": "Search query",
							},
						},
						"required": []string{"query"},
					},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ResponsesEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := logger.NewHTTPClient(OpenAITimeout)
	resp, err := client.Do(req)
	if err != nil {
		logger.L.Debug("openai search error", "err", err)
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
		logger.L.Debug("openai search status", "status", resp.Status)
		return "", err
	}

	var res struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.L.Debug("openai search decode", "err", err)
		return "", err
	}
	if len(res.Output) < 2 || len(res.Output[1].Content) == 0 {
		return "", errors.New("openai: empty response")
	}

	out := markdownToTelegramHTML(res.Output[1].Content[0].Text)
	logger.L.Debug("openai search result", "bytes", len(out))
	return out, nil
}
