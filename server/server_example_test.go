package server_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/harrydayexe/GoWebUtilities/server"
)

// Example demonstrates basic usage of the server package with Run.
func Example() {
	// Set environment variables for configuration
	os.Setenv("PORT", "8080")
	os.Setenv("ENVIRONMENT", "local")

	// Create a simple HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Note: In a real application, this would block and run the server
	// For this example, we just demonstrate the API
	fmt.Println("Server would start on port 8080")

	// In production:
	// if err := server.Run(context.Background(), mux); err != nil {
	//     log.Fatal(err)
	// }

	// Output:
	// Server would start on port 8080
}

// ExampleNewServerWithConfig shows how to create a server with custom configuration.
func ExampleNewServerWithConfig() {
	// Suppress log output for example
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Set custom configuration via environment
	os.Setenv("PORT", "3000")
	os.Setenv("READ_TIMEOUT", "30")
	os.Setenv("WRITE_TIMEOUT", "30")
	os.Setenv("ENVIRONMENT", "production")

	// Create a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create server with configuration
	srv, err := server.NewServerWithConfig(handler)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server configured on %s\n", srv.Addr)
	fmt.Printf("Read timeout: %v\n", srv.ReadTimeout)
	fmt.Printf("Write timeout: %v\n", srv.WriteTimeout)
}

// ExampleNewServerWithConfig_defaults shows server creation with default configuration.
func ExampleNewServerWithConfig_defaults() {
	// Suppress log output for example
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Clear environment to use defaults
	os.Unsetenv("PORT")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("READ_TIMEOUT")
	os.Unsetenv("WRITE_TIMEOUT")
	os.Unsetenv("IDLE_TIMEOUT")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv, err := server.NewServerWithConfig(handler)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server address: %s\n", srv.Addr)
	fmt.Printf("Read timeout: %v\n", srv.ReadTimeout)
	fmt.Printf("Idle timeout: %v\n", srv.IdleTimeout)
}

// ExampleNewServerWithConfig_errorHandling demonstrates error handling for invalid configuration.
func ExampleNewServerWithConfig_errorHandling() {
	// Set invalid environment
	os.Setenv("ENVIRONMENT", "staging")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	srv, err := server.NewServerWithConfig(handler)
	if err != nil {
		fmt.Println("Configuration error occurred")
		fmt.Println("Using fallback configuration")
		// Could create default server or handle error differently
		return
	}

	fmt.Printf("Server created: %v\n", srv)

	// Output:
	// Configuration error occurred
	// Using fallback configuration
}

// ExampleRun shows the complete pattern for running a server with graceful shutdown.
func ExampleRun() {
	// For example purposes, we'll demonstrate the pattern
	// without actually starting a long-running server

	// Set configuration
	os.Setenv("PORT", "8080")
	os.Setenv("ENVIRONMENT", "local")

	// Create handler
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	fmt.Println("Server configuration loaded")
	fmt.Println("Handler registered at /health")
	fmt.Println("Server would run until interrupted (Ctrl+C)")

	// In production:
	// if err := server.Run(context.Background(), mux); err != nil {
	//     log.Fatal(err)
	// }

	// Output:
	// Server configuration loaded
	// Handler registered at /health
	// Server would run until interrupted (Ctrl+C)
}

// ExampleRun_withContext shows context-based cancellation for graceful shutdown.
func ExampleRun_withContext() {
	// Suppress log output for example
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	os.Setenv("PORT", "0") // Use any available port
	os.Setenv("ENVIRONMENT", "test")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Run server in background
	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx, handler)
	}()

	// Simulate doing work
	time.Sleep(100 * time.Millisecond)

	// Cancel context to trigger graceful shutdown
	cancel()

	// Wait for server to shut down
	err := <-done
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Server shut down gracefully")
	}
}

// ExampleRun_healthCheck demonstrates a realistic health check endpoint pattern.
func ExampleRun_healthCheck() {
	// For documentation, show the pattern
	os.Setenv("PORT", "8080")
	os.Setenv("ENVIRONMENT", "production")

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"2024-01-01T00:00:00Z"}`))
	})

	// API endpoints
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"API endpoint"}`))
	})

	fmt.Println("Health check available at /health")
	fmt.Println("API available at /api/")
	fmt.Println("Server ready to accept requests")

	// In production:
	// if err := server.Run(context.Background(), mux); err != nil {
	//     log.Fatal(err)
	// }

	// Output:
	// Health check available at /health
	// API available at /api/
	// Server ready to accept requests
}
