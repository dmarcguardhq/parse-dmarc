package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "parse_dmarc"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	registry *prometheus.Registry

	// Build info
	BuildInfo *prometheus.GaugeVec

	// Report processing metrics
	ReportsFetched     prometheus.Counter
	ReportsParsed      prometheus.Counter
	ReportsStored      prometheus.Counter
	ReportParseErrors  prometheus.Counter
	ReportStoreErrors  prometheus.Counter
	AttachmentsTotal   prometheus.Counter
	FetchDuration      prometheus.Histogram
	LastFetchTimestamp prometheus.Gauge
	FetchCyclesTotal   prometheus.Counter
	FetchErrors        prometheus.Counter

	// IMAP connection metrics
	IMAPConnectionsTotal   *prometheus.CounterVec
	IMAPConnectionDuration prometheus.Histogram

	// DMARC statistics (gauges that reflect current state)
	TotalReports      prometheus.Gauge
	TotalMessages     prometheus.Gauge
	CompliantMessages prometheus.Gauge
	ComplianceRate    prometheus.Gauge
	UniqueSourceIPs   prometheus.Gauge
	UniqueDomains     prometheus.Gauge

	// Per-domain metrics
	MessagesByDomain      *prometheus.GaugeVec
	ComplianceByDomain    *prometheus.GaugeVec
	ReportsByOrg          *prometheus.GaugeVec
	MessagesByDisposition *prometheus.GaugeVec

	// Authentication results
	SPFResults  *prometheus.GaugeVec
	DKIMResults *prometheus.GaugeVec

	// HTTP server metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge
}

// New creates and registers all Prometheus metrics
func New(version, commit, buildDate string) *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		registry: registry,

		BuildInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "build_info",
				Help:      "Build information for parse-dmarc",
			},
			[]string{"version", "commit", "build_date"},
		),

		// Report processing
		ReportsFetched: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "fetched_total",
				Help:      "Total number of DMARC report emails fetched from IMAP",
			},
		),
		ReportsParsed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "parsed_total",
				Help:      "Total number of DMARC reports successfully parsed",
			},
		),
		ReportsStored: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "stored_total",
				Help:      "Total number of DMARC reports successfully stored in database",
			},
		),
		ReportParseErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "parse_errors_total",
				Help:      "Total number of DMARC report parse errors",
			},
		),
		ReportStoreErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "store_errors_total",
				Help:      "Total number of DMARC report storage errors",
			},
		),
		AttachmentsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "attachments_total",
				Help:      "Total number of attachments processed",
			},
		),
		FetchDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "fetch_duration_seconds",
				Help:      "Duration of report fetch operations",
				Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to ~51s
			},
		),
		LastFetchTimestamp: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "last_fetch_timestamp_seconds",
				Help:      "Unix timestamp of the last successful fetch operation",
			},
		),
		FetchCyclesTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "fetch_cycles_total",
				Help:      "Total number of fetch cycles executed",
			},
		),
		FetchErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "reports",
				Name:      "fetch_errors_total",
				Help:      "Total number of fetch cycle errors",
			},
		),

		// IMAP connection
		IMAPConnectionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "imap",
				Name:      "connections_total",
				Help:      "Total number of IMAP connection attempts",
			},
			[]string{"status"}, // "success" or "error"
		),
		IMAPConnectionDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "imap",
				Name:      "connection_duration_seconds",
				Help:      "Duration of IMAP connection establishment",
				Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~5s
			},
		),

		// DMARC statistics (current state)
		TotalReports: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "reports_total",
				Help:      "Total number of DMARC reports in database",
			},
		),
		TotalMessages: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "messages_total",
				Help:      "Total number of messages across all DMARC reports",
			},
		),
		CompliantMessages: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "compliant_messages_total",
				Help:      "Total number of DMARC-compliant messages",
			},
		),
		ComplianceRate: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "compliance_rate",
				Help:      "Overall DMARC compliance rate (0-100)",
			},
		),
		UniqueSourceIPs: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "unique_source_ips",
				Help:      "Number of unique source IP addresses",
			},
		),
		UniqueDomains: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "unique_domains",
				Help:      "Number of unique domains",
			},
		),

		// Per-domain metrics
		MessagesByDomain: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "messages_by_domain",
				Help:      "Number of messages per domain",
			},
			[]string{"domain"},
		),
		ComplianceByDomain: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "compliance_rate_by_domain",
				Help:      "DMARC compliance rate per domain (0-100)",
			},
			[]string{"domain"},
		),
		ReportsByOrg: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "reports_by_org",
				Help:      "Number of reports per reporting organization",
			},
			[]string{"org_name"},
		),
		MessagesByDisposition: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "messages_by_disposition",
				Help:      "Number of messages per disposition type",
			},
			[]string{"disposition"}, // none, quarantine, reject
		),

		// Authentication results
		SPFResults: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "spf_results",
				Help:      "Count of SPF authentication results",
			},
			[]string{"result"}, // pass, fail, neutral, etc.
		),
		DKIMResults: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "dmarc",
				Name:      "dkim_results",
				Help:      "Count of DKIM authentication results",
			},
			[]string{"result"}, // pass, fail, neutral, etc.
		),

		// HTTP server
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Duration of HTTP requests",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		),
	}

	// Register all metrics
	registry.MustRegister(
		// Standard Go collectors
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),

		// Build info
		m.BuildInfo,

		// Report processing
		m.ReportsFetched,
		m.ReportsParsed,
		m.ReportsStored,
		m.ReportParseErrors,
		m.ReportStoreErrors,
		m.AttachmentsTotal,
		m.FetchDuration,
		m.LastFetchTimestamp,
		m.FetchCyclesTotal,
		m.FetchErrors,

		// IMAP
		m.IMAPConnectionsTotal,
		m.IMAPConnectionDuration,

		// DMARC statistics
		m.TotalReports,
		m.TotalMessages,
		m.CompliantMessages,
		m.ComplianceRate,
		m.UniqueSourceIPs,
		m.UniqueDomains,

		// Per-domain
		m.MessagesByDomain,
		m.ComplianceByDomain,
		m.ReportsByOrg,
		m.MessagesByDisposition,

		// Authentication
		m.SPFResults,
		m.DKIMResults,

		// HTTP
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
	)

	// Set build info
	m.BuildInfo.WithLabelValues(version, commit, buildDate).Set(1)

	return m
}

