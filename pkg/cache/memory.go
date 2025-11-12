package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache is an in-memory cache implementation that supports automatic expiry.
type MemoryCache[T any] struct {
	mu    sync.RWMutex
	items map[string]entry[T]
	ttl   time.Duration
}

type entry[T any] struct {
	value     T
	expiresAt time.Time
}

// NewMemoryCache creates a new MemoryCache with the provided TTL.
// If ttl is zero or negative, the entries never expire.
func NewMemoryCache[T any](ttl time.Duration) *MemoryCache[T] {
	return &MemoryCache[T]{
		items: make(map[string]entry[T]),
		ttl:   ttl,
	}
}

// Get retrieves a value from the cache. The second return value indicates whether the key exists.
func (c *MemoryCache[T]) Get(ctx context.Context, key string) (T, bool) {
	_ = ctx
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()

	var zero T
	if !ok {
		return zero, false
	}

	if c.ttl > 0 && time.Now().After(e.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return zero, false
	}

	return e.value, true
}

// Set stores a value in the cache, overriding any existing value for the key.
func (c *MemoryCache[T]) Set(ctx context.Context, key string, value T) {
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := entry[T]{
		value: value,
	}
	if c.ttl > 0 {
		entry.expiresAt = time.Now().Add(c.ttl)
	}

	c.items[key] = entry
}

// Delete removes a value from the cache.
func (c *MemoryCache[T]) Delete(ctx context.Context, key string) {
	_ = ctx
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
