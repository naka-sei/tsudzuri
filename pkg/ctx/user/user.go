package user

import (
	"context"

	duser "github.com/naka-sei/tsudzuri/domain/user"
)

// userCtxKey is the context key for user.
type userCtxKey struct{}

// WithUser adds a user to the context.
func WithUser(ctx context.Context, user *duser.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

// UserFromContext retrieves the user from the context.
func UserFromContext(ctx context.Context) (*duser.User, bool) {
	v := ctx.Value(userCtxKey{})
	if v == nil {
		return nil, false
	}
	if u, ok := v.(*duser.User); ok {
		return u, true
	}
	return nil, false
}
