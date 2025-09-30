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
	token := c.tokenManager.GetToken()
	upstreamReq.Header.Set("Authorization", "Bearer "+token)

	// 5. Execute the request and return the response.
	// Do not close the response body here; the caller needs to stream it.
	return c.httpClient.Do(upstreamReq)
}
