package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	githubDeviceCodeURL  = "https://github.com/login/device/code"
	githubAccessTokenURL = "https://github.com/login/oauth/access_token"
	githubClientID       = "Iv1.b507a08c87ecfe98"
)

// DeviceCodeResponse holds the response from the GitHub device code endpoint.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// AccessTokenResponse holds the response from the GitHub access token endpoint.
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

// GetDeviceCode retrieves a device and user code from GitHub.
func GetDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	body, err := json.Marshal(map[string]string{
		"client_id": githubClientID,
		"scope":     "read:user",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", githubDeviceCodeURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create device code request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute device code request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	var deviceCodeResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceCodeResp); err != nil {
		return nil, fmt.Errorf("failed to decode device code response: %w", err)
	}

	return &deviceCodeResp, nil
}

// PollAccessToken polls GitHub for an access token using the device code.
func PollAccessToken(ctx context.Context, deviceCode *DeviceCodeResponse) (string, error) {
	interval := time.Duration(deviceCode.Interval) * time.Second

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
			body, err := json.Marshal(map[string]string{
				"client_id":   githubClientID,
				"device_code": deviceCode.DeviceCode,
				"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
			})
			if err != nil {
				return "", fmt.Errorf("failed to marshal request body: %w", err)
			}

			req, err := http.NewRequestWithContext(ctx, "POST", githubAccessTokenURL, bytes.NewBuffer(body))
			if err != nil {
				return "", fmt.Errorf("failed to create access token request: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				// Don't return, just log and continue polling
				fmt.Printf("failed to execute access token request: %v\n", err)
				continue
			}
			defer resp.Body.Close()

			var accessTokenResp AccessTokenResponse
			if err := json.NewDecoder(resp.Body).Decode(&accessTokenResp); err != nil {
				fmt.Printf("failed to decode access token response: %v\n", err)
				continue
			}

			if accessTokenResp.Error != "" {
				// These are expected errors while the user hasn't authorized.
				// e.g. "authorization_pending", "slow_down", "expired_token"
				if accessTokenResp.Error == "authorization_pending" {
					// continue polling
				} else if accessTokenResp.Error == "expired_token" {
					return "", fmt.Errorf("device code expired")
				}
				// "slow_down" means we should increase the interval, but for now we just continue
				continue
			}

			if accessTokenResp.AccessToken != "" {
				return accessTokenResp.AccessToken, nil
			}
		}
	}
}
