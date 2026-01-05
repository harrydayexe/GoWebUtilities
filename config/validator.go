package config

// Validator is an interface for configs that can be validated
type Validator interface {
	// Validate returns an error if the config is invalid, nil otherwise.
	Validate() error
}
