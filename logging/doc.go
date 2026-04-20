// Package logging provides utilities for configuring structured logging in web applications.
//
// The package simplifies logger setup by automatically configuring the default slog logger
// based on the application's environment. It selects appropriate log handlers (Text for
// local development, JSON for test/production) and log levels based on the LOG_LEVEL
// environment variable.
//
// Basic usage:
//
//	cfg := config.ServerConfig{
//		Environment: config.Production,
//		LogLevel:    slog.LevelWarn,
//	}
//	logging.SetDefaultLogger(cfg)
//
//	// Now use slog throughout your application
//	slog.Info("application started", "port", 8080)
//
// Environment-specific behavior:
//   - Local: Text handler for human-readable logs during development
//   - Test/Production: JSON handler for structured log aggregation
//
// Log level configuration (via LOG_LEVEL environment variable):
//   - "DEBUG": DEBUG level and above (all messages)
//   - "INFO": INFO level and above
//   - "WARN": WARN level and above (default)
//   - "ERROR": ERROR level only
//
// Concurrency:
//
// SetDefaultLogger is NOT safe for concurrent use. It should be called once during
// application initialization before spawning goroutines that use slog. After initialization,
// the configured logger is safe for concurrent use across goroutines.
//
// Integration with log/slog:
//
// This package configures the default logger used by slog.Info, slog.Debug, and other
// top-level slog functions via slog.SetDefault(). All standard slog functionality is
// available after configuration.
package logging
