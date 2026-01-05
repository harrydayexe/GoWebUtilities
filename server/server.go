package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/harrydayexe/GoWebUtilities/config"
	"github.com/harrydayexe/GoWebUtilities/logging"
)

// NewServerWithConfig creates a new http.Server configured from environment variables.
//
// The server is configured using config.ServerConfig, which loads settings for:
//   - Port (env: PORT, default: 8080)
//   - ReadTimeout (env: READ_TIMEOUT, default: 5 seconds)
//   - WriteTimeout (env: WRITE_TIMEOUT, default: 10 seconds)
//   - IdleTimeout (env: IDLE_TIMEOUT, default: 120 seconds)
//   - Environment (env: ENVIRONMENT, default: Local)
//
// The function returns an error if the configuration cannot be parsed or validated.
// Common error cases include invalid port numbers or timeout values.
//
// The returned server is ready to use with ListenAndServe or Shutdown methods.
// For automatic lifecycle management with graceful shutdown, use the Run function instead.
//
// This function is safe for concurrent use.
func NewServerWithConfig(handler http.Handler) (*http.Server, error) {
	cfg, err := config.ParseConfig[config.ServerConfig]()
	if err != nil {
		return nil, fmt.Errorf("failed to create config from environment: %w", err)
	}

	logging.SetDefaultLogger(cfg)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	slog.Default().Info("created server", slog.String("environment", cfg.Environment.String()))

	return httpServer, nil
}
