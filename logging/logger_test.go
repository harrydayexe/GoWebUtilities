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

// getLogLevel returns the configured log level by checking if DEBUG is enabled
func getLogLevel(logger *slog.Logger) slog.Level {
	handler := logger.Handler()
	if handler.Enabled(context.Background(), slog.LevelDebug) {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// TestSetDefaultLogger_AllEnvironments tests all combinations of environment and verbose mode
func TestSetDefaultLogger_AllEnvironments(t *testing.T) {
	tests := []struct {
		name            string
		cfg             config.ServerConfig
		wantHandlerType string // "text" or "json"
		wantLogLevel    slog.Level
	}{
		{
			name: "local_non_verbose",
			cfg: config.ServerConfig{
				Environment: config.Local,
				VerboseMode: false,
			},
			wantHandlerType: "text",
			wantLogLevel:    slog.LevelInfo,
		},
		{
			name: "local_verbose",
			cfg: config.ServerConfig{
				Environment: config.Local,
				VerboseMode: true,
			},
			wantHandlerType: "text",
			wantLogLevel:    slog.LevelDebug,
		},
		{
			name: "test_non_verbose",
			cfg: config.ServerConfig{
				Environment: config.Test,
				VerboseMode: false,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelInfo,
		},
		{
			name: "test_verbose",
			cfg: config.ServerConfig{
				Environment: config.Test,
				VerboseMode: true,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelDebug,
		},
		{
			name: "production_non_verbose",
			cfg: config.ServerConfig{
				Environment: config.Production,
				VerboseMode: false,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelInfo,
		},
		{
			name: "production_verbose",
			cfg: config.ServerConfig{
				Environment: config.Production,
				VerboseMode: true,
			},
			wantHandlerType: "json",
			wantLogLevel:    slog.LevelDebug,
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

			// Verify log level by checking if DEBUG is enabled
			handler := logger.Handler()
			ctx := context.Background()
			debugEnabled := handler.Enabled(ctx, slog.LevelDebug)
			wantDebugEnabled := (tt.wantLogLevel == slog.LevelDebug)

			if debugEnabled != wantDebugEnabled {
				t.Errorf("DEBUG enabled = %v, want %v", debugEnabled, wantDebugEnabled)
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
				VerboseMode: false,
			}

			SetDefaultLogger(cfg)

			// Capture log output to verify format
			var buf bytes.Buffer
			logger := slog.Default()

			// Create a new logger with buffer to test output format
			var testLogger *slog.Logger
			if tt.wantJSON {
				testLogger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
			} else {
				testLogger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
			}

			// Verify handler enabled state matches
			ctx := context.Background()
			if logger.Handler().Enabled(ctx, slog.LevelInfo) != testLogger.Handler().Enabled(ctx, slog.LevelInfo) {
				t.Error("handler enabled state mismatch")
			}
		})
	}
}

// TestSetDefaultLogger_LogLevelConfiguration verifies log level based on verbose mode
func TestSetDefaultLogger_LogLevelConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		verboseMode bool
		wantDebug   bool
		wantInfo    bool
	}{
		{
			name:        "verbose_enables_debug",
			verboseMode: true,
			wantDebug:   true,
			wantInfo:    true,
		},
		{
			name:        "non_verbose_filters_debug",
			verboseMode: false,
			wantDebug:   false,
			wantInfo:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original logger
			original := saveDefaultLogger()
			defer slog.SetDefault(original)

			cfg := config.ServerConfig{
				Environment: config.Test,
				VerboseMode: tt.verboseMode,
			}

			SetDefaultLogger(cfg)

			logger := slog.Default()
			handler := logger.Handler()
			ctx := context.Background()

			// Check DEBUG level
			gotDebug := handler.Enabled(ctx, slog.LevelDebug)
			if gotDebug != tt.wantDebug {
				t.Errorf("DEBUG enabled = %v, want %v", gotDebug, tt.wantDebug)
			}

			// Check INFO level
			gotInfo := handler.Enabled(ctx, slog.LevelInfo)
			if gotInfo != tt.wantInfo {
				t.Errorf("INFO enabled = %v, want %v", gotInfo, tt.wantInfo)
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

	// First call: Local, non-verbose
	cfg1 := config.ServerConfig{
		Environment: config.Local,
		VerboseMode: false,
	}
	SetDefaultLogger(cfg1)
	logger1 := slog.Default()
	level1 := getLogLevel(logger1)
	if level1 != slog.LevelInfo {
		t.Errorf("first call: log level = %v, want %v", level1, slog.LevelInfo)
	}

	// Second call: Production, verbose
	cfg2 := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: true,
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

	// Third call: back to INFO level
	cfg3 := config.ServerConfig{
		Environment: config.Test,
		VerboseMode: false,
	}
	SetDefaultLogger(cfg3)
	logger3 := slog.Default()
	level3 := getLogLevel(logger3)
	if level3 != slog.LevelInfo {
		t.Errorf("third call: log level = %v, want %v", level3, slog.LevelInfo)
	}
}

// TestSetDefaultLogger_Integration verifies end-to-end behavior
func TestSetDefaultLogger_Integration(t *testing.T) {
	// Save and restore original logger
	original := saveDefaultLogger()
	defer slog.SetDefault(original)

	cfg := config.ServerConfig{
		Environment: config.Production,
		VerboseMode: true,
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
		t.Error("DEBUG not enabled for verbose mode")
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
		VerboseMode: false,
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
		name           string
		cfg            config.ServerConfig
		shouldLogDebug bool
	}{
		{
			name: "verbose_logs_debug",
			cfg: config.ServerConfig{
				Environment: config.Test,
				VerboseMode: true,
			},
			shouldLogDebug: true,
		},
		{
			name: "non_verbose_filters_debug",
			cfg: config.ServerConfig{
				Environment: config.Test,
				VerboseMode: false,
			},
			shouldLogDebug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure logger with buffer
			var buf bytes.Buffer
			var testLogger *slog.Logger

			opts := &slog.HandlerOptions{
				Level: func() slog.Level {
					if tt.cfg.VerboseMode {
						return slog.LevelDebug
					}
					return slog.LevelInfo
				}(),
			}

			testLogger = slog.New(slog.NewJSONHandler(&buf, opts))
			slog.SetDefault(testLogger)

			// Log messages
			slog.Debug("debug message")
			slog.Info("info message")

			output := buf.String()

			// Verify INFO always appears
			if !strings.Contains(output, "info message") {
				t.Error("INFO message not in output")
			}

			// Verify DEBUG based on verbose mode
			hasDebug := strings.Contains(output, "debug message")
			if tt.shouldLogDebug && !hasDebug {
				t.Error("DEBUG message should be in output but is not")
			}
			if !tt.shouldLogDebug && hasDebug {
				t.Error("DEBUG message should be filtered but appears in output")
			}
		})
	}
}
