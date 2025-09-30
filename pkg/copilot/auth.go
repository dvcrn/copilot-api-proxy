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
