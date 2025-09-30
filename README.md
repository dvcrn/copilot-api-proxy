# Copilot Proxy

A reverse proxy server written in Go that forwards requests to the GitHub Copilot API with authentication.

## Setup

1. Set the `COPILOT_TOKEN` environment variable with your GitHub Copilot authentication token.
2. Optionally set `PROXY_PORT` (defaults to 8080).

## Usage

```bash
# Build and run
just build
just run

# Or run directly
go run ./cmd/copilot-proxy/main.go
```

## Configuration

- `COPILOT_TOKEN`: Required GitHub Copilot authentication token
- `PROXY_PORT`: Port to listen on (default: 8080)

## Architecture

The proxy consists of:
- `cmd/copilot-proxy/main.go`: Application entry point
- `internal/server/`: HTTP server and request handlers
- `pkg/copilot/`: Client for the upstream Copilot API
- `pkg/config/`: Configuration management
- `pkg/httpstreaming/`: HTTP response streaming utilities