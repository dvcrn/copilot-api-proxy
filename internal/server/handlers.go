package server

import (
	"net/http"

	"copilot-proxy/pkg/httpstreaming"
)

// registerRoutes sets up the routing for the server.
func (s *Server) registerRoutes(router *http.ServeMux) {
	router.HandleFunc("/", s.proxyHandler())
}

// proxyHandler is the main handler for all incoming requests.
func (s *Server) proxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info("Incoming request", "method", r.Method, "path", r.URL.Path)

		// Forward the request to the Copilot client
		upstreamResp, err := s.copilotClient.ForwardRequest(r.Context(), r)
		if err != nil {
			s.logger.Error("Upstream request failed", "error", err)
			http.Error(w, "Failed to proxy request", http.StatusBadGateway)
			return
		}
		defer upstreamResp.Body.Close()

		// Stream the response back to the original client
		httpstreaming.StreamResponse(w, upstreamResp, s.logger)
	}
}
