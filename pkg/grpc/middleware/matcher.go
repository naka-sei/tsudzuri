package middleware

import (
	"strings"
)

// NewHeaderMatcher controls which incoming HTTP headers are forwarded to gRPC metadata.
// - Allow "Authorization" explicitly so auth headers pass through.
// - Allow custom headers starting with "X-".
// The returned metadata key should be lowercase to satisfy gRPC metadata requirements.
func NewHeaderMatcher() func(string) (string, bool) {
	return func(key string) (string, bool) {
		k := strings.ToLower(key)
		switch {
		case strings.HasPrefix(k, "x-"):
			return k, true
		default:
			return "", false
		}
	}
}
