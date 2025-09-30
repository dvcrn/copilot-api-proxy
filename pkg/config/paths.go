package config

import (
	"os"
	"path/filepath"
)

func GetGitHubTokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "copilot-api-proxy", "github_token"), nil
}

func EnsurePaths() error {
	tokenPath, err := GetGitHubTokenPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Ensure the file exists with 0600 permissions
	file, err := os.OpenFile(tokenPath, os.O_RDONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	return file.Close()
}
