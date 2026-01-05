package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Environment defines which environment the application is running in.
// Valid values are Local, Test, and Production.
type Environment string

// String returns the string representation of the Environment.
func (e Environment) String() string {
	return string(e)
}

const (
	// Local represents a local development environment.
	Local Environment = "local"
	// Test represents a testing environment.
	Test Environment = "test"
	// Production represents a production environment.
	Production Environment = "production"
)

// ServerConfig holds the configuration for an HTTP server.
// All fields are populated from environment variables with sensible defaults.
type ServerConfig struct {
	// Environment specifies the runtime environment (local, test, or production).
	// Defaults to "local" if ENVIRONMENT is not set.
	Environment Environment `env:"ENVIRONMENT" envDefault:"local"`
	// VerboseMode enables verbose logging when true.
	// Defaults to false if VERBOSE is not set.
	VerboseMode bool `env:"VERBOSE" envDefault:"false"`
	// Port is the HTTP server port number.
	// Defaults to 8080 if PORT is not set.
	Port int `env:"PORT" envDefault:"8080"`
	// ReadTimeout is the maximum duration in seconds for reading the entire request.
	// Defaults to 15 seconds if READ_TIMEOUT is not set.
	ReadTimeout int `env:"READ_TIMEOUT" envDefault:"15"`
	// WriteTimeout is the maximum duration in seconds for writing the response.
	// Defaults to 15 seconds if WRITE_TIMEOUT is not set.
	WriteTimeout int `env:"WRITE_TIMEOUT" envDefault:"15"`
	// IdleTimeout is the maximum duration in seconds to wait for the next request
	// when keep-alives are enabled. Defaults to 60 seconds if IDLE_TIMEOUT is not set.
	IdleTimeout int `env:"IDLE_TIMEOUT" envDefault:"60"`
}

// Validate checks that the ServerConfig has valid values.
// Currently validates that Environment is one of Local, Test, or Production.
// Returns an error if validation fails, nil otherwise.
func (c ServerConfig) Validate() error {
	switch c.Environment {
	case Local, Test, Production:
		return nil
	default:
		return fmt.Errorf("invalid environment: %s (must be local, test or production)", c.Environment)
	}
}

// ParseConfig parses environment variables into a configuration struct of type C
// and validates the result. The type parameter C must implement the Validator interface.
// Returns an error if parsing or validation fails, allowing the caller to decide how to handle it.
//
// Example:
//
//	cfg, err := ParseConfig[ServerConfig]()
//	if err != nil {
//		log.Fatal(err)
//	}
func ParseConfig[C Validator]() (C, error) {
	var zero C
	cfg, err := env.ParseAs[C]()
	if err != nil {
		return zero, fmt.Errorf("failed to parse config from environment: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return zero, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}
