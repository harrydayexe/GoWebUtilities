package config_test

import (
	"fmt"
	"log"
	"os"

	"github.com/harrydayexe/GoWebUtilities/config"
)

// ExampleServerConfig demonstrates creating and validating a ServerConfig
// with default values.
func ExampleServerConfig() {
	cfg := config.ServerConfig{
		Environment:  config.Local,
		VerboseMode:  false,
		Port:         8080,
		ReadTimeout:  15,
		WriteTimeout: 15,
		IdleTimeout:  60,
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Verbose: %t\n", cfg.VerboseMode)

	// Output:
	// Environment: local
	// Port: 8080
	// Verbose: false
}

// ExampleServerConfig_Validate demonstrates validation of different
// environment values.
func ExampleServerConfig_Validate() {
	validConfig := config.ServerConfig{
		Environment: config.Production,
	}

	invalidConfig := config.ServerConfig{
		Environment: config.Environment("staging"),
	}

	// Valid configuration
	if err := validConfig.Validate(); err != nil {
		fmt.Printf("Valid config error: %v\n", err)
	} else {
		fmt.Println("Production config is valid")
	}

	// Invalid configuration
	if err := invalidConfig.Validate(); err != nil {
		fmt.Printf("Invalid config error: %v\n", err)
	}

	// Output:
	// Production config is valid
	// Invalid config error: invalid environment: staging (must be local, test or production)
}

// ExampleEnvironment demonstrates the Environment type and its constants.
func ExampleEnvironment() {
	environments := []config.Environment{
		config.Local,
		config.Test,
		config.Production,
	}

	for _, env := range environments {
		fmt.Printf("Environment: %s\n", env.String())
	}

	// Output:
	// Environment: local
	// Environment: test
	// Environment: production
}

// ExampleParseConfig demonstrates parsing configuration from environment
// variables with the generic ParseConfig function.
func ExampleParseConfig() {
	// Set environment variables for the example
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("PORT", "3000")
	os.Setenv("VERBOSE", "true")

	cfg, err := config.ParseConfig[config.ServerConfig]()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Environment: %s\n", cfg.Environment)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Verbose: %t\n", cfg.VerboseMode)

	// Output:
	// Environment: production
	// Port: 3000
	// Verbose: true
}

// ExampleParseConfig_errorHandling demonstrates custom error handling
// instead of using log.Fatal.
func ExampleParseConfig_errorHandling() {
	// Set an invalid environment value
	os.Setenv("ENVIRONMENT", "staging")

	cfg, err := config.ParseConfig[config.ServerConfig]()
	if err != nil {
		// Custom error handling - you decide what to do
		fmt.Printf("Configuration error: %v\n", err)
		fmt.Println("Using fallback configuration")
		// Could use default config, retry, etc.
		return
	}

	// This won't be reached in this example
	fmt.Printf("Loaded config: %v\n", cfg.Environment)

	// Output:
	// Configuration error: config validation failed: invalid environment: staging (must be local, test or production)
	// Using fallback configuration
}

// CustomConfig demonstrates implementing a custom configuration type
// with the Validator interface.
type CustomConfig struct {
	APIKey      string `env:"API_KEY"`
	MaxRequests int    `env:"MAX_REQUESTS" envDefault:"100"`
}

// Validate implements the config.Validator interface.
func (c CustomConfig) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API_KEY is required")
	}
	if c.MaxRequests < 1 {
		return fmt.Errorf("MAX_REQUESTS must be positive")
	}
	return nil
}

// ExampleParseConfig_customConfig demonstrates using ParseConfig with
// a custom configuration type that implements Validator.
func ExampleParseConfig_customConfig() {
	// Example of a custom config type
	customCfg := CustomConfig{
		APIKey:      "secret-key-123",
		MaxRequests: 100,
	}

	if err := customCfg.Validate(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("API Key set: %t\n", customCfg.APIKey != "")
	fmt.Printf("Max Requests: %d\n", customCfg.MaxRequests)

	// Output:
	// API Key set: true
	// Max Requests: 100
}

// ExampleValidator demonstrates the Validator interface usage.
func ExampleValidator() {
	// Any type implementing Validator can be used with ParseConfig
	var _ config.Validator = config.ServerConfig{}
	var _ config.Validator = CustomConfig{}

	fmt.Println("Both ServerConfig and CustomConfig implement Validator")

	// Output:
	// Both ServerConfig and CustomConfig implement Validator
}
