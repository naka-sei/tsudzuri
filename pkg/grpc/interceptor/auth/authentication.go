package auth

import (
	"context"
	"strings"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/api/firebase"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/presentation/errcode"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var publicMethods = map[string]struct{}{
	"/tsudzuri.v1.UserService/CreateUser": {},
}

func NewAuthenticationUnaryServerInterceptor(
	authenticator firebase.Authenticator,
	userRepo duser.UserRepository,
	userCache cache.Cache[*duser.User],
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := publicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errcode.ToGRPCStatus(duser.ErrUserNotFound)
		}

		idToken := extractAuthorization(md)
		if idToken == "" {
			return nil, errcode.ToGRPCStatus(duser.ErrUserNotFound)
		}

		token, err := authenticator.VerifyIDToken(ctx, idToken)
		if err != nil {
			return nil, errcode.ToGRPCStatus(duser.ErrUserNotFound)
		}

		if userCache != nil {
			if cachedUser, ok := userCache.Get(ctx, token.UID); ok {
				return handler(ctxuser.WithUser(ctx, cachedUser), req)
			}
		}

		user, err := userRepo.Get(ctx, token.UID)
		if err != nil {
			return nil, errcode.ToGRPCStatus(err)
		}
		if user == nil {
			return nil, errcode.ToGRPCStatus(duser.ErrUserNotFound)
		}

		if userCache != nil {
			userCache.Set(ctx, token.UID, user)
		}

		return handler(ctxuser.WithUser(ctx, user), req)
	}
}

func extractAuthorization(md metadata.MD) string {
	values := md.Get("authorization")
	if len(values) == 0 {
		values = md.Get("Authorization")
	}
	if len(values) == 0 {
		return ""
	}

	token := strings.TrimSpace(values[0])
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return strings.TrimSpace(token[7:])
	}
	return token
}
