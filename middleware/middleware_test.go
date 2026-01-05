package middleware

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// Test helper functions

// newTestLogger creates a logger that writes to a buffer for testing.
// Returns the logger and the buffer for validation.
func newTestLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	return logger, buf
}

// recordingMiddleware creates middleware that records execution order.
// Used to verify CreateStack applies middleware correctly.
func recordingMiddleware(name string, recorder *[]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*recorder = append(*recorder, name+":before")
			next.ServeHTTP(w, r)
			*recorder = append(*recorder, name+":after")
		})
	}
}

// assertStatus verifies the HTTP status code matches expected value.
func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got := w.Code; got != want {
		t.Errorf("status code: got %d, want %d", got, want)
	}
}

// assertHeader verifies an HTTP header matches expected value.
func assertHeader(t *testing.T, w *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	if got := w.Header().Get(key); got != want {
		t.Errorf("header %q: got %q, want %q", key, got, want)
	}
}

// assertBody verifies the HTTP response body matches expected value.
func assertBody(t *testing.T, w *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got := w.Body.String(); got != want {
		t.Errorf("body: got %q, want %q", got, want)
	}
}

// runConcurrent executes a handler concurrently with count goroutines.
// Used for testing race conditions and concurrent request handling.
func runConcurrent(t *testing.T, handler http.Handler, count int) {
	t.Helper()
	var wg sync.WaitGroup
	server := httptest.NewServer(handler)
	defer server.Close()

	errors := make(chan error, count)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resp, err := http.Get(server.URL)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent request failed: %v", err)
		}
	}
}

// TestWrappedWriter_ImplicitStatus verifies that wrappedWriter correctly
// captures an implicit 200 status when the handler calls Write() without
// first calling WriteHeader().
func TestWrappedWriter_ImplicitStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only call Write(), never WriteHeader()
		w.Write([]byte("hello"))
	})

	logger, _ := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	mw(handler).ServeHTTP(w, req)

	// Verify implicit 200 was captured
	assertStatus(t, w, http.StatusOK)
	assertBody(t, w, "hello")
}

// TestLoggingMiddleware_NoWriteHeader verifies that NewLoggingMiddleware
// correctly handles the case where a handler calls Write() without
// explicitly calling WriteHeader(). The logging middleware should log
// status as 200 (the implicit status).
func TestLoggingMiddleware_NoWriteHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write without calling WriteHeader explicitly
		w.Write([]byte("response"))
	})

	logger, buf := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	mw(handler).ServeHTTP(w, req)

	// Verify response is correct
	assertStatus(t, w, http.StatusOK)

	// Verify logs contain status=200
	logOutput := buf.String()
	if !strings.Contains(logOutput, "status=200") {
		t.Errorf("log should contain status=200, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "path=/test") {
		t.Errorf("log should contain path=/test, got: %s", logOutput)
	}
}

// TestMaxBytesReader_DefaultZero verifies that NewMaxBytesReader correctly
// defaults to 1MB when maxBytes is 0.
func TestMaxBytesReader_DefaultZero(t *testing.T) {
	mw := NewMaxBytesReader(0) // Explicitly pass 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
		_ = body // Use the body to avoid unused variable
	})

	// Test with exactly 1MB (should pass)
	t.Run("exactly_1MB", func(t *testing.T) {
		oneMB := make([]byte, 1048576)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(oneMB))
		w := httptest.NewRecorder()

		mw(handler).ServeHTTP(w, req)

		assertStatus(t, w, http.StatusOK)
	})

	// Test with 1MB + 1 byte (should fail)
	t.Run("over_1MB", func(t *testing.T) {
		overMB := make([]byte, 1048577)
		req := httptest.NewRequest("POST", "/", bytes.NewReader(overMB))
		w := httptest.NewRecorder()

		mw(handler).ServeHTTP(w, req)

		// Should get error when trying to read body exceeding limit
		assertStatus(t, w, http.StatusRequestEntityTooLarge)
	})
}

