package server

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"

	"github.com/dmarcguardhq/parse-dmarc/internal/logger"
)

var (
	log *zerolog.Logger

	// CLI flag destinations.
	configPath     string
	genConfig      bool
	fetchOnce      bool
	serveOnly      bool
	fetchInterval  int
	metricsEnabled bool
	mcpMode        bool
	mcpHTTPAddr    string

	// MCP OAuth flags.
	mcpOAuthEnabled       bool
	mcpOAuthIssuer        string
	mcpOAuthAudience      string
	mcpOAuthClientID      string
	mcpOAuthClientSecret  string
	mcpOAuthScopes        string
	mcpOAuthIntrospection string
	mcpOAuthResourceName  string
	mcpOAuthInsecure      bool
)

// Command returns the root parse-dmarc command, including its flags, the
// default fetch/serve action, and the version subcommand.
func Command(version, commit, date, builtBy string) *cli.Command {
	return &cli.Command{
		Name:                  "parse-dmarc",
		Usage:                 "Parse and analyze DMARC reports from IMAP mailbox",
		Version:               version,
		EnableShellCompletion: true,
		Suggest:               true,
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			log = logger.NewLogger("info", false)
			return ctx, nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Path to configuration file",
				Value:       "config.json",
				Sources:     cli.EnvVars("PARSE_DMARC_CONFIG"),
				Destination: &configPath,
			},
			&cli.BoolFlag{
				Name:        "gen-config",
				Usage:       "Generate sample configuration file",
				Sources:     cli.EnvVars("PARSE_DMARC_GEN_CONFIG"),
				Destination: &genConfig,
			},
			&cli.BoolFlag{
				Name:        "fetch-once",
				Usage:       "Fetch reports once and exit",
				Sources:     cli.EnvVars("PARSE_DMARC_FETCH_ONCE"),
				Destination: &fetchOnce,
			},
			&cli.BoolFlag{
				Name:        "serve-only",
				Usage:       "Only serve the dashboard without fetching",
				Sources:     cli.EnvVars("PARSE_DMARC_SERVE_ONLY"),
				Destination: &serveOnly,
			},
			&cli.IntFlag{
				Name:        "fetch-interval",
				Usage:       "Interval in seconds between fetch operations",
				Value:       300,
				Sources:     cli.EnvVars("PARSE_DMARC_FETCH_INTERVAL", "FETCH_INTERVAL"),
				Destination: &fetchInterval,
			},
			&cli.BoolFlag{
				Name:        "metrics",
				Usage:       "Enable Prometheus metrics endpoint at /metrics",
				Value:       true,
				Sources:     cli.EnvVars("PARSE_DMARC_METRICS"),
				Destination: &metricsEnabled,
			},
			&cli.BoolFlag{
				Name:        "mcp",
				Usage:       "Run as MCP (Model Context Protocol) server over stdio",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP"),
				Destination: &mcpMode,
			},
			&cli.StringFlag{
				Name:        "mcp-http",
				Usage:       "Run MCP server over Streamable HTTP at the specified address (e.g., :8081)",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_HTTP"),
				Destination: &mcpHTTPAddr,
			},
			// OAuth2 flags for MCP HTTP server.
			&cli.BoolFlag{
				Name:        "mcp-oauth",
				Usage:       "Enable OAuth2 authentication for MCP HTTP server",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH"),
				Destination: &mcpOAuthEnabled,
			},
			&cli.StringFlag{
				Name:        "mcp-oauth-issuer",
				Usage:       "OAuth2/OIDC issuer URL (e.g., https://auth.example.com/realms/master)",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_ISSUER"),
				Destination: &mcpOAuthIssuer,
			},
			&cli.StringFlag{
				Name:        "mcp-oauth-audience",
				Usage:       "Expected audience claim in tokens (usually the MCP server URL)",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_AUDIENCE"),
				Destination: &mcpOAuthAudience,
			},
			&cli.StringFlag{
				Name:        "mcp-oauth-client-id",
				Usage:       "OAuth2 client ID for token introspection",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_CLIENT_ID"),
				Destination: &mcpOAuthClientID,
			},
			&cli.StringFlag{
				Name:        "mcp-oauth-client-secret",
				Usage:       "OAuth2 client secret for token introspection",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_CLIENT_SECRET"),
				Destination: &mcpOAuthClientSecret,
			},
			&cli.StringFlag{
				Name:        "mcp-oauth-scopes",
				Usage:       "Required scopes (comma-separated, e.g., mcp:tools)",
				Value:       "mcp:tools",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_SCOPES"),
				Destination: &mcpOAuthScopes,
			},
			&cli.StringFlag{
				Name:        "mcp-oauth-introspection-endpoint",
				Usage:       "Token introspection endpoint URL (if set, uses introspection instead of JWT validation)",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_INTROSPECTION_ENDPOINT"),
				Destination: &mcpOAuthIntrospection,
			},
			&cli.StringFlag{
				Name:        "mcp-oauth-resource-name",
				Usage:       "Human-readable name for the MCP server (for metadata)",
				Value:       "Parse-DMARC MCP Server",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_RESOURCE_NAME"),
				Destination: &mcpOAuthResourceName,
			},
			&cli.BoolFlag{
				Name:        "mcp-oauth-insecure",
				Usage:       "Skip TLS certificate verification (development only)",
				Sources:     cli.EnvVars("PARSE_DMARC_MCP_OAUTH_INSECURE"),
				Destination: &mcpOAuthInsecure,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, version, commit, date)
		},
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Show version information",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Printf("Version:    %s\n", version)
					fmt.Printf("Commit:     %s\n", commit)
					fmt.Printf("Build Date: %s\n", date)
					fmt.Printf("Built By:   %s\n", builtBy)
					return nil
				},
			},
		},
	}
}
