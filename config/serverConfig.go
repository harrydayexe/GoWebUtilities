package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

// Environment defines which environment the application is running in
type Environment string

func (e Environment) String() string {
	return string(e)
}

const (
	Local      Environment = "local"
	Test       Environment = "test"
	Production Environment = "production"
)

// ServerConfig is a struct that holds the configuration for the server itself.
type ServerConfig struct {
	Environment  Environment `env:"ENVIRONMENT" envDefault:"local"`
	VerboseMode  bool        `env:"VERBOSE" envDefault:"false"`
	Port         int         `env:"PORT" envDefault:"8080"`
	ReadTimeout  int         `env:"READ_TIMEOUT" envDefault:"15"`
	WriteTimeout int         `env:"WRITE_TIMEOUT" envDefault:"15"`
	IdleTimeout  int         `env:"IDLE_TIMEOUT" envDefault:"60"`
}

func (c ServerConfig) Validate() error {
	switch c.Environment {
	case Local, Test, Production:
		return nil
	default:
		return fmt.Errorf("invalid environment: %s (must be local, test or production)", c.Environment)
	}
}

// ParseConfig creates a config object based on environment variables
func ParseConfig[C Validator]() C {
	cfg, err := env.ParseAs[C]()
	if err != nil {
		log.Fatal(err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	return cfg
}
