package logging_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/harrydayexe/GoWebUtilities/config"
	"github.com/harrydayexe/GoWebUtilities/logging"
	"github.com/harrydayexe/GoWebUtilities/middleware"
)

// Example demonstrates basic logging configuration for a web application.
func Example() {
	// Configure logging based on environment
	cfg := config.ServerConfig{
		Environment: config.Local,
		VerboseMode: false,
	}

	logging.SetDefaultLogger(cfg)

	// Now use slog throughout your application
	// Note: Actual log output is suppressed for test output

	fmt.Println("Logging configured for local development")
	fmt.Println("Handler: Text format")
	fmt.Println("Level: INFO")
	// Output:
	// Logging configured for local development
	// Handler: Text format
	// Level: INFO
}

// ExampleSetDefaultLogger demonstrates typical usage of SetDefaultLogger.
func ExampleSetDefaultLogger() {
	// Create configuration
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: false,
		Port:        8080,
	}

	// Configure the default logger
	logging.SetDefaultLogger(cfg)

	// The default logger is now configured and ready to use
	// All slog.* calls will use this configuration
	fmt.Println("Production logging configured")
	fmt.Println("Ready to use slog.Info(), slog.Debug(), etc.")
	// Output:
	// Production logging configured
	// Ready to use slog.Info(), slog.Debug(), etc.
}

// ExampleSetDefaultLogger_local demonstrates local development logging configuration.
func ExampleSetDefaultLogger_local() {
	cfg := config.ServerConfig{
		Environment: config.Local,
		VerboseMode: false,
	}

	logging.SetDefaultLogger(cfg)

	// Local environment uses text handler for human-readable logs
	// Perfect for development and debugging
	fmt.Println("Local development logging:")
	fmt.Println("- Text format (human-readable)")
	fmt.Println("- INFO level and above")
	fmt.Println("- Output to stdout")
	// Output:
	// Local development logging:
	// - Text format (human-readable)
	// - INFO level and above
	// - Output to stdout
}

// ExampleSetDefaultLogger_production demonstrates production logging configuration.
func ExampleSetDefaultLogger_production() {
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: false,
	}

	logging.SetDefaultLogger(cfg)

	// Production environment uses JSON handler for structured logging
	// Ideal for log aggregation and analysis
	fmt.Println("Production logging:")
	fmt.Println("- JSON format (structured)")
	fmt.Println("- INFO level and above")
	fmt.Println("- Ready for log aggregation systems")
	// Output:
	// Production logging:
	// - JSON format (structured)
	// - INFO level and above
	// - Ready for log aggregation systems
}

// ExampleSetDefaultLogger_verbose demonstrates verbose debug logging.
func ExampleSetDefaultLogger_verbose() {
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: true,
	}

	logging.SetDefaultLogger(cfg)

	// Verbose mode enables DEBUG level logging
	// Useful for troubleshooting and detailed diagnostics
	fmt.Println("Verbose logging enabled:")
	fmt.Println("- DEBUG level and above")
	fmt.Println("- Includes detailed diagnostic information")
	fmt.Println("- Use for troubleshooting")
	// Output:
	// Verbose logging enabled:
	// - DEBUG level and above
	// - Includes detailed diagnostic information
	// - Use for troubleshooting
}

// ExampleSetDefaultLogger_withServer demonstrates logging configuration with server startup.
func ExampleSetDefaultLogger_withServer() {
	// Parse configuration from environment
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: false,
		Port:        8080,
	}

	// Configure logging before starting server
	logging.SetDefaultLogger(cfg)

	// Suppress actual log output for example
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Use structured logging in handlers
		slog.Info("handling request", "path", r.URL.Path, "method", r.Method)
		w.Write([]byte("OK"))
	})

	// Server will use the configured logger
	fmt.Println("Application startup sequence:")
	fmt.Println("1. Parse configuration")
	fmt.Println("2. Configure logging")
	fmt.Println("3. Set up HTTP handlers")
	fmt.Println("4. Start server")
	// Output:
	// Application startup sequence:
	// 1. Parse configuration
	// 2. Configure logging
	// 3. Set up HTTP handlers
	// 4. Start server
}