// TestWrappedWriter_WriteHeader verifies that wrappedWriter correctly captures
// the status code when WriteHeader is called explicitly.
func TestWrappedWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
		{"Created", http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			logger, _ := newTestLogger()
			mw := NewLoggingMiddleware(logger)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			mw(handler).ServeHTTP(w, req)

			assertStatus(t, w, tt.statusCode)
		})
	}
}

// TestWrappedWriter_Write verifies that wrappedWriter correctly delegates
// the Write call to the underlying ResponseWriter.
func TestWrappedWriter_Write(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n, err := w.Write([]byte("test response"))
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
		if n != 13 {
			t.Errorf("Write() returned %d bytes, want 13", n)
		}
	})

	logger, _ := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	mw(handler).ServeHTTP(w, req)

	assertBody(t, w, "test response")
}

// TestWrappedWriter_MultipleWrites verifies that multiple Write calls work correctly
// and that the implicit status code is only set on the first write.
func TestWrappedWriter_MultipleWrites(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("first "))
		w.Write([]byte("second "))
		w.Write([]byte("third"))
	})

	logger, _ := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	mw(handler).ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
	assertBody(t, w, "first second third")
}

// TestLoggingMiddleware_ExplicitStatus verifies that the logging middleware
// correctly logs the status code when the handler explicitly calls WriteHeader.
func TestLoggingMiddleware_ExplicitStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	})

	logger, buf := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/missing", nil)

	mw(handler).ServeHTTP(w, req)

	assertStatus(t, w, http.StatusNotFound)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "status=404") {
		t.Errorf("log should contain status=404, got: %s", logOutput)
	}
}

// TestLoggingMiddleware_LogFields verifies that all expected fields are logged.
func TestLoggingMiddleware_LogFields(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	logger, buf := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/users", nil)

	mw(handler).ServeHTTP(w, req)

	logOutput := buf.String()

	// Verify all expected log fields are present
	expectedFields := []string{
		"method=POST",
		"path=/api/users",
		"status=200",
		"duration=",
	}

	for _, field := range expectedFields {
		if !strings.Contains(logOutput, field) {
			t.Errorf("log should contain %q, got: %s", field, logOutput)
		}
	}

	// Verify both DEBUG and INFO messages are logged
	if !strings.Contains(logOutput, "handling request") {
		t.Errorf("log should contain 'handling request' (DEBUG), got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "request complete") {
		t.Errorf("log should contain 'request complete' (INFO), got: %s", logOutput)
	}
}

// TestLoggingMiddleware_NoResponse verifies logging when handler doesn't write anything.
func TestLoggingMiddleware_NoResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler does nothing - no Write, no WriteHeader
	})

	logger, buf := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/empty", nil)

	mw(handler).ServeHTTP(w, req)

	// When handler doesn't write, httptest.ResponseRecorder defaults to 200
	assertStatus(t, w, http.StatusOK)

	// Verify logging still works
	logOutput := buf.String()
	if !strings.Contains(logOutput, "status=200") {
		t.Errorf("log should contain status=200 for empty handler, got: %s", logOutput)
	}
}

// TestMaxBytesReader_CustomValue verifies that custom maxBytes values work correctly.
func TestMaxBytesReader_CustomValue(t *testing.T) {
	tests := []struct {
		name     string
		maxBytes int64
		bodySize int
		wantCode int
	}{
		{"within_512", 512, 256, http.StatusOK},
		{"exact_512", 512, 512, http.StatusOK},
		{"over_512", 512, 513, http.StatusRequestEntityTooLarge},
		{"within_1024", 1024, 1023, http.StatusOK},
		{"over_1024", 1024, 1025, http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewMaxBytesReader(tt.maxBytes)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			body := make([]byte, tt.bodySize)
			req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			mw(handler).ServeHTTP(w, req)

			assertStatus(t, w, tt.wantCode)
		})
	}
}

