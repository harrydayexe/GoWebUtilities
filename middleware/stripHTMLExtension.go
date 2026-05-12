package middleware

import (
	"net/http"
	"strings"
)

// NewStripHTMLExtension returns middleware that rewrites incoming request paths
// by removing any trailing ".html" suffix before passing the request to the next
// handler. This allows file-based routers to serve clean URLs — e.g. a request
// for "/about.html" is handled as if the client requested "/about".
//
// Index pages are handled as a special case: paths that end in "/index" after
// stripping (including the bare "/index.html" root path) are further trimmed to
// the parent directory path with a trailing slash. For example, "/about/index.html"
// becomes "/about/" and "/index.html" becomes "/".
//
// When the URL contains percent-encoded characters that cause Go's HTTP parser to
// populate url.RawPath (most notably %2F, an encoded slash), RawPath is updated in
// sync with Path so that EscapedPath() and RequestURI() continue to return the
// rewritten path with correct encoding semantics.
//
// Paths that do not end in ".html" are forwarded to the next handler unchanged.
// Query parameters and URL fragments are never modified.
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
