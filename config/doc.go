// Package config provides environment-based configuration management with validation.
//
// The package offers a type-safe way to parse configuration from environment variables
// using the Validator interface. It includes a ServerConfig implementation for common
// HTTP server settings (port, timeouts, environment) with built-in validation.
//
// Example usage:
//
//	cfg := config.ParseConfig[config.ServerConfig]()
//	fmt.Printf("Server running on port %d in %s environment\n", cfg.Port, cfg.Environment)
package config
