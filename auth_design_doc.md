# Copilot Proxy - Authentication Design

This document provides a detailed implementation plan for the authentication components of the Copilot Proxy. It refines the initial `design_doc.md` by specifying the architecture required to handle the full, dynamic lifecycle of a Copilot API token.

## 1. Overview

The authentication system will be responsible for:
1.  Using a long-lived GitHub OAuth token.
2.  Exchanging it for a short-lived Copilot API token.
3.  Managing the Copilot token in memory.
4.  Automatically refreshing the token before it expires.
5.  Providing a valid token to the part of the proxy that forwards requests.

To achieve this, the `pkg/copilot/` directory will be structured to separate the stateless API calls from the stateful token management.

## 2. Proposed File Structure

The `pkg/copilot/` directory will be organized as follows:

```
/pkg/copilot/
├─── auth.go            # Stateless functions for the token exchange API call.
├─── token_manager.go   # Stateful, thread-safe token lifecycle management.
└─── client.go          # (Updated) The proxy client, which uses the TokenManager.
```

## 3. Component Deep Dive

### `pkg/copilot/auth.go`

**Responsibility:** Contains the low-level, stateless function for performing the GitHub-to-Copilot token exchange.

**Implementation Details:**

```go
package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ExchangeTokenResponse defines the structure of the JSON response from the token exchange endpoint.
type ExchangeTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	RefreshIn int64  `json:"refresh_in"`
}

// ExchangeGitHubToken takes a GitHub OAuth token and exchanges it for a short-lived Copilot token.
func ExchangeGitHubToken(ctx context.Context, githubToken string) (*ExchangeTokenResponse, error) {
	// 1. Create a new GET request to the exchange endpoint.
	url := "https://api.github.com/copilot_internal/v2/token"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange request: %w", err)
	}

	// 2. Add the required headers.
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/json")

	// 3. Execute the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute token exchange request: %w", err)
	}
	defer resp.Body.Close()

	// 4. Handle non-successful status codes.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange request failed with status: %s", resp.Status)
	}

	// 5. Unmarshal the JSON response.
	var tokenResponse ExchangeTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode token exchange response: %w", err)
	}

	// 6. Return the response.
	return &tokenResponse, nil
}
```

### `pkg/copilot/token_manager.go`

**Responsibility:** A stateful, thread-safe manager that holds the Copilot token and runs a background process to refresh it automatically.

**Implementation Details:**

```go
package copilot

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// TokenManager handles the Copilot token and its refresh cycle.
type TokenManager struct {
	mu           sync.RWMutex
	githubToken  string
	copilotToken string
	refreshesAt  time.Time
	logger       *slog.Logger
	stopCh       chan struct{}
}

// NewTokenManager creates a manager, gets the initial token, and starts the refresh loop.
func NewTokenManager(ctx context.Context, githubToken string, logger *slog.Logger) (*TokenManager, error) {
	tm := &TokenManager{
		githubToken: githubToken,
		logger:      logger,
		stopCh:      make(chan struct{}),
	}

	if err := tm.refresh(ctx); err != nil {
		return nil, fmt.Errorf("initial token refresh failed: %w", err)
	}

	go tm.refreshTokenLoop()
	return tm, nil
}

// GetToken returns the current, valid Copilot token in a thread-safe way.
func (tm *TokenManager) GetToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.copilotToken
}

// Close gracefully stops the background refresh loop.
func (tm *TokenManager) Close() {
	close(tm.stopCh)
}

// refreshTokenLoop runs in the background, refreshing the token before it expires.
func (tm *TokenManager) refreshTokenLoop() {
	for {
		tm.mu.RLock()
		duration := time.Until(tm.refreshesAt)
		tm.mu.RUnlock()

		select {
		case <-time.After(duration):
			tm.logger.Info("Refreshing Copilot token")
			if err := tm.refresh(context.Background()); err != nil {
				tm.logger.Error("Failed to refresh token", "error", err)
			}
		case <-tm.stopCh:
			tm.logger.Info("Token refresh loop stopped")
			return
		}
	}
}

// refresh executes the token exchange and updates the manager's state.
func (tm *TokenManager) refresh(ctx context.Context) error {
	resp, err := ExchangeGitHubToken(ctx, tm.githubToken)
	if err != nil {
		return err
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.copilotToken = resp.Token
	// Refresh 60 seconds before the official refresh_in time as a buffer.
	refreshDuration := time.Duration(resp.RefreshIn-60) * time.Second
	tm.refreshesAt = time.Now().Add(refreshDuration)

	tm.logger.Info("Successfully refreshed Copilot token", "expires_at", resp.ExpiresAt)
	return nil
}
```

### `pkg/copilot/client.go` (Updated)

**Responsibility:** The proxy client, updated to use the `TokenManager`.

**Implementation Details:**

```go
package copilot

import (
	"context"
	"net/http"
	"time"
)

// Client now holds a reference to the TokenManager.
type Client struct {
	httpClient   *http.Client
	tokenManager *TokenManager
}

// NewClient is updated to accept the TokenManager.
func NewClient(tokenManager *TokenManager, timeout time.Duration) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: timeout},
		tokenManager: tokenManager,
	}
}

// ForwardRequest gets the latest token from the manager before each request.
func (c *Client) ForwardRequest(ctx context.Context, incomingReq *http.Request) (*http.Response, error) {
	// ... (request creation logic remains the same)

	// Get the latest valid token for this specific request.
	token := c.tokenManager.GetToken()
	upstreamReq.Header.Set("Authorization", "Bearer "+token)

	// ... (request execution logic remains the same)
}
```
