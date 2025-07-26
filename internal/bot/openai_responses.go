package bot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"telegram-reminder/internal/logger"
)

// ResponsesEndpoint defines the OpenAI API endpoint for responses. Tests can override it.
var ResponsesEndpoint = "https://api.openai.com/v1/responses"

// ResponseRequest is the payload for the /v1/responses endpoint.
type ResponseRequest struct {
	Model  string         `json:"model"`
	Tools  []ResponseTool `json:"tools,omitempty"`
	Input  string         `json:"input"`
	Stream bool           `json:"stream"`
}

type ResponseTool struct {
	Type string `json:"type"`
}

// responseResult is the minimal response structure we care about.
type responseResult struct {
	OutputText string `json:"output_text"`
}

// callResponsesAPI performs a request to the given responses endpoint.
func callResponsesAPI(ctx context.Context, apiKey string, reqBody ResponseRequest, endpoint string) (string, error) {
	logger.L.Debug("responses api", "model", reqBody.Model)
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	if endpoint == "" {
		endpoint = ResponsesEndpoint
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	client := logger.NewHTTPClient(OpenAITimeout)
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.L.Debug("responses api error", "err", err)
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		logger.L.Debug("responses api status", "status", resp.Status, "body", string(data))
		err := fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
		return "", err
	}
	if reqBody.Stream {
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
		var buf strings.Builder
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || line == "data: [DONE]" {
				continue
			}
			line = strings.TrimPrefix(line, "data:")
			var chunk responseResult
			if err := json.Unmarshal([]byte(line), &chunk); err == nil {
				buf.WriteString(chunk.OutputText)
			}
		}
		if err := scanner.Err(); err != nil {
			logger.L.Debug("responses api read", "err", err)
			return "", err
		}
		out := strings.TrimSpace(buf.String())
		if out == "" {
			logger.L.Debug("responses api empty output")
			return "", errors.New("openai: empty response")
		}
		logger.L.Debug("responses api result", "bytes", len(out))
		return out, nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.L.Debug("responses api read", "err", err)
		return "", err
	}
	var res responseResult
	if err := json.Unmarshal(data, &res); err != nil {
		logger.L.Debug("responses api decode", "err", err)
		logger.L.Debug("responses api body", "body", string(data))
		return "", err
	}
	out := strings.TrimSpace(res.OutputText)
	if out == "" {
		logger.L.Debug("responses api empty output", "body", string(data))
		return "", errors.New("openai: empty response")
	}
	logger.L.Debug("responses api result", "bytes", len(out))
	return out, nil
}

// ResponsesCompletion sends input to the OpenAI responses API and returns output text.
func ResponsesCompletion(ctx context.Context, apiKey, input, model string) (string, error) {
	req := ResponseRequest{
		Model:  model,
		Input:  input,
		Stream: true,
	}
	if EnableWebSearch && supportsWebSearch(model) {
		req.Tools = []ResponseTool{{Type: "web_search"}}
	}
	return callResponsesAPI(ctx, apiKey, req, "")
}

// markdownToTelegramHTML converts a subset of Markdown to Telegram-compatible HTML.
func markdownToTelegramHTML(input string) string {
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	input = reBold.ReplaceAllString(input, "<b>$1</b>")

	reLink := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	input = reLink.ReplaceAllString(input, `<a href="$2">$1</a>`)

	return input
}

// ChatResponses sends a prompt to the OpenAI responses API using a virtual
// web_search function and returns the result formatted for Telegram HTML.
func ChatResponses(ctx context.Context, apiKey, model, prompt string) (string, error) {
	reqBody := map[string]any{
		"model":  model,
		"input":  []map[string]string{{"role": "user", "content": prompt}},
		"tools":  []map[string]any{{"type": "web_search"}},
		"stream": true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ResponsesEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	client := logger.NewHTTPClient(OpenAITimeout)
	resp, err := client.Do(httpReq)
	if err != nil {
		logger.L.Debug("responses api error", "err", err)
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
		logger.L.Debug("responses api status", "status", resp.Status)
		return "", err
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	var buf strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "data: [DONE]" {
			continue
		}
		line = strings.TrimPrefix(line, "data:")
		var res struct {
			Output []struct {
				Type    string `json:"type"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"output"`
		}
		if err := json.Unmarshal([]byte(line), &res); err == nil {
			if len(res.Output) > 0 && len(res.Output[len(res.Output)-1].Content) > 0 {
				buf.WriteString(res.Output[len(res.Output)-1].Content[0].Text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.L.Debug("responses api read", "err", err)
		return "", err
	}
	out := strings.TrimSpace(buf.String())
	if out == "" {
		return "", errors.New("openai: empty response")
	}
	out = markdownToTelegramHTML(out)
	logger.L.Debug("responses api result", "bytes", len(out))
	return out, nil
}
