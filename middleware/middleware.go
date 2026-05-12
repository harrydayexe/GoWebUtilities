package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler,
// providing functionality before and after execution
// of the wrapped handler.
type Middleware func(h http.Handler) http.Handler

// CreateStack composes multiple middleware into a single middleware.
// Middleware are applied in the order provided: the first middleware
// in the list will be the outermost wrapper (executed first on the request).
func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}

		return next
	}
}
