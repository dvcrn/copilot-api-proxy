package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"copilot-api-proxy/internal/server"
	"copilot-api-proxy/pkg/config"
	"copilot-api-proxy/pkg/copilot"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := config.EnsurePaths(); err != nil {
		logger.Error("Failed to ensure paths", "error", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printUsage(logger)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "auth":
		runAuth(logger)
	case "server":
		runServer(logger)
	default:
		logger.Error("Unknown command", "command", command)
		printUsage(logger)
		os.Exit(1)
	}
}

func printUsage(logger *slog.Logger) {
	fmt.Println("Usage: go run cmd/copilot-api-proxy/main.go [command]")
	fmt.Println("Commands:")
	fmt.Println("  auth    - Exchange a GitHub token for a Copilot token and print it.")
	fmt.Println("  server  - Run the Copilot proxy server.")
}

func runAuth(logger *slog.Logger) {
	logger.Info("Starting GitHub device authentication flow.")
	deviceCode, err := copilot.GetDeviceCode(context.Background())
	if err != nil {
		logger.Error("Failed to get device code", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Please enter the code \"%s\" in %s\n", deviceCode.UserCode, deviceCode.VerificationURI)

	accessToken, err := copilot.PollAccessToken(context.Background(), deviceCode)
	if err != nil {
		logger.Error("Failed to get access token", "error", err)
		os.Exit(1)
	}

	logger.Info("Successfully authenticated with GitHub.")

	// Save the token
	tokenPath, err := config.GetGitHubTokenPath()
	if err != nil {
		logger.Error("Failed to get token path", "error", err)
		os.Exit(1)
	}
	if err := os.WriteFile(tokenPath, []byte(accessToken), 0o600); err != nil {
		logger.Error("Failed to write GitHub token to file", "error", err)
		os.Exit(1)
	}
	logger.Info("GitHub token saved", "path", tokenPath)

	// Now, exchange the GitHub token for a Copilot token
	tokenResponse, err := copilot.ExchangeGitHubToken(context.Background(), accessToken)
	if err != nil {
		logger.Error("Failed to exchange GitHub token for Copilot token", "error", err)
		os.Exit(1)
	}

	fmt.Print(tokenResponse.Token)
}

func runServer(logger *slog.Logger) {
	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create the token manager for handling Copilot token lifecycle
	tm, err := copilot.NewTokenManager(context.Background(), cfg.GitHubToken, logger)
	if err != nil {
		logger.Error("Failed to create token manager", "error", err)
		os.Exit(1)
	}
	defer tm.Close()

	// Create an instance of the Copilot API client

	copilotClient := copilot.NewClient(tm, 30*time.Second)

	// Create a new server instance
	srv := server.New(cfg.Port, logger, copilotClient)

	// Set up graceful shutdown

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start the server
	logger.Info("Starting Copilot Proxy server")
	if err := srv.Start(ctx); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
