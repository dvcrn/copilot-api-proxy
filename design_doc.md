# **Technical Design Document: Copilot Proxy**

## 1. Overview

The goal of this project is to create a reverse proxy server written in Go. This server will accept incoming HTTP requests, inject a GitHub Copilot authentication token into the `Authorization` header, and forward the requests to the official GitHub Copilot API (`api.individual.githubcopilot.com`).

The server must be capable of handling streaming responses (`text/event-stream`) to ensure real-time communication between the client and the Copilot API. The architecture will be based on standard, idiomatic Go practices, emphasizing modularity and testability.

## 2. Architecture

The project will follow a standard Go project layout to separate concerns.

#### **Directory Structure**

```
/copilot-proxy/
├─── cmd/
│   └─── copilot-proxy/
│       └─── main.go           # Application entry point
├─── internal/
│   └─── server/
│       ├─── server.go         # HTTP server setup and lifecycle
│       └─── handlers.go       # HTTP request handler logic
├─── pkg/
│   ├─── copilot/
│   │   └─── client.go         # Client for the upstream Copilot API
│   ├─── config/
│   │   └─── config.go         # Environment variable configuration
│   └─── httpstreaming/
│       └─── streamer.go       # Utility for streaming HTTP responses
├─── go.mod
├─── go.sum
├─── justfile
└─── README.md
```

## 3. Component Deep Dive

### `cmd/copilot-proxy/main.go`

**Responsibility:** Initializes and starts the application. It wires all the components together.

**Implementation Details:**
The `main` function will perform the following steps:
1.  Initialize a structured logger (e.g., `slog.New`).
2.  Load configuration from environment variables using the `config` package. Exit if critical configuration (like the auth token) is missing.
3.  Create an instance of the Copilot API client (`copilot.NewClient`), passing it the auth token.
4.  Create a new server instance (`server.New`), injecting the logger and the Copilot client.
5.  Set up a mechanism for graceful shutdown using `context` and `os.Signal` to listen for `SIGINT` and `SIGTERM`.
6.  Start the server by calling its `Start()` method and log any fatal errors.

### `pkg/config/config.go`

**Responsibility:** Manages application configuration.

**Implementation Details:**

```go
package config

import (
	"errors"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	Port             string
	CopilotAuthToken string
}

// Load populates the Config struct from environment variables.
func Load() (*Config, error) {
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	token := os.Getenv("COPILOT_TOKEN")
	if token == "" {
		return nil, errors.New("COPILOT_TOKEN environment variable not set")
	}

	return &Config{
		Port:             port,
		CopilotAuthToken: token,
	}, nil
}
```

### `pkg/copilot/client.go`

**Responsibility:** Encapsulates all logic for communicating with the upstream GitHub Copilot API.

**Implementation Details:**

```go
package copilot

import (
	"context"
	"net/http"
	"time"
)

const (
	copilotAPIHost = "api.individual.githubcopilot.com"
)

// Client is an HTTP client for forwarding requests to the Copilot API.
type Client struct {
	httpClient *http.Client
	authToken  string
}

// NewClient creates a new Copilot client.
func NewClient(authToken string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		authToken:  authToken,
	}
}

// ForwardRequest creates and sends a new request to the Copilot API based on
// an incoming request, adding the necessary authentication.
// The caller is responsible for closing the response body.
func (c *Client) ForwardRequest(ctx context.Context, incomingReq *http.Request) (*http.Response, error) {
	// 1. Construct the target URL.
	targetURL := "https://" + copilotAPIHost + incomingReq.URL.Path
	
	// 2. Create a new request to the upstream API.
	// The body of the incoming request is passed directly.
	upstreamReq, err := http.NewRequestWithContext(ctx, incomingReq.Method, targetURL, incomingReq.Body)
	if err != nil {
		return nil, err
	}

	// 3. Copy headers from the original request.
	upstreamReq.Header = incomingReq.Header.Clone()

	// 4. Set the required headers for the Copilot API.
	upstreamReq.Host = copilotAPIHost
	upstreamReq.Header.Set("Authorization", "Bearer "+c.authToken)

	// 5. Execute the request and return the response.
	// Do not close the response body here; the caller needs to stream it.
	return c.httpClient.Do(upstreamReq)
}
```

