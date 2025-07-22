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
	"strconv"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// SearchResult represents a single search result from OpenAI.
type SearchResult struct {
	Document string
	Score    float64
	Text     string
}

// searchResponse mirrors the structure of OpenAI search API response.
type searchResponse struct {
	Data []struct {
		Document int     `json:"document"`
		Score    float64 `json:"score"`
		Text     string  `json:"text"`
	} `json:"data"`
}

// OpenAISearch queries the OpenAI Search API and returns ranked documents.
func OpenAISearch(query string) ([]SearchResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY not set")
	}

	reqBody := map[string]any{
		"query": query,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := "https://api.openai.com/v1/engines/davinci/search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var errResp openai.ErrorResponse
		b, _ := io.ReadAll(resp.Body)
		if json.Unmarshal(b, &errResp) == nil && errResp.Error != nil {
			errResp.Error.HTTPStatus = resp.Status
			errResp.Error.HTTPStatusCode = resp.StatusCode
			return nil, errResp.Error
		}
		return nil, &openai.RequestError{
			HTTPStatus:     resp.Status,
			HTTPStatusCode: resp.StatusCode,
			Err:            fmt.Errorf("unexpected status"),
			Body:           b,
		}
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	results := make([]SearchResult, len(sr.Data))
	for i, d := range sr.Data {
		results[i] = SearchResult{
			Document: strconv.Itoa(d.Document),
			Score:    d.Score,
			Text:     d.Text,
		}
	}
	return results, nil
}
