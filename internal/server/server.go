package server

import (
	"context"
	"copilot-api-proxy/pkg/copilot"
	"log/slog"
	"net/http"
	"time"
)

// Server is the main HTTP server for the proxy.
type Server struct {
	addr          string
	logger        *slog.Logger
	copilotClient *copilot.Client
}

// New creates a new server instance.
func New(port string, logger *slog.Logger, client *copilot.Client) *Server {
	return &Server{
		addr:          ":" + port,
		logger:        logger,
		copilotClient: client,
	}
}

// Start runs the HTTP server and blocks until the context is canceled.
func (s *Server) Start(ctx context.Context) error {
	router := http.NewServeMux()
	s.registerRoutes(router)

	httpServer := &http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	// Goroutine for graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(shutdownCtx)
	}()

	s.logger.Info("Server starting", "address", s.addr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