### `internal/server/server.go`

**Responsibility:** Defines the HTTP server and manages its lifecycle, including routing and graceful shutdown.

**Implementation Details:**

```go
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"copilot-proxy/pkg/copilot"
)

// Server is the main HTTP server for the proxy.
type Server struct {
	addr          string
	logger        *slog.Logger
	copilotClient *copilot.Client
}

// New creates a new server instance.
func New(port string, logger *slog.Logger, client *copilot.Client) *Server {
	return &Server{
		addr:          ":" + port,
		logger:        logger,
		copilotClient: client,
	}
}

// Start runs the HTTP server and blocks until the context is canceled.
func (s *Server) Start(ctx context.Context) error {
	router := http.NewServeMux()
	s.registerRoutes(router)

	httpServer := &http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	// Goroutine for graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(shutdownCtx)
	}()

	s.logger.Info("Server starting", "address", s.addr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
```

### `internal/server/handlers.go`

**Responsibility:** Contains the HTTP handler functions that perform the proxying logic.

**Implementation Details:**

```go
package server

import (
	"net/http"

	"copilot-proxy/pkg/httpstreaming"
)

// registerRoutes sets up the routing for the server.
func (s *Server) registerRoutes(router *http.ServeMux) {
	router.HandleFunc("/", s.proxyHandler())
}

// proxyHandler is the main handler for all incoming requests.
func (s *Server) proxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("Incoming request", "method", r.Method, "path", r.URL.Path)

		// Forward the request to the Copilot client
		upstreamResp, err := s.copilotClient.ForwardRequest(r.Context(), r)
		if err != nil {
			s.logger.Error("Upstream request failed", "error", err)
			http.Error(w, "Failed to proxy request", http.StatusBadGateway)
			return
		}
		defer upstreamResp.Body.Close()

		// Stream the response back to the original client
		httpstreaming.StreamResponse(w, upstreamResp, s.logger)
	}
}
```

### `pkg/httpstreaming/streamer.go`

**Responsibility:** Provides a utility to stream an HTTP response body back to a client, ensuring headers are copied and the body is flushed chunk-by-chunk.

**Implementation Details:**

```go
package httpstreaming

import (
	"io"
	"log/slog"
	"net/http"
)

// StreamResponse copies headers and streams the body from an upstream response
// to the client's response writer, flushing chunks as they arrive.
func StreamResponse(w http.ResponseWriter, upstreamResp *http.Response, logger *slog.Logger) {
	// Copy headers from the upstream response to our response writer.
	for key, values := range upstreamResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(upstreamResp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Warn("Response writer does not support flushing. Streaming may not be real-time.")
		io.Copy(w, upstreamResp.Body)
		return
	}

	// Stream the body, flushing after each write.
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := upstreamResp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				logger.Error("Failed to write chunk to client", "error", writeErr)
				break
			}
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("Error reading from upstream body", "error", err)
			break
		}
	}
}
```

## 4. Task Runner (`justfile`)

A `justfile` will be provided to standardize common development tasks.

```makefile
# justfile

# Build the application binary
build:
    go build -o ./bin/copilot-proxy ./cmd/copilot-proxy

# Run the application directly
run:
    go run ./cmd/copilot-proxy/main.go

# Format all Go source files
fmt:
    go fmt ./...

# Run all tests
test:
    go test -v ./...
```

## 5. Future Considerations

*   **Dynamic Token Authentication:** The current design uses a static token. A future iteration should replace this with a proper authentication mechanism, such as an OAuth2 flow, to dynamically fetch and refresh the Copilot token. This logic would be encapsulated within `pkg/copilot/client.go`.
*   **Rate Limiting & Caching:** To protect the service and the upstream API, middleware for rate limiting could be added in `internal/server/handlers.go`. Caching strategies could also be implemented for non-streaming, idempotent requests.
