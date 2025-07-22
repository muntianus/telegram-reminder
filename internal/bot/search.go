package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HTTPDoer defines minimal http client interface used by WebSearch.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// SearchHTTPClient is the HTTP client used for web search. It can be overridden in tests.
var (
	SearchHTTPClient HTTPDoer = http.DefaultClient
	SearchBaseURL             = "https://duckduckgo.com/"
)

// SearchResult represents a single web search result.
type SearchResult struct {
	Title string
	URL   string
}

// WebSearch performs a DuckDuckGo search and returns basic results.
func WebSearch(ctx context.Context, query string) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}
	u, _ := url.Parse(SearchBaseURL)
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("no_redirect", "1")
	q.Set("no_html", "1")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := SearchHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data struct {
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
			Topics   []struct {
				Text     string `json:"Text"`
				FirstURL string `json:"FirstURL"`
			} `json:"Topics"`
		} `json:"RelatedTopics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, t := range data.RelatedTopics {
		if t.Text != "" && t.FirstURL != "" {
			results = append(results, SearchResult{Title: t.Text, URL: t.FirstURL})
		}
		for _, sub := range t.Topics {
			if sub.Text != "" && sub.FirstURL != "" {
				results = append(results, SearchResult{Title: sub.Text, URL: sub.FirstURL})
			}
		}
	}
	return results, nil
}

// ExtractSearchQueries parses the prompt and returns search queries listed after the
// "ВЕБ-ПОИСК" section.
func ExtractSearchQueries(prompt string) []string {
	var queries []string
	lines := strings.Split(prompt, "\n")
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "ВЕБ-ПОИСК") {
			inBlock = true
			continue
		}
		if inBlock {
			if strings.HasPrefix(trimmed, "-") {
				q := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
				q = strings.Trim(q, "\"")
				if q != "" {
					queries = append(queries, q)
				}
			} else if trimmed == "" || !strings.HasPrefix(trimmed, "-") {
				break
			}
		}
	}
	return queries
}
