package server

import (
	"context"
	"os/signal"
	"strings"
	"syscall"

	mcpserver "github.com/dmarcguardhq/parse-dmarc/internal/mcp"
	"github.com/dmarcguardhq/parse-dmarc/internal/mcp/oauth"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
)

func runMCPServer(ctx context.Context, store *storage.Storage, version string) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Build OAuth config if enabled.
	var oauthCfg *oauth.Config
	if mcpOAuthEnabled {
		var scopes []string
		if mcpOAuthScopes != "" {
			for s := range strings.SplitSeq(mcpOAuthScopes, ",") {
				scopes = append(scopes, strings.TrimSpace(s))
			}
		}

		// Determine resource server URL and audience from HTTP address.
		var resourceServerURL, audience string
		if mcpOAuthAudience != "" {
			resourceServerURL = mcpOAuthAudience
			audience = mcpOAuthAudience
		} else if mcpHTTPAddr != "" {
			resourceServerURL = "http://localhost" + mcpHTTPAddr
			audience = resourceServerURL
		}

		oauthCfg = &oauth.Config{
			Enabled:               true,
			Issuer:                mcpOAuthIssuer,
			Audience:              audience,
			ClientID:              mcpOAuthClientID,
			ClientSecret:          mcpOAuthClientSecret,
			RequiredScopes:        scopes,
			IntrospectionEndpoint: mcpOAuthIntrospection,
			ResourceServerURL:     resourceServerURL,
			ResourceName:          mcpOAuthResourceName,
			InsecureSkipVerify:    mcpOAuthInsecure,
		}
	}

	mcpCfg := &mcpserver.Config{
		Version:  version,
		HTTPAddr: mcpHTTPAddr,
		Logger:   log,
		OAuth:    oauthCfg,
	}

	server := mcpserver.NewServer(store, mcpCfg)

	// If HTTP address is specified, run HTTP server.
	// Otherwise, run over stdio.
	if mcpHTTPAddr != "" {
		return server.RunHTTP(ctx, mcpCfg.HTTPAddr, oauthCfg)
	}
	return server.RunStdio(ctx)
}
