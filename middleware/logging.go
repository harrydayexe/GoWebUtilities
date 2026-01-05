package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// wrappedWriter wraps http.ResponseWriter to capture the status code.
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrappedWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// NewLoggingMiddleware returns middleware that logs HTTP requests.
// Logs include method, path, status code, and duration.
func NewLoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &wrappedWriter{
				ResponseWriter: w,
				statusCode:     0,
			}

			logger.DebugContext(r.Context(), "handling request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			next.ServeHTTP(wrapped, r)

			statusCode := wrapped.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			logger.InfoContext(r.Context(), "request complete",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", statusCode),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