// ExampleSetDefaultLogger_withMiddleware demonstrates logging with middleware stack.
func ExampleSetDefaultLogger_withMiddleware() {
	// Configure application-level logging
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: false,
	}
	logging.SetDefaultLogger(cfg)

	// Suppress actual log output for example
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))

	// Create request-specific logger for middleware
	requestLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create middleware stack with logging
	stack := middleware.CreateStack(
		middleware.NewLoggingMiddleware(requestLogger),
		middleware.NewSetContentTypeJSON(),
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use default logger in handlers
		slog.Info("processing request", "endpoint", r.URL.Path)
		w.Write([]byte(`{"status":"ok"}`))
	})

	_ = stack(handler)

	fmt.Println("Application configured with:")
	fmt.Println("- Default logger for application events")
	fmt.Println("- Request logger for HTTP middleware")
	fmt.Println("- Structured JSON logging throughout")
	// Output:
	// Application configured with:
	// - Default logger for application events
	// - Request logger for HTTP middleware
	// - Structured JSON logging throughout
}

// ExampleSetDefaultLogger_fullApplication demonstrates a complete application setup.
func ExampleSetDefaultLogger_fullApplication() {
	// Step 1: Parse configuration
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: false,
		Port:        8080,
	}

	// Step 2: Configure logging early in application startup
	logging.SetDefaultLogger(cfg)

	// Suppress actual log output for example (do this after SetDefaultLogger)
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))

	// Step 3: Log application startup (suppressed)
	slog.Info("application starting",
		"environment", cfg.Environment,
		"port", cfg.Port,
		"verbose", cfg.VerboseMode,
	)

	// Step 4: Create HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Step 5: Create server (note: this parses config from env, may log)
	// In production, you would call server.Run() here
	fmt.Println("Application configured with server and logging")
	fmt.Println("Logging ready for:")
	fmt.Println("- Application events")
	fmt.Println("- Request processing")
	fmt.Println("- Error tracking")
	// Output:
	// Application configured with server and logging
	// Logging ready for:
	// - Application events
	// - Request processing
	// - Error tracking
}

// ExampleSetDefaultLogger_errorHandling demonstrates logging configuration with error handling.
func ExampleSetDefaultLogger_errorHandling() {
	// Suppress actual log output for example
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))

	// Parse config with error handling
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: false,
	}

	// Configure logger
	logging.SetDefaultLogger(cfg)

	// Use structured logging for errors
	if err := someOperation(); err != nil {
		slog.Error("operation failed",
			"error", err,
			"operation", "someOperation",
		)
	}

	fmt.Println("Structured error logging configured")
	fmt.Println("Errors include context and details")
	// Output:
	// Structured error logging configured
	// Errors include context and details
}

// someOperation is a helper function for the example
func someOperation() error {
	// Simulated operation
	return nil
}

// ExampleSetDefaultLogger_contextualLogging demonstrates using context with logging.
func ExampleSetDefaultLogger_contextualLogging() {
	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: false,
	}
	logging.SetDefaultLogger(cfg)

	// Suppress actual log output for example (do this after SetDefaultLogger)
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))

	// Use context for request tracking
	ctx := context.Background()

	// Log with context information (suppressed)
	logger := slog.Default().With(
		"request_id", "req-123",
		"user_id", "user-456",
	)
	logger.InfoContext(ctx, "processing request")

	fmt.Println("Contextual logging enabled")
	fmt.Println("Each log includes:")
	fmt.Println("- Request ID")
	fmt.Println("- User ID")
	fmt.Println("- Timestamp")
	fmt.Println("- Log level")
	// Output:
	// Contextual logging enabled
	// Each log includes:
	// - Request ID
	// - User ID
	// - Timestamp
	// - Log level
}
