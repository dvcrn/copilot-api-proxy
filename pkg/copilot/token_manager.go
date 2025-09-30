package copilot

import (
	"context"
	"fmt"
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
