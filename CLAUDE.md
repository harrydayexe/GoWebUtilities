# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Maintaining This File

When working on new modules or components that are not yet documented in this file, update the Architecture section to include details about the new modules. This ensures future Claude Code sessions have accurate information about the codebase structure.

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

