package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	t.Parallel()

	type args struct {
		ttl     time.Duration
		sleep   time.Duration
		prepare func(context.Context, *MemoryCache[string])
	}

	type want struct {
		value  string
		exists bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "hit_after_set",
			args: args{
				ttl: time.Minute,
				prepare: func(ctx context.Context, c *MemoryCache[string]) {
					c.Set(ctx, "key", "value")
				},
			},
			want: want{value: "value", exists: true},
		},
		{
			name: "miss_after_delete",
			args: args{
				ttl: time.Minute,
				prepare: func(ctx context.Context, c *MemoryCache[string]) {
					c.Set(ctx, "key", "value")
					c.Delete(ctx, "key")
				},
			},
			want: want{value: "", exists: false},
		},
		{
			name: "miss_after_expiration",
			args: args{
				ttl:   10 * time.Millisecond,
				sleep: 25 * time.Millisecond,
				prepare: func(ctx context.Context, c *MemoryCache[string]) {
					c.Set(ctx, "key", "value")
				},
			},
			want: want{value: "", exists: false},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cache := NewMemoryCache[string](tt.args.ttl)
			testCtx := context.Background()
			if tt.args.prepare != nil {
				tt.args.prepare(testCtx, cache)
			}

			if tt.args.sleep > 0 {
				time.Sleep(tt.args.sleep)
			}

			got, ok := cache.Get(testCtx, "key")
			if got != tt.want.value || ok != tt.want.exists {
				t.Fatalf("cache result mismatch got=(%q,%t) want=(%q,%t)", got, ok, tt.want.value, tt.want.exists)
			}
		})
	}
}
