# GoWebUtilities

[![Go Reference](https://pkg.go.dev/badge/github.com/harrydayexe/GoWebUtilities.svg)](https://pkg.go.dev/github.com/harrydayexe/GoWebUtilities)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrydayexe/GoWebUtilities)](https://goreportcard.com/report/github.com/harrydayexe/GoWebUtilities)
[![Test](https://github.com/harrydayexe/GoWebUtilities/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/harrydayexe/GoWebUtilities/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library providing reusable utilities for building web applications.

## Installation

```bash
go get github.com/harrydayexe/GoWebUtilities
```

## Available Packages

### middleware

Composable HTTP middleware following Go's standard `net/http` pattern. All middleware share the type:

```go
type Middleware func(h http.Handler) http.Handler
```

Available middleware:

- **NewLoggingMiddleware** — structured request logging via `log/slog`, recording method, path, status code, and duration.
- **NewMaxBytesReader** — limits request body size to prevent resource exhaustion (defaults to 1 MB when 0 is passed).
- **NewSetContentType / NewSetContentTypeJSON** — sets the `Content-Type` response header for all responses.
- **NewStripHTMLExtension** — rewrites `.html` paths to clean URLs before routing (e.g. `/about.html` becomes `/about`; `/index.html` becomes `/`).

Use `CreateStack` to compose multiple middleware in order. The first argument is outermost and executes first on every request:

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/harrydayexe/GoWebUtilities/middleware"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status":"ok"}`))
    })

    stack := middleware.CreateStack(
        middleware.NewLoggingMiddleware(logger),
        middleware.NewMaxBytesReader(1024*1024), // 1 MB limit
        middleware.NewSetContentTypeJSON(),
    )

    http.ListenAndServe(":8080", stack(mux))
}
```

### config

Environment-based configuration management with validation. `ParseConfig` is a generic function that parses environment variables into any struct that implements the `Validator` interface and then validates the result. `ServerConfig` is the built-in implementation covering common HTTP server settings.

```go
cfg, err := config.ParseConfig[config.ServerConfig]()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("port %d, env %s\n", cfg.Port, cfg.Environment)
```

`ServerConfig` reads the following environment variables:

| Variable      | Default        | Description                                   |
|---------------|----------------|-----------------------------------------------|
| `PORT`        | `8080`         | HTTP listen port                              |
| `ENVIRONMENT` | `local`        | Runtime environment (`local`/`test`/`production`) |
| `LOG_LEVEL`   | `WARN`         | Minimum log level (`DEBUG`/`INFO`/`WARN`/`ERROR`) |
| `READ_TIMEOUT`  | `15`         | Max seconds to read a request                 |
| `WRITE_TIMEOUT` | `15`         | Max seconds to write a response               |
| `IDLE_TIMEOUT`  | `60`         | Max keep-alive idle seconds                   |

Custom config types only need to embed the env struct tags and implement `Validate() error`:

```go
type AppConfig struct {
    APIKey      string `env:"API_KEY"`
    MaxRequests int    `env:"MAX_REQUESTS" envDefault:"100"`
}

func (c AppConfig) Validate() error {
    if c.APIKey == "" {
        return fmt.Errorf("API_KEY is required")
    }
    return nil
}

cfg, err := config.ParseConfig[AppConfig]()
```

### logging

Configures the global `slog` default logger based on a `config.ServerConfig`. Call it once during application initialisation before spawning goroutines that log.

```go
cfg, err := config.ParseConfig[config.ServerConfig]()
if err != nil {
    log.Fatal(err)
}
logging.SetDefaultLogger(cfg)

// All slog.* calls now use the configured logger.
slog.Info("server starting", "port", cfg.Port)
```

Handler selection is automatic:

- **Local** environment — `slog.TextHandler` (human-readable output).
- **Test / Production** — `slog.JSONHandler` (structured output for log aggregation).

The log level is taken from `cfg.LogLevel`, which maps to the `LOG_LEVEL` environment variable.

### server

Creates and runs an HTTP server with environment-driven configuration and graceful shutdown.

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    if err := server.Run(context.Background(), mux); err != nil {
        log.Fatal(err)
    }
}
```

`Run` manages the full lifecycle:

1. Parses `ServerConfig` from environment variables (and configures the global logger as a side effect).
2. Starts `ListenAndServe` in a background goroutine.
3. Blocks until SIGINT (Ctrl+C) or context cancellation.
4. Performs graceful shutdown with a 10-second timeout.

For more control, use `NewServerWithConfig` to obtain a configured `*http.Server` and manage its lifecycle yourself.

## Typical startup sequence

```go
func main() {
    // 1. Parse config
    cfg, err := config.ParseConfig[config.ServerConfig]()
    if err != nil {
        log.Fatal(err)
    }

    // 2. Configure global logger
    logging.SetDefaultLogger(cfg)

    // 3. Build handler with middleware
    mux := http.NewServeMux()
    mux.HandleFunc("/api/", apiHandler)

    logger := slog.Default()
    stack := middleware.CreateStack(
        middleware.NewLoggingMiddleware(logger),
        middleware.NewMaxBytesReader(1024*1024),
        middleware.NewSetContentTypeJSON(),
    )

    // 4. Run (blocks until shutdown)
    if err := server.Run(context.Background(), stack(mux)); err != nil {
        log.Fatal(err)
    }
}
```

## Documentation

Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/harrydayexe/GoWebUtilities).

```bash
# View package documentation locally
go doc github.com/harrydayexe/GoWebUtilities/middleware
go doc github.com/harrydayexe/GoWebUtilities/config
go doc github.com/harrydayexe/GoWebUtilities/logging
go doc github.com/harrydayexe/GoWebUtilities/server
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run examples only
go test -v ./... -run Example
```

## Contributing

Contributions are welcome! Please ensure all tests pass and code follows Go conventions.
