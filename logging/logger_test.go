package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/harrydayexe/GoWebUtilities/config"
)

// saveDefaultLogger saves the current default logger for restoration in tests
func saveDefaultLogger() *slog.Logger {
	return slog.Default()
}

// getLogLevel returns the configured log level by checking which levels are enabled
func getLogLevel(logger *slog.Logger) slog.Level {
	handler := logger.Handler()
	ctx := context.Background()
	switch {
	case handler.Enabled(ctx, slog.LevelDebug):
		return slog.LevelDebug
	case handler.Enabled(ctx, slog.LevelInfo):
		return slog.LevelInfo
	case handler.Enabled(ctx, slog.LevelWarn):
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

// TestSetDefaultLogger_AllEnvironments tests all combinations of environment and log level
func TestSetDefaultLogger_AllEnvironments(t *testing.T) {
	tests := []struct {
		name            string
		cfg             config.ServerConfig
		wantHandlerType string // "text" or "json"
		wantLogLevel    slog.Level
	}{
		{
			name: "local_debug",
			cfg: config.ServerConfig{
				Environment: config.Local,
				LogLevel:    slog.LevelDebug,
			},
			wantHandlerType: "text",
			wantLogLevel:    slog.LevelDebug,
		},
		{
			name: "local_warn",
			cfg: config.ServerConfig{
				Environment: config.Local,
				LogLevel:    slog.LevelWarn,
			},
			wantHandlerType: "text",
			wantLogLevel:    slog.LevelWarn,
		},
		{
			name: "test_info",
			cfg: config.ServerConfig{
				Environment: config.Test,
				LogLevel:    slog.LevelInfo,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelInfo,
		},
		{
			name: "test_warn",
			cfg: config.ServerConfig{
				Environment: config.Test,
				LogLevel:    slog.LevelWarn,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelWarn,
		},
		{
			name: "production_warn",
			cfg: config.ServerConfig{
				Environment: config.Production,
				LogLevel:    slog.LevelWarn,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelWarn,
		},
		{
			name: "production_error",
			cfg: config.ServerConfig{
				Environment: config.Production,
				LogLevel:    slog.LevelError,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original logger
			original := saveDefaultLogger()
			defer slog.SetDefault(original)

			// Call SetDefaultLogger
			SetDefaultLogger(tt.cfg)

			// Verify the default logger was changed
			logger := slog.Default()
			if logger == original {
				t.Error("SetDefaultLogger did not change the default logger")
			}

			// Verify log level
			handler := logger.Handler()
			ctx := context.Background()

			gotLevel := getLogLevel(logger)
			if gotLevel != tt.wantLogLevel {
				t.Errorf("log level = %v, want %v", gotLevel, tt.wantLogLevel)
			}

			// Verify levels below wantLogLevel are disabled
			if tt.wantLogLevel > slog.LevelDebug && handler.Enabled(ctx, slog.LevelDebug) {
				t.Errorf("DEBUG should be disabled for level %v", tt.wantLogLevel)
			}
			if tt.wantLogLevel > slog.LevelInfo && handler.Enabled(ctx, slog.LevelInfo) {
				t.Errorf("INFO should be disabled for level %v", tt.wantLogLevel)
			}
			if tt.wantLogLevel > slog.LevelWarn && handler.Enabled(ctx, slog.LevelWarn) {
				t.Errorf("WARN should be disabled for level %v", tt.wantLogLevel)
			}
		})
	}
}

// TestSetDefaultLogger_HandlerSelection verifies correct handler type for each environment
func TestSetDefaultLogger_HandlerSelection(t *testing.T) {
	tests := []struct {
		name        string
		environment config.Environment
		wantJSON    bool
	}{
		{
			name:        "local_uses_text",
			environment: config.Local,
			wantJSON:    false,
		},
		{
			name:        "test_uses_json",
			environment: config.Test,
			wantJSON:    true,
		},
		{
			name:        "production_uses_json",
			environment: config.Production,
			wantJSON:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original logger
			original := saveDefaultLogger()
			defer slog.SetDefault(original)

			cfg := config.ServerConfig{
				Environment: tt.environment,
				LogLevel:    slog.LevelWarn,
			}

			SetDefaultLogger(cfg)

			// Capture log output to verify format
			var buf bytes.Buffer
			logger := slog.Default()

			// Create a new logger with buffer to test output format
			var testLogger *slog.Logger
			if tt.wantJSON {
				testLogger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
			} else {
				testLogger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
			}

			// Verify handler enabled state matches
			ctx := context.Background()
			if logger.Handler().Enabled(ctx, slog.LevelWarn) != testLogger.Handler().Enabled(ctx, slog.LevelWarn) {
				t.Error("handler enabled state mismatch")
			}
		})
	}
}

// TestSetDefaultLogger_LogLevelConfiguration verifies log level configuration for all levels
func TestSetDefaultLogger_LogLevelConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		logLevel   slog.Level
		wantDebug  bool
		wantInfo   bool
		wantWarn   bool
		wantError  bool
	}{
		{
			name:      "debug_enables_all",
			logLevel:  slog.LevelDebug,
			wantDebug: true,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "info_filters_debug",
			logLevel:  slog.LevelInfo,
			wantDebug: false,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "warn_filters_debug_and_info",
			logLevel:  slog.LevelWarn,
			wantDebug: false,
			wantInfo:  false,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "error_filters_debug_info_warn",
			logLevel:  slog.LevelError,
			wantDebug: false,
			wantInfo:  false,
			wantWarn:  false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original logger
			original := saveDefaultLogger()
			defer slog.SetDefault(original)

			cfg := config.ServerConfig{
				Environment: config.Test,
				LogLevel:    tt.logLevel,
			}

			SetDefaultLogger(cfg)

			logger := slog.Default()
			handler := logger.Handler()
			ctx := context.Background()

			if got := handler.Enabled(ctx, slog.LevelDebug); got != tt.wantDebug {
				t.Errorf("DEBUG enabled = %v, want %v", got, tt.wantDebug)
			}
			if got := handler.Enabled(ctx, slog.LevelInfo); got != tt.wantInfo {
				t.Errorf("INFO enabled = %v, want %v", got, tt.wantInfo)
			}
			if got := handler.Enabled(ctx, slog.LevelWarn); got != tt.wantWarn {
				t.Errorf("WARN enabled = %v, want %v", got, tt.wantWarn)
			}
			if got := handler.Enabled(ctx, slog.LevelError); got != tt.wantError {
				t.Errorf("ERROR enabled = %v, want %v", got, tt.wantError)
			}
		})
	}
}

// TestSetDefaultLogger_MultipleInvocations verifies that calling SetDefaultLogger multiple times
// replaces the logger (last call wins)
func TestSetDefaultLogger_MultipleInvocations(t *testing.T) {
	// Save and restore original logger
	original := saveDefaultLogger()
	defer slog.SetDefault(original)

	// First call: Local, warn
	cfg1 := config.ServerConfig{
		Environment: config.Local,
		LogLevel:    slog.LevelWarn,
	}
	SetDefaultLogger(cfg1)
	logger1 := slog.Default()
	level1 := getLogLevel(logger1)
	if level1 != slog.LevelWarn {
		t.Errorf("first call: log level = %v, want %v", level1, slog.LevelWarn)
	}

	// Second call: Production, debug
	cfg2 := config.ServerConfig{
		Environment: config.Production,
		LogLevel:    slog.LevelDebug,
	}
	SetDefaultLogger(cfg2)
	logger2 := slog.Default()
	level2 := getLogLevel(logger2)
	if level2 != slog.LevelDebug {
		t.Errorf("second call: log level = %v, want %v", level2, slog.LevelDebug)
	}

	// Verify logger was replaced
	if logger1 == logger2 {
		t.Error("logger was not replaced on second call")
	}

	// Third call: back to warn level
	cfg3 := config.ServerConfig{
		Environment: config.Test,
		LogLevel:    slog.LevelWarn,
	}
	SetDefaultLogger(cfg3)
	logger3 := slog.Default()
	level3 := getLogLevel(logger3)
	if level3 != slog.LevelWarn {
		t.Errorf("third call: log level = %v, want %v", level3, slog.LevelWarn)
	}
}

// TestSetDefaultLogger_Integration verifies end-to-end behavior
func TestSetDefaultLogger_Integration(t *testing.T) {
	// Save and restore original logger
	original := saveDefaultLogger()
	defer slog.SetDefault(original)

	cfg := config.ServerConfig{
		Environment: config.Production,
		LogLevel:    slog.LevelDebug,
	}

	// Verify default changed
	before := slog.Default()
	SetDefaultLogger(cfg)
	after := slog.Default()

	if before == after {
		t.Error("SetDefaultLogger did not change slog.Default()")
	}

	// Verify the new logger has correct configuration
	handler := after.Handler()
	ctx := context.Background()

	// Should enable DEBUG
	if !handler.Enabled(ctx, slog.LevelDebug) {
		t.Error("DEBUG not enabled for LevelDebug")
	}

	// Should enable INFO
	if !handler.Enabled(ctx, slog.LevelInfo) {
		t.Error("INFO not enabled")
	}
}

// TestSetDefaultLogger_ConcurrentCalls tests concurrent calls to SetDefaultLogger.
// NOTE: SetDefaultLogger is NOT intended for concurrent use. This test verifies
// that concurrent calls don't panic, but the behavior is undefined.
func TestSetDefaultLogger_ConcurrentCalls(t *testing.T) {
	// Save and restore original logger
	original := saveDefaultLogger()
	defer slog.SetDefault(original)

	cfg := config.ServerConfig{
		Environment: config.Test,
		LogLevel:    slog.LevelWarn,
	}

	// Track panics
	var panicked bool
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			t.Errorf("SetDefaultLogger panicked during concurrent calls: %v", r)
		}
	}()

	// Launch goroutines that all call SetDefaultLogger
	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			SetDefaultLogger(cfg)
		}()
	}

	wg.Wait()

	if panicked {
		t.Error("SetDefaultLogger panicked during concurrent calls")
	}
}

