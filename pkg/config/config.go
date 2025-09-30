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