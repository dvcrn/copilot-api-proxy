package config

import (
	"errors"
	"fmt"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	Port        string
	GitHubToken string
}

// Load populates the Config struct from environment variables.
func Load() (*Config, error) {
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// If env var is not set, try to read from file
		tokenPath, err := GetGitHubTokenPath()
		if err != nil {
			return nil, fmt.Errorf("failed to get token path: %w", err)
		}
		tokenBytes, err := os.ReadFile(tokenPath)
		if err != nil {
			return nil, errors.New("GITHUB_TOKEN environment variable not set and failed to read token from file")
		}
		token = string(tokenBytes)
	}

	if token == "" {
		return nil, errors.New("GITHUB_TOKEN is empty")
	}

	return &Config{
		Port:        port,
		GitHubToken: token,
	}, nil
}
