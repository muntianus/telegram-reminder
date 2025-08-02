package bot

import (
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
	Model string         `json:"model"`
	Tools []ResponseTool `json:"tools,omitempty"`
	Input string         `json:"input"`
}

type ResponseTool struct {
	Type string `json:"type"`
}

// responseResult is the minimal response structure we care about.
// responseContent represents a single content part in the Responses API output.
type responseContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// responseOutput represents an output item returned by the Responses API.
type responseOutput struct {
	Type    string            `json:"type"`
	Content []responseContent `json:"content"`
}

// responseResult contains only the fields we need from the Responses API.
type responseResult struct {
	Output []responseOutput `json:"output"`
}

// extractOutputText concatenates all text parts from the Responses API output.
func extractOutputText(res responseResult) string {
	var parts []string
	for _, out := range res.Output {
		for _, c := range out.Content {
			if t := strings.TrimSpace(c.Text); t != "" {
				parts = append(parts, t)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.L.Debug("failed to close response body", "err", err)
		}
	}()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		logger.L.Debug("responses api status", "status", resp.Status, "body", string(data))
		err := fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
		return "", err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.L.Debug("responses api read", "err", err)
		return "", err
	}
	logger.L.Debug("responses api raw", "body", string(data))
	var res responseResult
	if err := json.Unmarshal(data, &res); err != nil {
		logger.L.Debug("responses api decode", "err", err)
		logger.L.Debug("responses api body", "body", string(data))
		return "", err
	}
	out := extractOutputText(res)
	if out == "" {
		logger.L.Debug("responses api empty output", "body", string(data))
		return "", errors.New("openai: empty response")
	}
	logger.L.Debug("responses api result", "bytes", len(out), "preview", logger.Truncate(out, 200))
	return out, nil
}

// ResponsesCompletion sends input to the OpenAI responses API and returns output text.
func ResponsesCompletion(ctx context.Context, apiKey, input, model string) (string, error) {
	req := ResponseRequest{
		Model: model,
		Input: input,
	}
	if getRuntimeConfig().EnableWebSearch && supportsWebSearch(model) {
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
	req := ResponseRequest{
		Model: model,
		Input: prompt,
		Tools: []ResponseTool{{Type: "web_search"}},
	}
	out, err := callResponsesAPI(ctx, apiKey, req, "")
	if err != nil {
		return "", err
	}
	out = markdownToTelegramHTML(out)
	return out, nil
}

// GetResponse fetches a previously created model response by ID and returns the
// output text.
func GetResponse(ctx context.Context, apiKey, id string) (string, error) {
	url := fmt.Sprintf("%s/%s", ResponsesEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	client := logger.NewHTTPClient(OpenAITimeout)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.L.Debug("failed to close response body", "err", err)
		}
	}()
	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var res responseResult
	if err := json.Unmarshal(data, &res); err != nil {
		return "", err
	}
	out := extractOutputText(res)
	if out == "" {
		return "", errors.New("openai: empty response")
	}
	return out, nil
}

// DeleteResponse deletes the response with the given ID.
func DeleteResponse(ctx context.Context, apiKey, id string) error {
	url := fmt.Sprintf("%s/%s", ResponsesEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	client := logger.NewHTTPClient(OpenAITimeout)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.L.Debug("failed to close response body", "err", err)
		}
	}()
	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
	}
	return nil
}

// CancelResponse cancels a background response generation.
func CancelResponse(ctx context.Context, apiKey, id string) (string, error) {
	url := fmt.Sprintf("%s/%s/cancel", ResponsesEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	client := logger.NewHTTPClient(OpenAITimeout)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.L.Debug("failed to close response body", "err", err)
		}
	}()
	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var res responseResult
	if err := json.Unmarshal(data, &res); err != nil {
		return "", err
	}
	out := extractOutputText(res)
	if out == "" {
		return "", errors.New("openai: empty response")
	}
	return out, nil
}

// InputItem represents an item used to generate a response.
type InputItem struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Role    string            `json:"role"`
	Content []responseContent `json:"content"`
}

// ListInputItems returns the list of input items for a response ID.
func ListInputItems(ctx context.Context, apiKey, id string) ([]InputItem, error) {
	url := fmt.Sprintf("%s/%s/input_items", ResponsesEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	client := logger.NewHTTPClient(OpenAITimeout)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.L.Debug("failed to close response body", "err", err)
		}
	}()
	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res struct {
		Data []InputItem `json:"data"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
