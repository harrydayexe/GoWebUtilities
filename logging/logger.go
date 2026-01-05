package logging

import (
	"log/slog"
	"os"

	"github.com/harrydayexe/GoWebUtilities/config"
)

// SetDefaultLogger sets the default slog logger to be used in the application
// based on the server config provided
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
