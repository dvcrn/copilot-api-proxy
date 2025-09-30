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
	httpClient   *http.Client
	tokenManager *TokenManager
}

// NewClient creates a new Copilot client.
func NewClient(tokenManager *TokenManager, timeout time.Duration) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: timeout},
		tokenManager: tokenManager,
	}
}

// ForwardRequest creates and sends a new request to the Copilot API based on
// an incoming request, adding the necessary authentication.
// The caller is responsible for closing the response body.
func (c *Client) ForwardRequest(ctx context.Context, incomingReq *http.Request) (*http.Response, error) {
	// 1. Construct the target URL.
	path := incomingReq.URL.Path
	if path == "/v1/chat/completions" {
		path = "/chat/completions"
	}
	targetURL := "https://" + copilotAPIHost + path

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
	token := c.tokenManager.GetToken()
	upstreamReq.Header.Set("Authorization", "Bearer "+token)
	upstreamReq.Header.Set("editor-version", "vscode/1.98.1")
	upstreamReq.Header.Set("editor-plugin-version", "copilot-chat/0.26.7")
	upstreamReq.Header.Set("user-agent", "GitHubCopilotChat/0.26.7")
	upstreamReq.Header.Set("x-github-api-version", "2025-04-01")
	upstreamReq.Header.Set("copilot-integration-id", "vscode-chat")
	upstreamReq.Header.Set("openai-intent", "conversation-panel")
	upstreamReq.Header.Set("x-vscode-user-agent-library-version", "electron-fetch")

	// 5. Execute the request and return the response.
	// Do not close the response body here; the caller needs to stream it.
	return c.httpClient.Do(upstreamReq)
}
