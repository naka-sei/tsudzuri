package cache

import "context"

// Cache represents a generic cache interface following the cache-aside pattern.
// Implementations should be safe for concurrent use.
type Cache[T any] interface {
	Get(ctx context.Context, key string) (T, bool)
	Set(ctx context.Context, key string, value T)
	Delete(ctx context.Context, key string)
}
