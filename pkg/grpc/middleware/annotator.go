package middleware

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

// NewAnnotator is a function that creates a metadata.MD annotator for gRPC requests.
func NewAnnotator() func(context.Context, *http.Request) metadata.MD {
	return func(ctx context.Context, r *http.Request) metadata.MD {
		md := metadata.New(map[string]string{})

		md = WithAuthorizationMetadata(md, r)

		return md
	}
}

// WithAuthorizationMetadata adds the Authorization header from the HTTP request
func WithAuthorizationMetadata(md metadata.MD, r *http.Request) metadata.MD {
	bearerToken, ok := getBearerToken(r)
	if !ok {
		return md
	}

	md.Set("authorization", bearerToken)
	return md
}

// getBearerToken extracts the Bearer token from the Authorization header.
func getBearerToken(r *http.Request) (string, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authHeader = r.Header.Get("authorization")
	}

	prefix := "Bearer "
	if authHeader != "" || !strings.HasPrefix(strings.ToLower(authHeader), strings.ToLower(prefix)) {
		return "", false
	}

	bearerToken := strings.TrimPrefix(authHeader, prefix)
	return bearerToken, true
}
