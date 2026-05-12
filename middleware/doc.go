// Package middleware provides composable HTTP middleware for web applications.
//
// All middleware follow Go's standard net/http handler wrapping pattern:
//
//	type Middleware func(h http.Handler) http.Handler
//
// Multiple middleware can be composed into a single middleware using CreateStack,
// which applies them in the order provided so that the first argument is the
// outermost wrapper and therefore the first to execute on each request.
//
// Available middleware:
//
//   - NewLoggingMiddleware: structured request logging via log/slog, recording
//     method, path, status code, and duration.
//   - NewMaxBytesReader: limits request body size to prevent resource exhaustion.
//   - NewSetContentType / NewSetContentTypeJSON: sets the Content-Type response header.
//   - NewStripHTMLExtension: rewrites ".html" paths to clean URLs before routing.
//
// Example — composing a middleware stack for a JSON API:
//
//	stack := middleware.CreateStack(
//	    middleware.NewLoggingMiddleware(logger),
//	    middleware.NewMaxBytesReader(1024*1024), // 1 MB
//	    middleware.NewSetContentTypeJSON(),
//	)
//	http.ListenAndServe(":8080", stack(mux))
package middleware