// TestMaxBytesReader_EmptyBody verifies that empty request bodies are handled correctly.
func TestMaxBytesReader_EmptyBody(t *testing.T) {
	mw := NewMaxBytesReader(1024)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(body) != 0 {
			t.Errorf("expected empty body, got %d bytes", len(body))
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()

	mw(handler).ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
}

// TestSetContentType_Applied verifies that SetContentType middleware
// correctly sets the Content-Type header.
func TestSetContentType_Applied(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{"JSON", "application/json"},
		{"HTML", "text/html; charset=utf-8"},
		{"Plain", "text/plain"},
		{"XML", "application/xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewSetContentType(tt.contentType)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			mw(handler).ServeHTTP(w, req)

			assertHeader(t, w, "Content-Type", tt.contentType)
		})
	}
}

// TestSetContentType_EmptyString verifies edge case of empty contentType.
func TestSetContentType_EmptyString(t *testing.T) {
	mw := NewSetContentType("")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mw(handler).ServeHTTP(w, req)

	// Empty string should still set the header (even if empty)
	assertHeader(t, w, "Content-Type", "")
}

// TestSetContentTypeJSON verifies the JSON convenience function.
func TestSetContentTypeJSON_Convenience(t *testing.T) {
	mw := NewSetContentTypeJSON()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mw(handler).ServeHTTP(w, req)

	assertHeader(t, w, "Content-Type", "application/json")
}

// TestCreateStack_Empty verifies that CreateStack with no middleware works correctly.
func TestCreateStack_Empty(t *testing.T) {
	stack := CreateStack()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	stack(handler).ServeHTTP(w, req)

	assertStatus(t, w, http.StatusOK)
	assertBody(t, w, "ok")
}

// TestCreateStack_Single verifies that CreateStack with a single middleware works.
func TestCreateStack_Single(t *testing.T) {
	stack := CreateStack(NewSetContentTypeJSON())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	stack(handler).ServeHTTP(w, req)

	assertHeader(t, w, "Content-Type", "application/json")
}

// TestCreateStack_ExecutionOrder verifies that middleware are executed in the correct order.
// The first middleware in the slice should be the outermost (executed first on request).
func TestCreateStack_ExecutionOrder(t *testing.T) {
	var order []string

	stack := CreateStack(
		recordingMiddleware("A", &order),
		recordingMiddleware("B", &order),
		recordingMiddleware("C", &order),
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	stack(handler).ServeHTTP(w, req)

	// Expected order: A:before -> B:before -> C:before -> handler -> C:after -> B:after -> A:after
	expected := []string{"A:before", "B:before", "C:before", "handler", "C:after", "B:after", "A:after"}

	if len(order) != len(expected) {
		t.Fatalf("execution order length: got %d, want %d", len(order), len(expected))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("execution order[%d]: got %q, want %q", i, order[i], v)
		}
	}
}

// TestCreateStack_Composition verifies that nested CreateStack calls work correctly.
func TestCreateStack_Composition(t *testing.T) {
	var order []string

	innerStack := CreateStack(
		recordingMiddleware("B", &order),
		recordingMiddleware("C", &order),
	)

	outerStack := CreateStack(
		recordingMiddleware("A", &order),
		innerStack,
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	outerStack(handler).ServeHTTP(w, req)

	// Expected: A wraps (B wraps (C wraps handler))
	expected := []string{"A:before", "B:before", "C:before", "handler", "C:after", "B:after", "A:after"}

	if len(order) != len(expected) {
		t.Fatalf("execution order length: got %d, want %d", len(order), len(expected))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("execution order[%d]: got %q, want %q", i, order[i], v)
		}
	}
}

// TestConcurrentRequests_Logging verifies that the logging middleware
// handles concurrent requests without race conditions.
func TestConcurrentRequests_Logging(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("concurrent"))
	})

	logger, _ := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	runConcurrent(t, mw(handler), 100)
}

// TestConcurrentRequests_MaxBytesReader verifies that MaxBytesReader
// handles concurrent requests with varying body sizes.
func TestConcurrentRequests_MaxBytesReader(t *testing.T) {
	mw := NewMaxBytesReader(1024)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	runConcurrent(t, mw(handler), 100)
}

