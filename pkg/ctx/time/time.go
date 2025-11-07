package time

import (
	"context"
	"time"
)

// timeCtxKey is the context key for storing time values.
type timeCtxKey struct{}

// WithTime returns a new context with the given time value.
func WithTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, timeCtxKey{}, t)
}

// Now retrieves the time value from the context.
func Now(ctx context.Context) time.Time {
	if t, ok := ctx.Value(timeCtxKey{}).(time.Time); ok {
		return t
	}
	return time.Now()
}
