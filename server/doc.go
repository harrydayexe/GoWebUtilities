// Package server provides utilities for creating and running HTTP servers
// with environment-based configuration and graceful shutdown handling.
//
// The server package integrates with the config package to automatically
// load server settings (port, timeouts) from environment variables. It
// provides graceful shutdown on interrupt signals with proper resource cleanup.
//
// Basic usage:
//
//	func main() {
//	    mux := http.NewServeMux()
//	    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	        w.Write([]byte("Hello, World!"))
//	    })
//
//	    if err := server.Run(context.Background(), mux); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// The Run function handles all server lifecycle management including:
//   - Loading configuration from environment variables
//   - Starting the HTTP server in a goroutine
//   - Listening for interrupt signals (SIGINT, SIGTERM)
//   - Performing graceful shutdown with a 10-second timeout
//
// For more control over the server instance, use NewServerWithConfig to
// create an *http.Server and manage its lifecycle manually.
//
// All functions are safe for concurrent use.
package server
