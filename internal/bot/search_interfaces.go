package bot

import (
	"context"
	"time"
)

// SearchProvider defines the interface for web search providers
type SearchProvider interface {
	Search(ctx context.Context, query string) (string, error)
	SupportsModel(model string) bool
}

// CacheConfig holds configuration for search caching
type CacheConfig struct {
	TTL      time.Duration
	MaxSize  int
	Disabled bool
}

// SearchCache defines the interface for search result caching
type SearchCache interface {
	Get(query string) (string, bool)
	Set(query, result string)
	Clear()
}

// SearchService provides web search functionality with caching
type SearchService struct {
	provider SearchProvider
	cache    SearchCache
	config   CacheConfig
}

// NewSearchService creates a new search service
func NewSearchService(provider SearchProvider, cache SearchCache, config CacheConfig) *SearchService {
	return &SearchService{
		provider: provider,
		cache:    cache,
		config:   config,
	}
}

// Search performs a web search with caching
func (s *SearchService) Search(ctx context.Context, query string) (string, error) {
	normalized := normalizeQuery(query)

	// Check cache first if enabled
	if !s.config.Disabled {
		if result, found := s.cache.Get(normalized); found {
			return result, nil
		}
	}

	// Perform search
	result, err := s.provider.Search(ctx, normalized)
	if err != nil {
		return "", err
	}

	// Cache result if enabled
	if !s.config.Disabled && result != "" {
		s.cache.Set(normalized, result)
	}

	return result, nil
}

// SupportsModel checks if the search provider supports the given model
func (s *SearchService) SupportsModel(model string) bool {
	return s.provider.SupportsModel(model)
}
