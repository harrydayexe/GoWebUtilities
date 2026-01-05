package config

// Validator is an interface for configuration types that support validation.
// Configuration structs should implement this interface to enable validation
// of their fields after parsing from environment variables.
type Validator interface {
	// Validate checks the configuration for invalid values or inconsistencies.
	// It returns an error describing what is invalid, or nil if the configuration is valid.
	Validate() error
}
