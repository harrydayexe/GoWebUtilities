// Package config provides environment-based configuration management with validation.
//
// The package offers a type-safe way to parse configuration from environment variables
// using the Validator interface. It includes a ServerConfig implementation for common
// HTTP server settings (port, timeouts, environment) with built-in validation.
//
// Configuration structs must implement the Validator interface so that semantic
// constraints (e.g. valid environment names) are checked after the raw environment
// variables have been parsed. ParseConfig handles both steps and returns a combined
// error so callers can decide how to react — log.Fatal, a fallback config, etc.
//
// Example usage:
//
//	cfg, err := config.ParseConfig[config.ServerConfig]()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Server running on port %d in %s environment\n", cfg.Port, cfg.Environment)
package config
