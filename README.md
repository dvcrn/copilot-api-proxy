# Copilot API Proxy

A reverse proxy server written in Go that forwards `/v1/completion` requests to the GitHub Copilot API, to expose the Copilot API to other tools

> [!WARNING]
> This is a reverse-engineered proxy of GitHub Copilot API. It is not supported by GitHub, and may break unexpectedly. Use at your own risk.

> [!WARNING]
> **GitHub Security Notice:**
> Excessive automated or scripted use of Copilot (including rapid or bulk requests, such as via automated tools) may trigger GitHub's abuse-detection systems.
> You may receive a warning from GitHub Security, and further anomalous activity could result in temporary suspension of your Copilot access.
>
> GitHub prohibits use of their servers for excessive automated bulk activity or any activity that places undue burden on their infrastructure.
>
> Please review:
>
> - [GitHub Acceptable Use Policies](https://docs.github.com/site-policy/acceptable-use-policies/github-acceptable-use-policies#4-spam-and-inauthentic-activity-on-github)
> - [GitHub Copilot Terms](https://docs.github.com/site-policy/github-terms/github-terms-for-additional-products-and-features#github-copilot)
>
> Use this proxy responsibly to avoid account restrictions.

## Setup

```
go install github.com/dvcrn/copilot-api-proxy/cmd/
```

1. Set the `COPILOT_TOKEN` environment variable with your GitHub Copilot authentication token.
2. Optionally set `PROXY_PORT` (defaults to 8080).

## Usage

```bash
# Build and run
just build
just run

# Or run directly
go run ./cmd/copilot-api-proxy/main.go
```

## Configuration

- `COPILOT_TOKEN`: Required GitHub Copilot authentication token
- `PROXY_PORT`: Port to listen on (default: 8080)

## Architecture

The proxy consists of:
- `cmd/copilot-api-proxy/main.go`: Application entry point
- `internal/server/`: HTTP server and request handlers
- `pkg/copilot/`: Client for the upstream Copilot API
- `pkg/config/`: Configuration management
- `pkg/httpstreaming/`: HTTP response streaming utilities
