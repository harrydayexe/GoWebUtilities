package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// NewCacheControl returns middleware that sets the Cache-Control header
// on every response to "public, max-age=<seconds>", where seconds is derived
// from ttl. The header value is computed once at middleware creation time.
func NewCacheControl(ttl time.Duration) Middleware {
	value := fmt.Sprintf("public, max-age=%d", int(ttl.Seconds()))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", value)
			next.ServeHTTP(w, r)
		})
	}
}
