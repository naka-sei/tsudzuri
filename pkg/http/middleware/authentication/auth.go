package authentication

import (
	"net/http"
	"slices"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/api/firebase"
	"github.com/naka-sei/tsudzuri/pkg/cache"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	merr "github.com/naka-sei/tsudzuri/pkg/http/middleware/error"
)

var publicRouteMap = map[string][]string{
	"POST": {"/api/v1/users"},
}

// AuthHTTPMiddleware is the HTTP middleware for authentication.
func AuthHTTPMiddleware(authenticator firebase.Authenticator, userRepo duser.UserRepository, userCache cache.Cache[*duser.User]) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			publicRoutes, ok := publicRouteMap[r.Method]
			if !ok || !slices.Contains(publicRoutes, r.URL.Path) {
				idToken := r.Header.Get("Authorization")
				if idToken == "" {
					merr.WriteError(w, duser.ErrUserNotFound)
					return
				}

				token, err := authenticator.VerifyIDToken(r.Context(), idToken)
				if err != nil {
					merr.WriteError(w, duser.ErrUserNotFound)
					return
				}

				if userCache != nil {
					if cachedUser, ok := userCache.Get(r.Context(), token.UID); ok {
						next.ServeHTTP(w, r.WithContext(ctxuser.WithUser(r.Context(), cachedUser)))
						return
					}
				}

				u, err := userRepo.Get(r.Context(), token.UID)
				if err != nil {
					merr.WriteError(w, duser.ErrUserNotFound)
					return
				}

				if userCache != nil {
					userCache.Set(r.Context(), token.UID, u)
				}

				r = r.WithContext(ctxuser.WithUser(r.Context(), u))
			}

			next.ServeHTTP(w, r)
		})
	}
}
