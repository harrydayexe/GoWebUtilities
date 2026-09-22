package middleware

import (
	"net/http"
	"strings"
)

// NewRedirectWWW returns middleware that redirects requests to the "www."
// subdomain to the root domain.
//
// Example:
//   - "https://www.example.com/foo?bar=baz" -> 301 "https://example.com/foo?bar=baz"
//   - "http://www.example.com" -> 301 "https://example.com"
func NewRedirectWWW() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if host, ok := strings.CutPrefix(r.Host, "www."); ok {
				http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusMovedPermanently)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
