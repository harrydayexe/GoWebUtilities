package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper Functions

// clearServerEnvVars clears all server configuration environment variables
func clearServerEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{"PORT", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT", "ENVIRONMENT", "VERBOSE"}
	for _, v := range envVars {
		t.Setenv(v, "")
	}
}

// clearOtherServerEnvVars clears all server env vars except PORT
func clearOtherServerEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{"READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT", "ENVIRONMENT", "VERBOSE"}
	for _, v := range envVars {
		t.Setenv(v, "")
	}
}

// findAvailablePort finds an available port for testing
func findAvailablePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

// assertContains checks if a string contains a substring
func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected %q to contain %q", got, want)
	}
}

// assertNotNil checks if a value is not nil
func assertNotNil(t *testing.T, val interface{}, name string) {
	t.Helper()
	if val == nil {
		t.Errorf("expected %s to not be nil", name)
	}
}

// NewServerWithConfig Tests

func TestNewServerWithConfig_DefaultConfiguration(t *testing.T) {
	// Clear all environment variables to use defaults
	clearServerEnvVars(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv, err := NewServerWithConfig(handler)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	assertNotNil(t, srv, "server")

	// Verify default values
	if srv.Addr != ":8080" {
		t.Errorf("expected default port :8080, got: %s", srv.Addr)
	}
	if srv.ReadTimeout != 15*time.Second {
		t.Errorf("expected ReadTimeout 15s, got: %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 15*time.Second {
		t.Errorf("expected WriteTimeout 15s, got: %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 60*time.Second {
		t.Errorf("expected IdleTimeout 60s, got: %v", srv.IdleTimeout)
	}
	if srv.Handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestNewServerWithConfig_CustomConfiguration(t *testing.T) {
	// Set custom environment variables
	t.Setenv("PORT", "3000")
	t.Setenv("READ_TIMEOUT", "30")
	t.Setenv("WRITE_TIMEOUT", "45")
	t.Setenv("IDLE_TIMEOUT", "120")
	t.Setenv("ENVIRONMENT", "production")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	srv, err := NewServerWithConfig(handler)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	assertNotNil(t, srv, "server")

	// Verify custom values
	if srv.Addr != ":3000" {
		t.Errorf("expected port :3000, got: %s", srv.Addr)
	}
	if srv.ReadTimeout != 30*time.Second {
		t.Errorf("expected ReadTimeout 30s, got: %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 45*time.Second {
		t.Errorf("expected WriteTimeout 45s, got: %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("expected IdleTimeout 120s, got: %v", srv.IdleTimeout)
	}
}

func TestNewServerWithConfig_InvalidEnvironment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "staging") // Invalid value

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	srv, err := NewServerWithConfig(handler)

	// Should return error
	if err == nil {
		t.Fatal("expected error for invalid environment, got nil")
	}

	assertContains(t, err.Error(), "config validation failed")
	assertContains(t, err.Error(), "invalid environment: staging")

	// Server should be nil
	if srv != nil {
		t.Errorf("expected nil server on error, got: %v", srv)
	}
}

func TestNewServerWithConfig_InvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-number")
	t.Setenv("ENVIRONMENT", "local")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	srv, err := NewServerWithConfig(handler)

	// Should return error
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}

	assertContains(t, err.Error(), "failed to parse config")

	// Server should be nil
	if srv != nil {
		t.Errorf("expected nil server on error, got: %v", srv)
	}
}

func TestNewServerWithConfig_HandlerIntegration(t *testing.T) {
	clearServerEnvVars(t)

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	srv, err := NewServerWithConfig(handler)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Create test request/response to verify handler
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.Handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got: %d", w.Code)
	}
	if w.Body.String() != "test response" {
		t.Errorf("expected 'test response', got: %s", w.Body.String())
	}
}

func TestNewServerWithConfig_ConcurrentCreation(t *testing.T) {
	clearServerEnvVars(t)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			srv, err := NewServerWithConfig(handler)
			if err != nil {
				errors <- err
				return
			}
			if srv == nil {
				errors <- fmt.Errorf("got nil server")
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent creation error: %v", err)
	}
}

// Run Function Tests

func TestRun_ContextCancellation(t *testing.T) {
	// Suppress log output for this test
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Use high port to avoid conflicts
	port := findAvailablePort(t)
	t.Setenv("PORT", fmt.Sprintf("%d", port))
	clearOtherServerEnvVars(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())

	runComplete := make(chan error, 1)
	go func() {
		runComplete <- Run(ctx, handler)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Verify Run returns within reasonable time
	select {
	case err := <-runComplete:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not complete within timeout after context cancellation")
	}
}

func TestRun_ConfigurationError(t *testing.T) {
	t.Setenv("ENVIRONMENT", "invalid")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	err := Run(context.Background(), handler)

	// Should return error
	if err == nil {
		t.Fatal("expected error for invalid configuration, got nil")
	}

	assertContains(t, err.Error(), "failed to create server with config")
}

func TestRun_GracefulShutdown(t *testing.T) {
	// Suppress log output for this test
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	port := findAvailablePort(t)
	t.Setenv("PORT", fmt.Sprintf("%d", port))
	clearOtherServerEnvVars(t)

	requestStarted := make(chan struct{})
	requestComplete := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		// Simulate slow request
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		close(requestComplete)
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		Run(ctx, handler)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start a request in background
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", port))
		if err != nil {
			t.Logf("request error: %v", err)
			return
		}
		resp.Body.Close()
	}()

	// Wait for request to start processing
	<-requestStarted

	// Trigger shutdown
	cancel()

	// Verify request completes
	select {
	case <-requestComplete:
		// Success - request completed during graceful shutdown
	case <-time.After(5 * time.Second):
		t.Error("request did not complete during graceful shutdown")
	}
}

func TestRun_MultipleShutdowns(t *testing.T) {
	// Suppress log output for this test
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	port := findAvailablePort(t)
	t.Setenv("PORT", fmt.Sprintf("%d", port))
	clearOtherServerEnvVars(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())

	runComplete := make(chan error, 1)
	go func() {
		runComplete <- Run(ctx, handler)
	}()

	time.Sleep(100 * time.Millisecond)

	// Cancel multiple times
	cancel()
	cancel()
	cancel()

	// Should not panic or hang
	select {
	case err := <-runComplete:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not complete within timeout")
	}
}
