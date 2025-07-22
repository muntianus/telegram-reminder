package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
type responseResult struct {
	OutputText string `json:"output_text"`
}

// callResponsesAPI performs a request to the given responses endpoint.
func callResponsesAPI(ctx context.Context, apiKey string, reqBody ResponseRequest, endpoint string) (string, error) {
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

	client := &http.Client{Timeout: OpenAITimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s", strings.TrimSpace(string(data)))
	}
	var res responseResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return strings.TrimSpace(res.OutputText), nil
}

// ResponsesCompletion sends input to the OpenAI responses API and returns output text.
func ResponsesCompletion(ctx context.Context, apiKey, input, model string) (string, error) {
	req := ResponseRequest{
		Model: model,
		Input: input,
	}
	if EnableWebSearch && supportsWebSearch(model) {
		req.Tools = []ResponseTool{{Type: "web_search_preview"}}
	}
	return callResponsesAPI(ctx, apiKey, req, "")
}
