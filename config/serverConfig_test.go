package config

import (
	"fmt"
	"testing"
)

func TestEnvironment_String(t *testing.T) {
	tests := []struct {
		name string
		env  Environment
		want string
	}{
		{
			name: "Local environment",
			env:  Local,
			want: "local",
		},
		{
			name: "Test environment",
			env:  Test,
			want: "test",
		},
		{
			name: "Production environment",
			env:  Production,
			want: "production",
		},
		{
			name: "Custom environment value",
			env:  Environment("custom"),
			want: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.env.String(); got != tt.want {
				t.Errorf("Environment.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid Local environment",
			config: ServerConfig{
				Environment: Local,
			},
			wantErr: false,
		},
		{
			name: "Valid Test environment",
			config: ServerConfig{
				Environment: Test,
			},
			wantErr: false,
		},
		{
			name: "Valid Production environment",
			config: ServerConfig{
				Environment: Production,
			},
			wantErr: false,
		},
		{
			name: "Invalid environment - empty",
			config: ServerConfig{
				Environment: "",
			},
			wantErr: true,
			errMsg:  "invalid environment:  (must be local, test or production)",
		},
		{
			name: "Invalid environment - unknown value",
			config: ServerConfig{
				Environment: "staging",
			},
			wantErr: true,
			errMsg:  "invalid environment: staging (must be local, test or production)",
		},
		{
			name: "Invalid environment - wrong case",
			config: ServerConfig{
				Environment: "LOCAL",
			},
			wantErr: true,
			errMsg:  "invalid environment: LOCAL (must be local, test or production)",
		},
		{
			name: "Valid config with all fields populated",
			config: ServerConfig{
				Environment:  Production,
				VerboseMode:  true,
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
				IdleTimeout:  120,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("ServerConfig.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestParseConfig_ServerConfig_Defaults(t *testing.T) {
	// Clear all relevant environment variables to test defaults
	envVars := []string{"ENVIRONMENT", "VERBOSE", "PORT", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT"}
	for _, v := range envVars {
		t.Setenv(v, "")
	}

	cfg, err := ParseConfig[ServerConfig]()
	if err != nil {
		t.Fatalf("ParseConfig() with defaults should not error, got: %v", err)
	}

	// Check default values
	if cfg.Environment != Local {
		t.Errorf("Default Environment = %v, want %v", cfg.Environment, Local)
	}
	if cfg.VerboseMode != false {
		t.Errorf("Default VerboseMode = %v, want %v", cfg.VerboseMode, false)
	}
	if cfg.Port != 8080 {
		t.Errorf("Default Port = %v, want %v", cfg.Port, 8080)
	}
	if cfg.ReadTimeout != 15 {
		t.Errorf("Default ReadTimeout = %v, want %v", cfg.ReadTimeout, 15)
	}
	if cfg.WriteTimeout != 15 {
		t.Errorf("Default WriteTimeout = %v, want %v", cfg.WriteTimeout, 15)
	}
	if cfg.IdleTimeout != 60 {
		t.Errorf("Default IdleTimeout = %v, want %v", cfg.IdleTimeout, 60)
	}
}

func TestParseConfig_ServerConfig_CustomValues(t *testing.T) {
	// Set custom environment variables
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("VERBOSE", "true")
	t.Setenv("PORT", "3000")
	t.Setenv("READ_TIMEOUT", "30")
	t.Setenv("WRITE_TIMEOUT", "30")
	t.Setenv("IDLE_TIMEOUT", "120")

	cfg, err := ParseConfig[ServerConfig]()
	if err != nil {
		t.Fatalf("ParseConfig() with valid custom values should not error, got: %v", err)
	}

	// Verify custom values are set correctly
	if cfg.Environment != Production {
		t.Errorf("Environment = %v, want %v", cfg.Environment, Production)
	}
	if cfg.VerboseMode != true {
		t.Errorf("VerboseMode = %v, want %v", cfg.VerboseMode, true)
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %v, want %v", cfg.Port, 3000)
	}
	if cfg.ReadTimeout != 30 {
		t.Errorf("ReadTimeout = %v, want %v", cfg.ReadTimeout, 30)
	}
	if cfg.WriteTimeout != 30 {
		t.Errorf("WriteTimeout = %v, want %v", cfg.WriteTimeout, 30)
	}
	if cfg.IdleTimeout != 120 {
		t.Errorf("IdleTimeout = %v, want %v", cfg.IdleTimeout, 120)
	}
}

func TestParseConfig_ValidationError(t *testing.T) {
	// Set an invalid environment value
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("PORT", "8080")

	cfg, err := ParseConfig[ServerConfig]()
	if err == nil {
		t.Fatal("ParseConfig() with invalid environment should return error, got nil")
	}

	// Verify error message contains validation error
	expectedMsg := "config validation failed"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Error message should contain %q, got: %v", expectedMsg, err.Error())
	}

	// Verify the underlying validation error is wrapped
	expectedValidationMsg := "invalid environment: staging"
	if !contains(err.Error(), expectedValidationMsg) {
		t.Errorf("Error message should contain %q, got: %v", expectedValidationMsg, err.Error())
	}

	// Verify zero value is returned on error
	if cfg.Port != 0 {
		t.Errorf("ParseConfig() should return zero value on error, got Port = %d", cfg.Port)
	}
}

func TestParseConfig_ParsingError(t *testing.T) {
	// Set an invalid type for PORT (should be int)
	t.Setenv("PORT", "not-a-number")
	t.Setenv("ENVIRONMENT", "local")

	cfg, err := ParseConfig[ServerConfig]()
	if err == nil {
		t.Fatal("ParseConfig() with invalid PORT should return error, got nil")
	}

	// Verify error message contains parsing error
	expectedMsg := "failed to parse config from environment"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Error message should contain %q, got: %v", expectedMsg, err.Error())
	}

	// Verify zero value is returned on error
	if cfg.Port != 0 {
		t.Errorf("ParseConfig() should return zero value on error, got Port = %d", cfg.Port)
	}
}

// TestParseConfig_CustomValidator demonstrates ParseConfig with a custom type
func TestParseConfig_CustomValidator(t *testing.T) {
	// Test type that implements Validator
	type TestConfig struct {
		Value int `env:"TEST_VALUE" envDefault:"42"`
	}

	// Implement Validator interface
	testValidate := func(tc TestConfig) error {
		if tc.Value < 0 {
			return fmt.Errorf("value must be non-negative")
		}
		return nil
	}

	// We can't directly use ParseConfig with inline types, but we can verify
	// the pattern works by testing with ServerConfig and showing the interface
	t.Setenv("ENVIRONMENT", "test")

	cfg, err := ParseConfig[ServerConfig]()
	if err != nil {
		t.Fatalf("ParseConfig() should succeed, got error: %v", err)
	}

	if cfg.Environment != Test {
		t.Errorf("Environment = %v, want %v", cfg.Environment, Test)
	}

	// Demonstrate the validation function signature
	tc := TestConfig{Value: 42}
	if err := testValidate(tc); err != nil {
		t.Errorf("Validation should pass for valid config, got: %v", err)
	}

	tc.Value = -1
	if err := testValidate(tc); err == nil {
		t.Error("Validation should fail for negative value")
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestValidator_Interface verifies that ServerConfig implements Validator
func TestValidator_Interface(t *testing.T) {
	var _ Validator = ServerConfig{}
}
