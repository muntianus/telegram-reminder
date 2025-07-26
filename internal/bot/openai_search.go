package bot

import (
	"context"
	"errors"
	"os"
	"time"

	"telegram-reminder/internal/logger"
)

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

	out, err := ChatResponses(ctx, apiKey, CurrentModel, query)
	if err != nil {
		logger.L.Debug("openai search error", "err", err)
		return "", err
	}
	logger.L.Debug("openai search result", "bytes", len(out))
	return out, nil
}
