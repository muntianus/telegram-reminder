package bot

import (
	"context"

	"telegram-reminder/internal/logger"
)

// OpenAISearch performs a web search using the OpenAI responses API and returns
// the result formatted for Telegram HTML.
func OpenAISearch(query string) (string, error) {
	logger.L.Debug("openai search", "query", query)
	ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
	defer cancel()

	out, err := defaultWebSearch(ctx, query)
	if err != nil {
		logger.L.Debug("openai search error", "err", err)
		return "", err
	}
	logger.L.Debug("openai search result", "bytes", len(out))
	return out, nil
}
