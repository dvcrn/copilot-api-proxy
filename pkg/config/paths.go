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
	return filepath.Join(home, ".local", "share", "copilot-oauth-proxy", "github_token"), nil
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

	// Migrate the token from the previous storage location (copilot-api-proxy)
	// if it exists and the new path does not yet have one.
	if home, err := os.UserHomeDir(); err == nil {
		oldPath := filepath.Join(home, ".local", "share", "copilot-api-proxy", "github_token")
		if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
			if data, err := os.ReadFile(oldPath); err == nil {
				if err := os.WriteFile(tokenPath, data, 0o600); err == nil {
					_ = os.Remove(oldPath)
				}
			}
		}
	}

	// Ensure the file exists with 0600 permissions
	file, err := os.OpenFile(tokenPath, os.O_RDONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	return file.Close()
}
