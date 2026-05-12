package middleware

import (
	"net/http"
	"strings"
)

func NewStripHTMLExtension() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trimmed, ok := strings.CutSuffix(r.URL.Path, ".html"); ok {
				if trimmed == "/index" || strings.HasSuffix(trimmed, "/index") {
					trimmed = strings.TrimSuffix(trimmed, "index")
				}
				r.URL.Path = trimmed

				// Keep RawPath in sync so EscapedPath() and RequestURI() return
				// the rewritten path. Without this, percent-encoded paths like
				// /foo%2Fbar.html lose their encoding semantics after rewriting.
				if r.URL.RawPath != "" {
					if rawTrimmed, ok := strings.CutSuffix(r.URL.RawPath, ".html"); ok {
						if rawTrimmed == "/index" || strings.HasSuffix(rawTrimmed, "/index") {
							rawTrimmed = strings.TrimSuffix(rawTrimmed, "index")
						}
						r.URL.RawPath = rawTrimmed
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
