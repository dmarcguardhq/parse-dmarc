package api

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"

	"github.com/dmarcguardhq/parse-dmarc/internal/metrics"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
)

//go:embed dist
var distFS embed.FS

// VersionInfo is the build the server runs, as main's ldflags set it.
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// Server represents the API server
type Server struct {
	storage *storage.Storage
	metrics *metrics.Metrics
	log     *zerolog.Logger
	addr    string
	version VersionInfo
	cors    corsPolicy
}

// Option configures a Server.
type Option func(*Server)

// WithCORSAllowedOrigins lets the listed origins (comma-separated, or "*")
// read the API from another origin, as a dashboard pointed at a remote API
// endpoint does. Without it the API sends no CORS headers.
func WithCORSAllowedOrigins(list string) Option {
	return func(s *Server) { s.cors = parseCORSOrigins(list) }
}

// NewServer creates a new API server
func NewServer(store *storage.Storage, host string, port int, m *metrics.Metrics, log *zerolog.Logger, version VersionInfo, opts ...Option) *Server {
	s := &Server{
		storage: store,
		metrics: m,
		log:     log,
		addr:    fmt.Sprintf("%s:%d", host, port),
		version: version,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/reports", s.handleReports)
	mux.HandleFunc("/api/reports/", s.handleReportDetail)
	mux.HandleFunc("/api/statistics", s.handleStatistics)
	mux.HandleFunc("/api/top-sources", s.handleTopSources)
	mux.HandleFunc("/api/version", s.handleVersion)

	// Prometheus metrics endpoint
	if s.metrics != nil {
		mux.Handle("/metrics", s.metrics.Handler())
	}

	// Serve frontend
	// Try to serve embedded files, fallback to nothing if not embedded
	distFiles, err := fs.Sub(distFS, "dist")
	if err == nil {
		mux.Handle("/", http.FileServer(http.FS(distFiles)))
	} else {
		// If dist folder is not embedded, serve a simple message
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html")
				_, _ = fmt.Fprintf(w, `
					<!DOCTYPE html>
					<html>
					<head><title>DMARC Dashboard</title></head>
					<body>
						<h1>DMARC Report Dashboard API</h1>
						<p>API is running. Frontend assets not embedded yet.</p>
						<ul>
							<li><a href="/api/statistics">Statistics</a></li>
							<li><a href="/api/reports">Reports</a></li>
							<li><a href="/api/top-sources">Top Sources</a></li>
							<li><a href="/metrics">Prometheus Metrics</a></li>
						</ul>
					</body>
					</html>
				`)
			} else {
				http.NotFound(w, r)
			}
		})
	}

	// Build handler chain: CORS -> Metrics -> Routes
	var handler http.Handler = mux
	if s.metrics != nil {
		handler = s.metrics.HTTPMiddleware(handler)
	}
	handler = s.corsMiddleware(handler)

	server := newHTTPServer(s.addr, handler)

	go func() {
		<-ctx.Done()
		s.log.Info().Msg("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			s.log.Error().Err(err).Msg("server shutdown error")
		}
	}()

	s.log.Info().Str("addr", s.addr).Msg("starting server")
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server listen on %s: %w", s.addr, err)
	}
	return nil
}

// Timeouts of the dashboard HTTP server. They mirror the ones the MCP server
// sets in internal/mcp/server.go, so both operator-facing surfaces bound idle
// and slow clients the same way.
const (
	readTimeout    = 15 * time.Second
	writeTimeout   = 15 * time.Second
	idleTimeout    = 60 * time.Second
	maxHeaderBytes = 1 << 20 // 1 MiB
)

// newHTTPServer builds the dashboard server with the same slow-client bounds
// the MCP server applies.
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

// corsMiddleware answers cross-origin requests only from the configured
// origins. The API has no authentication, so a wildcard would let any web
// page a user opens read the reports of a dashboard their browser can
// reach, loopback and LAN included.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" || !s.cors.allows(origin) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Origin")
		if s.cors.any {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// corsPolicy is the set of origins allowed to read the API cross-origin.
type corsPolicy struct {
	any     bool
	origins map[string]bool
}

// parseCORSOrigins reads a comma-separated origin list; "*" allows any.
func parseCORSOrigins(list string) corsPolicy {
	p := corsPolicy{origins: map[string]bool{}}
	for _, o := range strings.Split(list, ",") {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		switch o {
		case "":
		case "*":
			p.any = true
		default:
			p.origins[strings.ToLower(o)] = true
		}
	}
	return p
}

func (p corsPolicy) allows(origin string) bool {
	return p.any || p.origins[strings.ToLower(origin)]
}

// handleReports returns a list of reports
func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > storage.MaxReportLimit {
		limit = storage.MaxReportLimit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	reports, err := s.storage.GetReports(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, reports)
}

// handleReportDetail returns a single report detail
func (s *Server) handleReportDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL
	idStr := r.URL.Path[len("/api/reports/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	report, err := s.storage.GetReportByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	s.writeJSON(w, report)
}

// handleStatistics returns dashboard statistics
func (s *Server) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := s.storage.GetStatistics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, stats)
}

// handleTopSources returns top source IPs
func (s *Server) handleTopSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	sources, err := s.storage.GetTopSourceIPs(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, sources)
}

// handleVersion returns the build the server runs
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.writeJSON(w, s.version)
}

// writeJSON writes JSON response
func (s *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Error().Err(err).Msg("failed to encode JSON")
	}
}

// RefreshMetrics updates all Prometheus metrics from current database state
func (s *Server) RefreshMetrics() {
	if s.metrics == nil {
		return
	}

	// Update basic statistics
	stats, err := s.storage.GetStatistics()
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get statistics for metrics")
	} else {
		s.metrics.UpdateStatistics(
			stats.TotalReports,
			stats.TotalMessages,
			stats.CompliantMessages,
			stats.UniqueSourceIPs,
			stats.UniqueDomains,
			stats.ComplianceRate,
		)
	}

	// Update per-domain metrics
	domainStats, err := s.storage.GetDomainStats()
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get domain stats for metrics")
	} else {
		for _, ds := range domainStats {
			s.metrics.UpdateDomainMetrics(ds.Domain, ds.TotalMessages, ds.ComplianceRate)
		}
	}

	// Update per-organization metrics
	orgStats, err := s.storage.GetOrgStats()
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get org stats for metrics")
	} else {
		for _, os := range orgStats {
			s.metrics.UpdateOrgMetrics(os.OrgName, os.Reports)
		}
	}

	// Update disposition metrics
	dispStats, err := s.storage.GetDispositionStats()
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get disposition stats for metrics")
	} else {
		for _, ds := range dispStats {
			s.metrics.UpdateDispositionMetrics(ds.Disposition, ds.Count)
		}
	}

	// Update authentication results
	spfStats, errSpf := s.storage.GetSPFStats()
	dkimStats, errDkim := s.storage.GetDKIMStats()
	if errSpf != nil {
		s.log.Error().Err(errSpf).Msg("failed to get SPF stats for metrics")
	}
	if errDkim != nil {
		s.log.Error().Err(errDkim).Msg("failed to get DKIM stats for metrics")
	}
	if errSpf == nil && errDkim == nil {
		spfResults := make(map[string]int)
		for _, s := range spfStats {
			spfResults[s.Result] = s.Count
		}
		dkimResults := make(map[string]int)
		for _, d := range dkimStats {
			dkimResults[d.Result] = d.Count
		}
		s.metrics.UpdateAuthResults(spfResults, dkimResults)
	}
}

// GetMetrics returns the metrics instance
func (s *Server) GetMetrics() *metrics.Metrics {
	return s.metrics
}
