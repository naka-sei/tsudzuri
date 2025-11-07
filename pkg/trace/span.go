package trace

import (
	"context"

	"go.opentelemetry.io/otel"
)

// StartSpan starts a new span with the given name.
// caller can terminate the span by executing `defer func()`.
func StartSpan(ctx context.Context, name string) (context.Context, func()) {
	ctx, span := otel.Tracer(name).Start(ctx, name)
	return ctx, func() {
		span.End()
	}
}
