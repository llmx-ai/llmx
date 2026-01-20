package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/llmx-ai/llmx"
)

// Cache interface for response caching
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
}

// MemoryCache implements an in-memory cache
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// Set stores a value in cache
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Delete removes a value from cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// cleanup removes expired items
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// CacheMiddleware creates a caching middleware
func CacheMiddleware(cache Cache, ttl time.Duration) Middleware {
	if cache == nil {
		cache = NewMemoryCache()
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, req *llmx.ChatRequest) (*llmx.ChatResponse, error) {
			// Generate cache key
			key := generateCacheKey(req)

			// Check cache
			if cached, ok := cache.Get(key); ok {
				if resp, ok := cached.(*llmx.ChatResponse); ok {
					return resp, nil
				}
			}

			// Execute request
			resp, err := next(ctx, req)
			if err != nil {
				return nil, err
			}

			// Store in cache
			cache.Set(key, resp, ttl)

			return resp, nil
		}
	}
}

// generateCacheKey creates a cache key from request
func generateCacheKey(req *llmx.ChatRequest) string {
	// Serialize request
	data, _ := json.Marshal(req)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}
