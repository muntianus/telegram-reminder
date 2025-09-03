package bot

import (
	"context"
	"os"
	"sync"
	"time"
)

var (
	globalSearchService *SearchService
	searchServiceOnce   sync.Once
)

// getSearchService returns the global search service instance
func getSearchService() *SearchService {
	searchServiceOnce.Do(func() {
		// Create cache config
		cacheConfig := CacheConfig{
			TTL:     10 * time.Minute,
			MaxSize: 100,
		}

		// Create cache
		cache := NewMemoryCache(cacheConfig)

		// Create provider - prefer Responses API for direct search
		apiKey := os.Getenv("OPENAI_API_KEY")
		var provider SearchProvider
		if apiKey != "" {
			provider = NewResponsesSearchProvider(apiKey)
		} else {
			// Fallback to chat completion provider if no API key
			provider = NewChatCompletionSearchProvider(nil)
		}

		globalSearchService = NewSearchService(provider, cache, cacheConfig)
	})

	return globalSearchService
}

// OpenAISearch performs a web search using the configured search service and returns
// the result formatted for Telegram HTML.
func OpenAISearch(query string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OpenAITimeout)
	defer cancel()

	searchService := getSearchService()
	out, err := searchService.Search(ctx, query)
	if err != nil {
		return "", err
	}

	out = markdownToTelegramHTML(out)
	return out, nil
}
