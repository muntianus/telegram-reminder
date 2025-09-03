package bot

import (
	"sync"
	"time"
)

// memoryCache implements SearchCache using in-memory storage with LRU eviction
type memoryCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	config  CacheConfig
}

type cacheEntry struct {
	result   string
	created  time.Time
	accessed time.Time
}

// NewMemoryCache creates a new in-memory search cache
func NewMemoryCache(config CacheConfig) SearchCache {
	return &memoryCache{
		entries: make(map[string]cacheEntry),
		config:  config,
	}
}

// Get retrieves a cached search result
func (c *memoryCache) Get(query string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[query]
	if !exists {
		return "", false
	}

	// Check if entry is expired
	if time.Since(entry.created) > c.config.TTL {
		delete(c.entries, query)
		return "", false
	}

	// Update access time for LRU
	entry.accessed = time.Now()
	c.entries[query] = entry

	return entry.result, true
}

// Set stores a search result in the cache
func (c *memoryCache) Set(query, result string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	entry := cacheEntry{
		result:   result,
		created:  now,
		accessed: now,
	}

	// Clean expired entries first
	c.cleanExpired()

	// If at capacity, remove LRU entry
	if len(c.entries) >= c.config.MaxSize {
		c.removeLRU()
	}

	c.entries[query] = entry
}

// Clear removes all cached entries
func (c *memoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

// cleanExpired removes expired entries (called with lock held)
func (c *memoryCache) cleanExpired() {
	now := time.Now()
	for query, entry := range c.entries {
		if now.Sub(entry.created) > c.config.TTL {
			delete(c.entries, query)
		}
	}
}

// removeLRU removes the least recently used entry (called with lock held)
func (c *memoryCache) removeLRU() {
	if len(c.entries) == 0 {
		return
	}

	var oldestQuery string
	var oldestTime time.Time = time.Now()

	for query, entry := range c.entries {
		if entry.accessed.Before(oldestTime) {
			oldestTime = entry.accessed
			oldestQuery = query
		}
	}

	if oldestQuery != "" {
		delete(c.entries, oldestQuery)
	}
}
