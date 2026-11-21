package api

import (
	"context"
	"net/http"
	"time"
)

// readyTimeout bounds the database check of a readiness probe, below the
// one-second default probe timeout of Kubernetes.
const readyTimeout = 800 * time.Millisecond

// handleHealthz answers a liveness probe: the process serves HTTP. It
// checks nothing else, so a database outage does not get the process
// restarted.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

// handleReadyz answers a readiness probe: the database answers, so the
// dashboard and API can serve.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
	defer cancel()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if err := s.storage.Ping(ctx); err != nil {
		s.log.Warn().Err(err).Msg("readiness check failed")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("database unavailable\n"))
		return
	}
	_, _ = w.Write([]byte("ok\n"))
}
