package middleware

import "net/http"

// NewSetContentType returns middleware that sets the Content-Type header
// for all responses. Common values: "application/json", "text/html; charset=utf-8".
func NewSetContentType(contentType string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

// NewSetContentTypeJSON returns middleware that sets the Content-Type header
// to application/json. This is a convenience wrapper around NewSetContentType
// for JSON APIs. Apply this to route groups where all endpoints return JSON
// to avoid repetitive header setting in individual handlers.
func NewSetContentTypeJSON() Middleware {
	return NewSetContentType("application/json")
}
