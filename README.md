# GoWebUtilities

A Go library providing reusable utilities for building web applications.

[![Go Reference](https://pkg.go.dev/badge/github.com/harrydayexe/GoWebUtilities.svg)](https://pkg.go.dev/github.com/harrydayexe/GoWebUtilities)
[![Go Report Card](https://goreportcard.com/badge/github.com/harrydayexe/GoWebUtilities)](https://goreportcard.com/report/github.com/harrydayexe/GoWebUtilities)
[![Test](https://github.com/harrydayexe/GoWebUtilities/actions/workflows/test.yml/badge.svg?event=push)](https://github.com/harrydayexe/GoWebUtilities/actions/workflows/test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


## Installation

```bash
go get github.com/harrydayexe/GoWebUtilities
```

## Available Modules

### Middleware

Composable HTTP middleware following Go's standard `net/http` pattern. All middleware follow the signature:

```go
type Middleware func(h http.Handler) http.Handler
```

Available middleware:
- **Logging** - Request logging with structured logging (`log/slog`)
- **MaxBytesReader** - Request body size limiting
- **SetContentType** - Response Content-Type header setting

**Example:**

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

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status":"ok"}`))
    })

    stack := middleware.CreateStack(
        middleware.NewLoggingMiddleware(logger),
        middleware.NewMaxBytesReader(1024*1024), // 1MB limit
        middleware.NewSetContentTypeJSON(),
    )

    http.ListenAndServe(":8080", stack(handler))
}
```

## Documentation

For detailed documentation, use `go doc`:

```bash
# View package documentation
go doc github.com/harrydayexe/GoWebUtilities/middleware

# View specific function documentation
go doc github.com/harrydayexe/GoWebUtilities/middleware.NewLoggingMiddleware
```

Full documentation is also available on [pkg.go.dev](https://pkg.go.dev/github.com/harrydayexe/GoWebUtilities).

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

Contributions are welcome! Please ensure all tests pass and code follows Go conventions.
