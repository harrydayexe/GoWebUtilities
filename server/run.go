package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/harrydayexe/GoWebUtilities/config"
)

// Run starts the HTTP server with the provided handler.
func Run(
	ctx context.Context,
	srv http.Handler,
	cfg config.ServerConfig,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.Default()

	httpServer, err := NewServerWithConfig(srv)
	if err != nil {
		return fmt.Errorf("failed to create server with config from environment: %w", err)
	}

	go func() {
		logger.Info(
			"server listening",
			slog.String("address", httpServer.Addr),
			slog.String("environment", cfg.Environment.String()),
		)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// wait for ctx cancellation
		<-ctx.Done()
		// make a new context for the Shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}