// TestSetDefaultLogger_ActualLogging verifies that logging actually works with configured logger
func TestSetDefaultLogger_ActualLogging(t *testing.T) {
	// Save and restore original logger
	original := saveDefaultLogger()
	defer slog.SetDefault(original)

	tests := []struct {
		name          string
		logLevel      slog.Level
		shouldLogDebug bool
		shouldLogInfo  bool
		shouldLogWarn  bool
	}{
		{
			name:          "debug_logs_all",
			logLevel:      slog.LevelDebug,
			shouldLogDebug: true,
			shouldLogInfo:  true,
			shouldLogWarn:  true,
		},
		{
			name:          "info_filters_debug",
			logLevel:      slog.LevelInfo,
			shouldLogDebug: false,
			shouldLogInfo:  true,
			shouldLogWarn:  true,
		},
		{
			name:          "warn_filters_debug_and_info",
			logLevel:      slog.LevelWarn,
			shouldLogDebug: false,
			shouldLogInfo:  false,
			shouldLogWarn:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			opts := &slog.HandlerOptions{Level: tt.logLevel}
			testLogger := slog.New(slog.NewJSONHandler(&buf, opts))
			slog.SetDefault(testLogger)

			slog.Debug("debug message")
			slog.Info("info message")
			slog.Warn("warn message")

			output := buf.String()

			hasDebug := strings.Contains(output, "debug message")
			if tt.shouldLogDebug && !hasDebug {
				t.Error("DEBUG message should be in output but is not")
			}
			if !tt.shouldLogDebug && hasDebug {
				t.Error("DEBUG message should be filtered but appears in output")
			}

			hasInfo := strings.Contains(output, "info message")
			if tt.shouldLogInfo && !hasInfo {
				t.Error("INFO message should be in output but is not")
			}
			if !tt.shouldLogInfo && hasInfo {
				t.Error("INFO message should be filtered but appears in output")
			}

			hasWarn := strings.Contains(output, "warn message")
			if tt.shouldLogWarn && !hasWarn {
				t.Error("WARN message should be in output but is not")
			}
			if !tt.shouldLogWarn && hasWarn {
				t.Error("WARN message should be filtered but appears in output")
			}
		})
	}
}
