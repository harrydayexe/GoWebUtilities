package middleware_test

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/harrydayexe/GoWebUtilities/middleware"
)

// Example demonstrates a realistic JSON API middleware stack.
func Example() {
	// Create a logger for production use
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create a typical middleware stack for JSON APIs
	// Order: logging (first) -> size limiting -> content-type (last)
	apiStack := middleware.CreateStack(
		middleware.NewLoggingMiddleware(logger),
		middleware.NewMaxBytesReader(10*1024*1024), // 10MB limit
		middleware.NewSetContentTypeJSON(),
	)

	// Create a mux and add API endpoints
	mux := http.NewServeMux()

	// User endpoint
	getUserHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"name":"Alice"}`))
	})

	// Apply middleware to all API routes
	mux.Handle("/api/users/", apiStack(getUserHandler))

	fmt.Println("JSON API server configured with logging, size limiting, and content-type middleware")
	// Output:
	// JSON API server configured with logging, size limiting, and content-type middleware
}

// ExampleNewSetContentTypeJSON demonstrates creating and using the JSON content-type middleware.
func ExampleNewSetContentTypeJSON() {
	// Create a JSON content-type middleware
	jsonMiddleware := middleware.NewSetContentTypeJSON()

	// Create a simple handler that returns JSON data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Apply the middleware to the handler
	wrappedHandler := jsonMiddleware(handler)

	// Create a test server
	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	// Output:
	// Content-Type: application/json
}

// ExampleNewMaxBytesReader demonstrates limiting request body size.
func ExampleNewMaxBytesReader() {
	// Create middleware that limits request body to 512 bytes
	limitMiddleware := middleware.NewMaxBytesReader(512)

	// Create a handler that reads the request body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		fmt.Fprintf(w, "Received %d bytes", len(body))
	})

	// Apply the middleware
	wrappedHandler := limitMiddleware(handler)

	// Create a test server
	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	// Test with small payload (should succeed)
	smallPayload := strings.NewReader("small data")
	resp, err := http.Post(server.URL, "text/plain", smallPayload)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", body)
	// Output:
	// Status: 200
	// Response: Received 10 bytes
}

// ExampleNewLoggingMiddleware demonstrates structured logging with middleware.
func ExampleNewLoggingMiddleware() {
	// Create a structured logger (logs to stderr in production)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create logging middleware
	loggingMw := middleware.NewLoggingMiddleware(logger)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Apply the middleware
	wrappedHandler := loggingMw(handler)

	// Create a test server
	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	// Make a request (logging happens automatically)
	resp, err := http.Get(server.URL + "/api/test")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Request logged with status: %d\n", resp.StatusCode)
	// Output:
	// Request logged with status: 200
}

// ExampleCreateStack demonstrates composing multiple middleware.
func ExampleCreateStack() {
	// Create a middleware stack with multiple middleware
	stack := middleware.CreateStack(
		middleware.NewSetContentTypeJSON(),
		middleware.NewMaxBytesReader(1024),
	)

	// Create a base handler
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"hello"}`))
	})

	// Apply the stack to the handler
	wrappedHandler := stack(baseHandler)

	// Create a test server
	server := httptest.NewServer(wrappedHandler)
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	// Output:
	// Content-Type: application/json
}

// ExampleCreateStack_complete demonstrates a complete HTTP server setup with middleware.
func ExampleCreateStack_complete() {
	// Create a logger (discarding output for this example)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	// Create a complete middleware stack
	stack := middleware.CreateStack(
		middleware.NewLoggingMiddleware(logger),
		middleware.NewSetContentTypeJSON(),
		middleware.NewMaxBytesReader(1024),
	)

	// Create a mux and add handlers
	mux := http.NewServeMux()

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"hello"}`))
	})

	// Apply middleware to the handler
	mux.Handle("/api/", stack(apiHandler))

	// Create a test server
	server := httptest.NewServer(mux)
	defer server.Close()

	// Make a request
	resp, err := http.Get(server.URL + "/api/test")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	// Output:
	// Status: 200
	// Content-Type: application/json
}
