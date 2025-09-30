package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	// Create an instance of the Copilot API client
	copilotClient := copilot.NewClient(cfg.CopilotAuthToken)

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