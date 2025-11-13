package middleware

import "strings"

// NewHeaderMatcher returns a header matcher that matches headers starting with "X-".
func NewHeaderMatcher() func(string) (string, bool) {
	return func(key string) (string, bool) {
		if strings.HasPrefix(strings.ToLower(key), "x-") {
			return key, true
		}
		return "", false
	}
}
