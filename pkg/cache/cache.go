package cache

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_cache/cache.go -source=./cache.go -package=mockcache

// Cache represents a generic cache interface following the cache-aside pattern.
// Implementations should be safe for concurrent use.
type Cache[T any] interface {
	Get(ctx context.Context, key string) (T, bool)
	Set(ctx context.Context, key string, value T)
	Delete(ctx context.Context, key string)
}
