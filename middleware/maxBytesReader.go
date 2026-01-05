package middleware

import "net/http"

// NewMaxBytesReader returns middleware that limits request body size to maxBytes.
// Bodies exceeding this limit will cause an error response. This prevents
// resource exhaustion from overly large requests.
// If maxBytes is 0, defaults to 1MB
func NewMaxBytesReader(maxBytes int64) Middleware {
	if maxBytes == 0 {
		maxBytes = 1048576 // 1MB default
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
