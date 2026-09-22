package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Run starts the HTTP server with the provided handler and manages its lifecycle.
//
// This function handles the complete server lifecycle including:
//   - Loading configuration from environment variables via NewServerWithConfig
//   - Starting the HTTP server in a background goroutine
//   - Listening for SIGINT (Ctrl+C) or context cancellation
//   - Performing graceful shutdown with a 10-second timeout when interrupted
//
// The function blocks until the server is shut down, either by:
//   - An interrupt signal (SIGINT / Ctrl+C)
//   - Cancellation of the provided context
//   - A fatal error during server creation
//   - A failure to listen and serve (e.g., the port is already in use)
//
// Returns an error if server creation fails (e.g., invalid configuration) or if
// ListenAndServe fails for any reason other than a normal shutdown. Errors from
// Shutdown are written to stderr but do not cause the function to return an
// error, as they can occur during normal shutdown.
//
// Example usage:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
//	    w.WriteHeader(http.StatusOK)
//	})
//	if err := server.Run(context.Background(), mux); err != nil {
//	    log.Fatal(err)
//	}
//
// This function is safe for concurrent use.
func Run(
	ctx context.Context,
	srv http.Handler,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.Default()

	httpServer, err := NewServerWithConfig(srv)
	if err != nil {
		return fmt.Errorf("failed to create server with config from environment: %w", err)
	}

	// Buffered so the goroutine never blocks if we return via ctx.Done first.
	listenErr := make(chan error, 1)
	go func() {
		logger.Info(
			"server listening",
			slog.String("address", httpServer.Addr),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			listenErr <- err
		}
	}()

	select {
	case err := <-listenErr:
		return fmt.Errorf("error listening and serving: %w", err)
	case <-ctx.Done():
		// make a new context for the Shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}
	return nil
}
