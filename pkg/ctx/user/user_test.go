package user

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	duser "github.com/naka-sei/tsudzuri/domain/user"
)

func TestWithUser(t *testing.T) {
	tests := []struct {
		name string
		base context.Context
		set  *duser.User
	}{
		{
			name: "stores_user",
			base: context.Background(),
			set:  duser.NewUser("test-uid"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := WithUser(tt.base, tt.set)
			got, _ := UserFromContext(ctx)
			if diff := cmp.Diff(tt.set, got, cmpopts.IgnoreUnexported(duser.User{})); diff != "" {
				t.Fatalf("user mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUserFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type want struct {
		user *duser.User
		ok   bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "with_user",
			args: args{
				ctx: WithUser(context.Background(), duser.NewUser("test-uid")),
			},
			want: want{
				user: duser.NewUser("test-uid"),
				ok:   true,
			},
		},
		{
			name: "without_user",
			args: args{
				ctx: context.Background(),
			},
			want: want{
				user: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := UserFromContext(tt.args.ctx)
			if ok != tt.want.ok {
				t.Fatalf("ok mismatch: want %v got %v", tt.want.ok, ok)
			}
			if diff := cmp.Diff(tt.want.user, got, cmpopts.IgnoreUnexported(duser.User{})); diff != "" {
				t.Fatalf("user mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
