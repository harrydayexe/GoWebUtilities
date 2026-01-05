package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/harrydayexe/GoWebUtilities/config"
)

func NewServerWithConfig(handler http.Handler) (*http.Server, error) {
	cfg, err := config.ParseConfig[config.ServerConfig]()
	if err != nil {
		return nil, fmt.Errorf("failed to create config from environment: %w", err)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	return httpServer, nil
}
