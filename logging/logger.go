package logging

import (
	"log/slog"
	"os"

	"github.com/harrydayexe/GoWebUtilities/config"
)

// SetDefaultLogger configures the default slog logger based on the provided ServerConfig.
// It sets the global default logger used by slog.Info, slog.Debug, and other top-level
// slog functions.
//
// The function configures two aspects:
//   - Log level: DEBUG if cfg.VerboseMode is true, otherwise INFO
//   - Handler type: Text for Local environment, JSON for Test/Production
//
// Log handlers write to os.Stdout. All log output includes timestamps and context fields.
//
// This function is NOT safe for concurrent use and modifies global state via slog.SetDefault.
// Call it once during application initialization (e.g., in main(), before starting the server)
// before any goroutines that use logging are spawned.
//
// Example:
//
//	cfg, _ := config.ParseConfig[config.ServerConfig]()
//	logging.SetDefaultLogger(cfg)
//	slog.Info("server starting", "environment", cfg.Environment)
func SetDefaultLogger(cfg config.ServerConfig) {
	var logger *slog.Logger
	var handlerOptions slog.HandlerOptions

	if cfg.VerboseMode {
		handlerOptions = slog.HandlerOptions{Level: slog.LevelDebug}
	} else {
		handlerOptions = slog.HandlerOptions{Level: slog.LevelInfo}
	}

	if cfg.Environment == config.Local {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &handlerOptions))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &handlerOptions))
	}

	slog.SetDefault(logger)
}
