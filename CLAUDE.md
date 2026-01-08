# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Maintaining This File

**IMPORTANT**: After creating, modifying, or reviewing any new modules or packages in this codebase, you MUST update the "Package Structure" section under Architecture to document them. This is a required step, not optional.

Actions that require updating this file:
- Creating a new package or module
- Adding significant functionality to an existing package
- Reviewing or documenting a module that isn't listed in the Package Structure section

Failing to update this file means future Claude Code sessions will lack critical context about the codebase structure.

## Project Overview

GoWebUtilities is a Go library providing reusable  utilities for web applications. It follows Go's standard `net/http` middleware pattern where middleware are functions that wrap `http.Handler` instances.

## Architecture

### Middleware Pattern

The core abstraction is defined in `middleware/middleware.go:8`:

```go
type Middleware func(h http.Handler) http.Handler
```

All middleware in this library follow this signature. Middleware can be composed using `CreateStack()`, which applies middleware in the order provided (first middleware = outermost wrapper = executed first on requests).

### Package Structure

- `middleware/` - Contains all middleware implementations
  - `middleware.go` - Core types and `CreateStack()` composition function
  - `logging.go` - Request logging with slog integration, uses `wrappedWriter` to capture status codes
  - `maxBytesReader.go` - Request body size limiting (default 1MB)
  - `setContentType.go` - Response Content-Type header setting
  - `middleware_example_test.go` - Example functions demonstrating middleware usage following Go's standard example conventions

- `config/` - Environment-based configuration management with validation
  - `doc.go` - Package documentation
  - `validator.go` - `Validator` interface for configuration types that support validation
  - `serverConfig.go` - `ServerConfig` implementation for HTTP server settings (port, timeouts, environment) and `ParseConfig[C Validator]()` generic function for parsing and validating any config type from environment variables
  - Uses `github.com/caarlos0/env/v11` for environment variable parsing
  - Supports three environments: Local, Test, Production
  - All configuration parsing includes automatic validation; returns errors for invalid config allowing callers to decide how to handle failures

- `logging/` - Centralized logger configuration for structured logging
  - `doc.go` - Package documentation
  - `logger.go` - `SetDefaultLogger()` configures global slog logger based on environment
  - Integrates with config package for environment-based setup
  - Selects handler type (Text for Local, JSON for Test/Production)
  - Configures log level (DEBUG if verbose, INFO otherwise)
  - Sets global default via slog.SetDefault()
  - NOT safe for concurrent use - call once during initialization

- `server/` - HTTP server creation and lifecycle management
  - `doc.go` - Package documentation with usage examples
  - `server.go` - `NewServerWithConfig()` creates http.Server instances configured from environment variables via config.ServerConfig
  - `run.go` - `Run()` function providing complete server lifecycle management with graceful shutdown
  - Integrates with config package for environment-based configuration (port, timeouts)
  - Handles interrupt signals (SIGINT) for graceful shutdown with 10-second timeout
  - Logs server lifecycle events using structured logging (slog)
  - Safe for concurrent use

## Development Commands

### Building and Testing

```bash
# Build the module
go build ./...

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests for specific package
go test ./middleware

# Run examples
go test -v ./middleware -run Example

# Verify module dependencies
go mod verify
go mod tidy
```

### Releasing

This project uses goreleaser for automated releases:

- Releases are triggered by pushing tags matching `v*` pattern
- Tag format: `v1.2.3` (semantic versioning)
- The `.goreleaser.yaml` is configured to skip building binaries (library-only project)
- GitHub Actions workflow (`.github/workflows/release.yml`) handles automated releases

To create a release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

## Code Conventions

### Middleware Constructor Pattern

Middleware that require configuration use the `New*` constructor pattern:

```go
func NewLoggingMiddleware(logger *slog.Logger) Middleware
func NewMaxBytesReader(maxBytes int64) Middleware
func NewSetContentType(contentType string) Middleware
```

This allows parameterized middleware while maintaining the standard `Middleware` signature for composition.

### Configuration Pattern

Configuration structs should:
- Implement the `config.Validator` interface
- Use struct tags for environment variable mapping: `env:"VAR_NAME" envDefault:"default_value"`
- Be parsed using the generic `config.ParseConfig[T]()` function which handles parsing and validation
- Handle errors returned by `ParseConfig()` appropriately (e.g., log.Fatal, fallback config, retry)

Example:
```go
type ServerConfig struct {
    Port int `env:"PORT" envDefault:"8080"`
}

func (c ServerConfig) Validate() error {
    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("invalid port: %d", c.Port)
    }
    return nil
}

// Usage - fail-fast approach
cfg, err := config.ParseConfig[ServerConfig]()
if err != nil {
    log.Fatal(err)
}

// Alternative - graceful error handling
cfg, err := config.ParseConfig[ServerConfig]()
if err != nil {
    log.Printf("Config error: %v, using defaults", err)
    cfg = ServerConfig{Port: 8080} // fallback
}
```

### Logging

- Use structured logging with `log/slog`
- Request start logged at DEBUG level
- Request completion logged at INFO level with method, path, status, and duration

### Examples and Testing

The middleware package includes idiomatic Go examples in `middleware_example_test.go`:

- **Package**: `middleware_test` (black-box testing from external perspective)
- **Example Functions**: Follow Go's `Example*` naming convention
- **Testable Examples**: Use `// Output:` comments where output is deterministic
- **Coverage**: Examples demonstrate all exported middleware functions:
  - `ExampleNewSetContentTypeJSON()` - Basic middleware usage
  - `ExampleNewMaxBytesReader()` - Request size limiting
  - `ExampleNewLoggingMiddleware()` - Structured logging integration
  - `ExampleCreateStack()` - Middleware composition
  - `ExampleCreateStack_complete()` - Full HTTP server setup
  - `ExampleCreateStack_jsonAPI()` - Realistic JSON API scenario

Examples serve as both documentation (visible in `godoc` and `pkg.go.dev`) and executable tests.

### Standard Library

The library endevours to use the standard library as much as possible, for example `net/http` for routing

