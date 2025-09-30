package server

import (
	"bytes"
	"encoding/json"
	"io"
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

		// Read the body to log the model
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			s.logger.Error("Failed to read request body", "error", err)
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		// Restore the body so it can be read again
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Log the model
		if r.URL.Path == "/v1/chat/completions" {
			var chatReq struct {
				Model string `json:"model"`
			}
			if err := json.Unmarshal(bodyBytes, &chatReq); err == nil {
				s.logger.Info("Request model", "model", chatReq.Model)
			}
		}

		// Forward the request to the Copilot client
		upstreamResp, err := s.copilotClient.ForwardRequest(r.Context(), r)
		if err != nil {
			s.logger.Error("Upstream request failed", "error", err)
			http.Error(w, "Failed to proxy request", http.StatusBadGateway)
			return
		}
		defer upstreamResp.Body.Close()

		if upstreamResp.StatusCode != http.StatusOK {
			bodyBytes, err := io.ReadAll(upstreamResp.Body)
			if err != nil {
				s.logger.Error("Failed to read upstream error response body", "error", err)
				http.Error(w, "Failed to read upstream error response", http.StatusBadGateway)
				return
			}
			s.logger.Error("Upstream request returned non-OK status",
				"status", upstreamResp.Status,
				"body", string(bodyBytes))

			// We still want to forward the response to the client
			w.WriteHeader(upstreamResp.StatusCode)
			w.Write(bodyBytes)
			return
		}

		// Stream the response back to the original client
		httpstreaming.StreamResponse(w, upstreamResp, s.logger)
	}
}
