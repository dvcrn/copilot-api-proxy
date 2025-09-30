package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"copilot-proxy/internal/server"
	"copilot-proxy/pkg/config"
	"copilot-proxy/pkg/copilot"
)

func main() {
	// Initialize a structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create the token manager for handling Copilot token lifecycle
	tm, err := copilot.NewTokenManager(context.Background(), cfg.CopilotAuthToken, logger)
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
