// Package mcp provides an MCP (Model Context Protocol) server for parse-dmarc.
// It exposes DMARC report data and analysis tools to MCP clients like Claude.
package mcp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dmarcguardhq/parse-dmarc/internal/mcp/oauth"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
)

// Server wraps the MCP server with storage access.
type Server struct {
	mcpServer *mcp.Server
	store     *storage.Storage
	logger    *zerolog.Logger
}

// Config holds MCP server configuration.
type Config struct {
	// HTTPAddr is the address for HTTP/SSE transport. If empty, only stdio is used.
	HTTPAddr string
	// Version info for the server implementation.
	Version string
	// Logger for the server
	Logger *zerolog.Logger
	// OAuth holds OAuth2 configuration for protected HTTP endpoints.
	OAuth *oauth.Config
}

// NewServer creates a new MCP server with all DMARC tools registered.
func NewServer(store *storage.Storage, cfg *Config) *Server {
	version := cfg.Version
	if version == "" {
		version = "dev"
	}

	opts := &mcp.ServerOptions{
		Instructions: `Parse-DMARC MCP Server provides tools to query and analyze DMARC aggregate reports.

Available tools:
- get_statistics: Get overall DMARC compliance statistics
- get_reports: List DMARC reports with pagination
- get_report_by_id: Get detailed report by ID
- get_top_source_ips: Get top sending IP addresses
- get_domain_stats: Get per-domain compliance statistics
- get_org_stats: Get statistics by reporting organization
- get_spf_stats: Get SPF authentication result statistics
- get_dkim_stats: Get DKIM authentication result statistics
- parse_dmarc_report: Parse a raw DMARC XML report`,
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "parse-dmarc",
			Version: version,
		},
		opts,
	)

	s := &Server{
		mcpServer: mcpServer,
		store:     store,
		logger:    cfg.Logger,
	}

	if s.logger != nil {
		mcpServer.AddReceivingMiddleware(s.loggingMiddleware())
	}

	// Register all tools
	s.registerTools()

	return s
}

func (s *Server) loggingMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			logger := s.logger.With().
				Str("method", method).
				Str("session_id", req.GetSession().ID()).
				Logger()

			logger.Debug().
				Bool("has_params", req.GetParams() != nil).
				Msg("MCP request started")

			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				logger.Info().
					Str("tool", ctr.Params.Name).
					Msg("calling tool")
			}

			start := time.Now()
			result, err := next(ctx, method, req)
			duration := time.Since(start)

			if err != nil {
				logger.Error().
					Err(err).
					Dur("duration", duration).
					Msg("MCP request failed")
			} else {
				logEvent := logger.Debug().
					Dur("duration", duration).
					Bool("has_result", result != nil)

				if ctr, ok := result.(*mcp.CallToolResult); ok {
					logEvent = logEvent.Bool("is_error", ctr.IsError)
				}

				logEvent.Msg("MCP request completed")
			}

			return result, err
		}
	}
}

// RunStdio runs the MCP server over stdio transport.
func (s *Server) RunStdio(ctx context.Context) error {
	if s.logger != nil {
		s.logger.Info().Msg("starting MCP server over stdio")
	}
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// RunHTTP runs the MCP server over HTTP/SSE transport.
func (s *Server) RunHTTP(ctx context.Context, addr string, oauthCfg *oauth.Config) error {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)

	// Build the HTTP handler with optional OAuth middleware
	var handler http.Handler = mcpHandler

	// Create a mux to handle both the MCP endpoint and the metadata endpoint
	mux := http.NewServeMux()

	// Check if OAuth is enabled and properly configured
	if oauthCfg != nil && oauthCfg.Enabled {
		if err := oauthCfg.Validate(); err != nil {
			return fmt.Errorf("OAuth config validation failed: %w", err)
		}

		// Register the Protected Resource Metadata endpoint (RFC 9728)
		mux.Handle(oauth.MetadataPath, oauth.MetadataHandler(oauthCfg))

		// Create the token verifier and middleware
		verifier := oauth.NewVerifier(oauthCfg)
		authMiddleware := oauth.NewBearerAuthMiddleware(oauthCfg, verifier, s.logger)

		// Wrap the MCP handler with auth middleware
		handler = authMiddleware.Wrap(mcpHandler)

		if s.logger != nil {
			s.logger.Info().
				Str("issuer", oauthCfg.Issuer).
				Str("audience", oauthCfg.Audience).
				Strs("scopes", oauthCfg.RequiredScopes).
				Str("metadata_url", oauth.GetMetadataURL(oauthCfg.ResourceServerURL)).
				Msg("OAuth2 authentication enabled")
		}
	}

	// Register the MCP handler (with or without auth middleware)
	mux.Handle("/", handler)

	if s.logger != nil {
		s.logger.Info().Str("addr", addr).Msg("starting MCP server over HTTP")
	}

	server := &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MiB
	}

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		if s.logger != nil {
			s.logger.Info().Msg("shutting down MCP HTTP server")
		}
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// registerTools registers all DMARC analysis tools with the MCP server.
func (s *Server) registerTools() {
	// get_statistics - Get overall DMARC compliance statistics
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_statistics",
		Description: "Get overall DMARC compliance statistics including total reports, messages, compliance rate, unique source IPs, and unique domains.",
	}, s.getStatistics)

	// get_reports - List DMARC reports with pagination
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_reports",
		Description: "List DMARC reports with pagination. Returns report summaries including ID, organization, domain, date range, message counts, and compliance rate.",
	}, s.getReports)

	// get_report_by_id - Get detailed report by ID
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_report_by_id",
		Description: "Get full details of a specific DMARC report by its database ID. Returns the complete parsed report including all records and authentication results.",
	}, s.getReportByID)

	// get_top_source_ips - Get top sending IP addresses
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_top_source_ips",
		Description: "Get the top sending IP addresses ranked by message count. Shows pass/fail breakdown for each IP to help identify potential spoofing sources.",
	}, s.getTopSourceIPs)

	// get_domain_stats - Get per-domain compliance statistics
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_domain_stats",
		Description: "Get DMARC compliance statistics grouped by domain. Shows total messages, compliant messages, and compliance rate for each domain.",
	}, s.getDomainStats)

	// get_org_stats - Get statistics by reporting organization
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_org_stats",
		Description: "Get report counts grouped by reporting organization (e.g., Google, Microsoft, Yahoo). Helps understand which email providers are sending DMARC reports.",
	}, s.getOrgStats)

	// get_spf_stats - Get SPF authentication result statistics
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_spf_stats",
		Description: "Get SPF (Sender Policy Framework) authentication result statistics. Shows counts for each result type (pass, fail, softfail, neutral, etc.).",
	}, s.getSPFStats)

	// get_dkim_stats - Get DKIM authentication result statistics
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_dkim_stats",
		Description: "Get DKIM (DomainKeys Identified Mail) authentication result statistics. Shows counts for each result type (pass, fail, none, etc.).",
	}, s.getDKIMStats)

	// parse_dmarc_report - Parse a raw DMARC XML report
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "parse_dmarc_report",
		Description: "Parse a raw DMARC aggregate report from XML data. Accepts gzip, zip, or plain XML. The report_data should be base64 encoded. Returns the parsed report structure.",
	}, s.parseDMARCReport)
}