// Handler returns the Prometheus HTTP handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// RecordFetchDuration records the duration of a fetch operation
func (m *Metrics) RecordFetchDuration(duration time.Duration) {
	m.FetchDuration.Observe(duration.Seconds())
}

// RecordIMAPConnection records an IMAP connection attempt
func (m *Metrics) RecordIMAPConnection(success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	m.IMAPConnectionsTotal.WithLabelValues(status).Inc()
	m.IMAPConnectionDuration.Observe(duration.Seconds())
}

// UpdateStatistics updates the DMARC statistics gauges
func (m *Metrics) UpdateStatistics(totalReports, totalMessages, compliantMessages, uniqueIPs, uniqueDomains int, complianceRate float64) {
	m.TotalReports.Set(float64(totalReports))
	m.TotalMessages.Set(float64(totalMessages))
	m.CompliantMessages.Set(float64(compliantMessages))
	m.ComplianceRate.Set(complianceRate)
	m.UniqueSourceIPs.Set(float64(uniqueIPs))
	m.UniqueDomains.Set(float64(uniqueDomains))
}

// UpdateDomainMetrics updates per-domain metrics
func (m *Metrics) UpdateDomainMetrics(domain string, messages int, complianceRate float64) {
	m.MessagesByDomain.WithLabelValues(domain).Set(float64(messages))
	m.ComplianceByDomain.WithLabelValues(domain).Set(complianceRate)
}

// UpdateOrgMetrics updates per-organization metrics
func (m *Metrics) UpdateOrgMetrics(orgName string, reports int) {
	m.ReportsByOrg.WithLabelValues(orgName).Set(float64(reports))
}

// UpdateDispositionMetrics updates disposition counts
func (m *Metrics) UpdateDispositionMetrics(disposition string, count int) {
	m.MessagesByDisposition.WithLabelValues(disposition).Set(float64(count))
}

// UpdateAuthResults updates SPF and DKIM result counts
func (m *Metrics) UpdateAuthResults(spfResults, dkimResults map[string]int) {
	for result, count := range spfResults {
		m.SPFResults.WithLabelValues(result).Set(float64(count))
	}
	for result, count := range dkimResults {
		m.DKIMResults.WithLabelValues(result).Set(float64(count))
	}
}

// HTTPMiddleware wraps an HTTP handler with metrics instrumentation
func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip metrics endpoint to avoid recursion
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		m.HTTPRequestsInFlight.Inc()
		defer m.HTTPRequestsInFlight.Dec()

		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		path := normalizePath(r.URL.Path)

		m.HTTPRequestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(wrapped.statusCode)).Inc()
		m.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(duration.Seconds())
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// normalizePath normalizes URL paths to prevent high cardinality
func normalizePath(path string) string {
	// Normalize common API paths
	switch {
	case path == "/":
		return "/"
	case path == "/api/statistics":
		return "/api/statistics"
	case path == "/api/reports":
		return "/api/reports"
	case path == "/api/top-sources":
		return "/api/top-sources"
	case len(path) > 13 && path[:13] == "/api/reports/":
		return "/api/reports/:id"
	case path == "/metrics", path == "/healthz", path == "/readyz":
		return path
	default:
		// Group static assets
		if len(path) >= 7 && path[:7] == "/assets" {
			return "/assets/*"
		}
		return "/other"
	}
}
