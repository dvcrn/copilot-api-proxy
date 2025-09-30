package httpstreaming

import (
	"io"
	"log/slog"
	"net/http"
)

// StreamResponse copies headers and streams the body from an upstream response
// to the client's response writer, flushing chunks as they arrive.
func StreamResponse(w http.ResponseWriter, upstreamResp *http.Response, logger *slog.Logger) {
	// Copy headers from the upstream response to our response writer.
	for key, values := range upstreamResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(upstreamResp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Warn("Response writer does not support flushing. Streaming may not be real-time.")
		io.Copy(w, upstreamResp.Body)
		return
	}

	// Stream the body, flushing after each write.
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := upstreamResp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				logger.Error("Failed to write chunk to client", "error", writeErr)
				break
			}
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("Error reading from upstream body", "error", err)
			break
		}
	}
}