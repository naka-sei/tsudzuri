package time

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestWithTime(t *testing.T) {
	tests := []struct {
		name string
		base context.Context
		set  time.Time
	}{
		{
			name: "stores_time",
			base: context.Background(),
			set:  time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := WithTime(tt.base, tt.set)
			got := Now(ctx)
			if diff := cmp.Diff(tt.set, got); diff != "" {
				t.Fatalf("time mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNow(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type want struct {
		time.Time
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "with_time",
			args: args{
				ctx: WithTime(context.Background(), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			want: want{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Now(tt.args.ctx)
			if diff := cmp.Diff(tt.want.Time, got); diff != "" {
				t.Fatalf("time mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
