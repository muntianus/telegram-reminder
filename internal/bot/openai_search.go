package bot

import (
	"context"

	"telegram-reminder/internal/logger"
)

// OpenAISearch performs a web search using the OpenAI responses API and returns
// the result formatted for Telegram HTML.
func OpenAISearch(query string) (string, error) {
	// Debug logging removed to prevent Telegram spam
	ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
	defer cancel()

	out, err := defaultWebSearch(ctx, query)
	if err != nil {
		return "", err
	}
	out = markdownToTelegramHTML(out)
	return out, nil
}