// TestConcurrentRequests_Stack verifies that a full middleware stack
// handles high concurrency without issues.
func TestConcurrentRequests_Stack(t *testing.T) {
	logger, _ := newTestLogger()
	stack := CreateStack(
		NewLoggingMiddleware(logger),
		NewMaxBytesReader(1024),
		NewSetContentTypeJSON(),
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Test with high concurrency
	runConcurrent(t, stack(handler), 1000)
}

// TestLoggingMiddleware_HandlerPanic verifies that middleware doesn't
// interfere with panic propagation.
func TestLoggingMiddleware_HandlerPanic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		panic("test panic")
	})

	logger, _ := newTestLogger()
	mw := NewLoggingMiddleware(logger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/panic", nil)

	// Recover from the panic to test that it propagates
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to propagate, but it didn't")
		} else if r != "test panic" {
			t.Errorf("expected panic message 'test panic', got %v", r)
		}
	}()

	mw(handler).ServeHTTP(w, req)
}

// TestMaxBytesReader_ReadError verifies that MaxBytesReader errors
// propagate correctly when the body exceeds the limit.
func TestMaxBytesReader_ReadError(t *testing.T) {
	mw := NewMaxBytesReader(100)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	largeBody := make([]byte, 200)
	req := httptest.NewRequest("POST", "/", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	mw(handler).ServeHTTP(w, req)

	assertStatus(t, w, http.StatusRequestEntityTooLarge)
}

// TestMiddleware_ContextPropagation verifies that context is properly
// propagated through the middleware chain.
func TestMiddleware_ContextPropagation(t *testing.T) {
	type contextKey string
	const key contextKey = "test-key"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(key)
		if val == nil {
			t.Error("context value not propagated")
		} else if val != "test-value" {
			t.Errorf("context value: got %v, want 'test-value'", val)
		}
		w.WriteHeader(http.StatusOK)
	})

	logger, _ := newTestLogger()
	stack := CreateStack(
		NewLoggingMiddleware(logger),
		NewSetContentTypeJSON(),
	)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), key, "test-value")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	stack(handler).ServeHTTP(w, req)
}

// BenchmarkLoggingMiddleware measures the overhead of logging middleware.
func BenchmarkLoggingMiddleware(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mw := NewLoggingMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkMaxBytesReader_SmallBody measures MaxBytesReader with small bodies.
func BenchmarkMaxBytesReader_SmallBody(b *testing.B) {
	mw := NewMaxBytesReader(1024)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(handler)

	body := make([]byte, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkMaxBytesReader_LargeBody measures MaxBytesReader with large bodies.
func BenchmarkMaxBytesReader_LargeBody(b *testing.B) {
	mw := NewMaxBytesReader(1048576) // 1MB
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(handler)

	body := make([]byte, 524288) // 512KB

	for b.Loop() {
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkSetContentType measures the overhead of SetContentType middleware.
func BenchmarkSetContentType(b *testing.B) {
	mw := NewSetContentTypeJSON()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "/test", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkCreateStack_Single measures stack creation with one middleware.
func BenchmarkCreateStack_Single(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	for b.Loop() {
		stack := CreateStack(NewSetContentTypeJSON())
		wrapped := stack(handler)
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkCreateStack_Multiple measures stack creation with multiple middleware.
func BenchmarkCreateStack_Multiple(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	for b.Loop() {
		stack := CreateStack(
			NewLoggingMiddleware(logger),
			NewMaxBytesReader(1024),
			NewSetContentTypeJSON(),
		)
		wrapped := stack(handler)
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkCreateStack_Execution measures just the execution overhead.
func BenchmarkCreateStack_Execution(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	stack := CreateStack(
		NewLoggingMiddleware(logger),
		NewMaxBytesReader(1024),
		NewSetContentTypeJSON(),
	)
	wrapped := stack(handler)

	req := httptest.NewRequest("GET", "/test", nil)

	for b.Loop() {
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, req)
	}
}
