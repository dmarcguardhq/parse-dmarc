# Parse DMARC - Project Guide

A Go application that fetches DMARC reports from IMAP mailboxes, parses them, and displays them in a Vue.js dashboard. Also supports MCP (Model Context Protocol) for AI assistant integration.

## Tech Stack

- **Backend**: Go 1.25.4 (see go.mod for exact version)
- **Frontend**: Vue.js 3 with Vite
- **Database**: SQLite (supports both CGO and pure-Go variants)
- **Package Manager**: Bun (for frontend)
- **Task Runner**: Just (Justfile)
- **CLI Framework**: urfave/cli/v3
- **JSON Library**: goccy/go-json (high-performance)
- **Logging**: rs/zerolog (structured logging)
- **Metrics**: Prometheus client_golang
- **MCP SDK**: modelcontextprotocol/go-sdk

## Project Structure

```
parse-dmarc/
├── main.go                    # Main application entry point
├── internal/
│   ├── api/                   # REST API server and embedded frontend
│   │   └── server.go          # HTTP server, routes, metrics middleware
│   ├── config/                # Configuration management (JSON + env vars)
│   │   └── config.go          # Config loading and validation
│   ├── imap/                  # IMAP client for fetching emails
│   │   └── client.go          # Email fetching logic
│   ├── logger/                # Structured logging setup
│   │   └── logger.go          # Zerolog configuration
│   ├── mcp/                   # MCP (Model Context Protocol) server
│   │   ├── server.go          # MCP server (stdio and HTTP/SSE)
│   │   ├── tools.go           # MCP tool implementations
│   │   └── oauth/             # OAuth2 authentication for MCP
│   │       ├── config.go      # OAuth configuration
│   │       ├── middleware.go  # Bearer auth middleware
│   │       ├── metadata.go    # RFC 9728 metadata endpoint
│   │       └── verifier.go    # Token verification
│   ├── metrics/               # Prometheus metrics
│   │   └── metrics.go         # Metrics definitions and HTTP middleware
│   ├── parser/                # DMARC XML parser
│   │   ├── dmarc.go           # Parsing logic
│   │   └── dmarc_test.go      # Parser tests
│   └── storage/               # SQLite database layer
│       ├── common.go          # Shared SQL queries and types
│       ├── sqlite_cgo.go      # CGO SQLite (mattn/go-sqlite3)
│       └── sqlite_no_cgo.go   # Pure Go SQLite (modernc.org/sqlite)
├── src/                       # Vue.js 3 frontend source
│   ├── App.vue                # Main application component
│   ├── main.js                # Vue entry point
│   ├── assets/
│   │   └── base.css           # Base styles
│   ├── stores/                # Pinia-like state management
│   │   ├── index.js           # Store exports
│   │   ├── theme.js           # Theme (dark/light/system) store
│   │   └── settings.js        # API endpoint settings store
│   └── components/
│       ├── dashboard/
│       │   ├── DashboardHero.vue   # Dashboard header/hero section
│       │   ├── RecentReports.vue   # Recent reports list
│       │   └── ReportDrawer.vue    # Report detail drawer
│       ├── settings/
│       │   └── SettingsModal.vue   # Settings modal (theme, API endpoint)
│       └── tools/
│           └── DnsGenerator.vue    # DMARC DNS record generator
├── public/                    # Static frontend assets (favicons, logos)
├── assets/                    # Project assets (screenshots, images)
├── deploy/                    # Deployment configurations
│   ├── coolify.yaml           # Coolify deployment
│   ├── captain-definition     # CapRover deployment
│   ├── digitalocean/          # DigitalOcean Droplet/Marketplace
│   │   ├── packer.pkr.hcl     # Packer image build
│   │   ├── marketplace.yaml   # DO Marketplace metadata
│   │   └── scripts/           # Setup scripts
│   └── dokploy/               # Dokploy deployment
│       ├── template.toml
│       └── docker-compose.yml
├── grafana/                   # Grafana monitoring
│   ├── dashboard.json         # Pre-built dashboard
│   └── provisioning.yaml      # Auto-provisioning config
├── scripts/                   # Utility scripts
├── Justfile                   # Build commands
├── Dockerfile                 # Multi-stage Docker build
├── compose.yml                # Docker Compose for local dev
├── parse-dmarc.service        # systemd service file
├── zeabur.yml                 # Zeabur deployment template
├── render.yaml                # Render deployment config
├── Northflank.json            # Northflank deployment config
├── ROADMAP.md                 # Product roadmap
├── CONTRIBUTING.md            # Contribution guidelines
├── .goreleaser.yml            # Release automation
└── .github/workflows/ci.yml   # CI/CD pipeline
```

## Development Commands

```bash
# Install all dependencies (Go + Node)
just install-deps

# Build full application (frontend + backend)
just build

# Build with CGO (for native SQLite)
just build-cgo

# Run development server with hot reload (uses air)
just dev

# Run frontend dev server only
just frontend-dev

# Run tests
just test

# Generate config file
just config

# Clean build artifacts
just clean

# Install binary to /usr/local/bin
just install

# Update Zeabur template
just update-zeabur-template
```

## Building

The build process:

1. Frontend is built with `bun run build`
2. Frontend dist is copied to `internal/api/dist/`
3. Go binary embeds the frontend and serves it
4. Output binary: `bin/parse-dmarc`

Two build modes:

- `just build` - Pure Go (CGO_ENABLED=0), uses modernc.org/sqlite
- `just build-cgo` - CGO enabled, uses mattn/go-sqlite3

## Testing

```bash
# Run all Go tests
go test -v ./...

# Run tests for specific package
go test -v ./internal/parser/...
```

## Running the Application

### Standard Mode (Dashboard + IMAP Fetching)

```bash
# Continuous fetching and dashboard
./parse-dmarc --config config.json

# Fetch once and exit
./parse-dmarc --config config.json --fetch-once

# Dashboard only (no IMAP fetching)
./parse-dmarc --config config.json --serve-only
```

### MCP Mode (AI Assistant Integration)

```bash
# MCP over stdio (for Claude Desktop, etc.)
./parse-dmarc --mcp

# MCP over HTTP/SSE
./parse-dmarc --mcp-http :8081

# MCP with OAuth2 authentication
./parse-dmarc --mcp-http :8081 --mcp-oauth \
  --mcp-oauth-issuer https://auth.example.com \
  --mcp-oauth-audience https://mcp.example.com
```

## CLI Flags

| Flag                                 | Env Var                                        | Description                                           |
| ------------------------------------ | ---------------------------------------------- | ----------------------------------------------------- |
| `--config, -c`                       | `PARSE_DMARC_CONFIG`                           | Config file path (default: config.json)               |
| `--gen-config`                       | `PARSE_DMARC_GEN_CONFIG`                       | Generate sample config                                |
| `--fetch-once`                       | `PARSE_DMARC_FETCH_ONCE`                       | Fetch reports once and exit                           |
| `--serve-only`                       | `PARSE_DMARC_SERVE_ONLY`                       | Dashboard only, no fetching                           |
| `--fetch-interval`                   | `PARSE_DMARC_FETCH_INTERVAL`                   | Fetch interval in seconds (default: 300)              |
| `--metrics`                          | `PARSE_DMARC_METRICS`                          | Enable Prometheus metrics (default: true)             |
| `--mcp`                              | `PARSE_DMARC_MCP`                              | Run as MCP server over stdio                          |
| `--mcp-http`                         | `PARSE_DMARC_MCP_HTTP`                         | Run MCP over HTTP at address                          |
| `--mcp-oauth`                        | `PARSE_DMARC_MCP_OAUTH`                        | Enable OAuth2 for MCP HTTP                            |
| `--mcp-oauth-issuer`                 | `PARSE_DMARC_MCP_OAUTH_ISSUER`                 | OAuth2/OIDC issuer URL                                |
| `--mcp-oauth-audience`               | `PARSE_DMARC_MCP_OAUTH_AUDIENCE`               | Expected token audience                               |
| `--mcp-oauth-client-id`              | `PARSE_DMARC_MCP_OAUTH_CLIENT_ID`              | OAuth2 client ID for token introspection              |
| `--mcp-oauth-client-secret`          | `PARSE_DMARC_MCP_OAUTH_CLIENT_SECRET`          | OAuth2 client secret for token introspection          |
| `--mcp-oauth-scopes`                 | `PARSE_DMARC_MCP_OAUTH_SCOPES`                 | Required scopes (comma-separated, default: mcp:tools) |
| `--mcp-oauth-introspection-endpoint` | `PARSE_DMARC_MCP_OAUTH_INTROSPECTION_ENDPOINT` | Token introspection endpoint URL                      |
| `--mcp-oauth-resource-name`          | `PARSE_DMARC_MCP_OAUTH_RESOURCE_NAME`          | Human-readable name for MCP server metadata           |
| `--mcp-oauth-insecure`               | `PARSE_DMARC_MCP_OAUTH_INSECURE`               | Skip TLS certificate verification (dev only)          |

## Code Style

- Go: Standard gofmt formatting, golangci-lint for linting
- Frontend: Vue 3 Composition API, Prettier for formatting
- Pre-commit hooks configured in `.pre-commit-config.yaml`

## Key Files

### Backend

- `main.go` - CLI entry point with flag parsing, signal handling
- `internal/api/server.go` - HTTP server, API routes, metrics middleware
- `internal/config/config.go` - Configuration loading (JSON + env vars)
- `internal/parser/dmarc.go` - DMARC XML parsing (gzip, zip, raw XML)
- `internal/imap/client.go` - IMAP email fetching
- `internal/storage/common.go` - SQL queries, data types
- `internal/mcp/server.go` - MCP server implementation
- `internal/mcp/tools.go` - MCP tool handlers
- `internal/metrics/metrics.go` - Prometheus metrics definitions

### Frontend

- `src/App.vue` - Main Vue.js dashboard component
- `src/stores/theme.js` - Theme state management (light/dark/system)
- `src/stores/settings.js` - API endpoint settings management
- `src/components/dashboard/DashboardHero.vue` - Statistics overview
- `src/components/dashboard/RecentReports.vue` - Reports list
- `src/components/dashboard/ReportDrawer.vue` - Report detail view
- `src/components/settings/SettingsModal.vue` - Settings dialog (theme, API endpoint)
- `src/components/tools/DnsGenerator.vue` - DMARC DNS record generator

## API Endpoints

### REST API

- `GET /api/statistics` - Dashboard statistics
- `GET /api/reports` - List reports (paginated: `?limit=50&offset=0`)
- `GET /api/reports/:id` - Single report details
- `GET /api/top-sources` - Top sending source IPs

### Metrics

- `GET /metrics` - Prometheus metrics endpoint

## MCP Tools

When running in MCP mode, the following tools are available:

| Tool                 | Description                          |
| -------------------- | ------------------------------------ |
| `get_statistics`     | Overall DMARC compliance statistics  |
| `get_reports`        | List reports with pagination         |
| `get_report_by_id`   | Get detailed report by ID            |
| `get_top_source_ips` | Top sending IP addresses             |
| `get_domain_stats`   | Per-domain compliance stats          |
| `get_org_stats`      | Stats by reporting organization      |
| `get_spf_stats`      | SPF authentication result stats      |
| `get_dkim_stats`     | DKIM authentication result stats     |
| `parse_dmarc_report` | Parse raw DMARC XML (base64 encoded) |

## Prometheus Metrics

Key metrics exposed at `/metrics`:

### Report Processing

- `parse_dmarc_reports_fetched_total` - Reports fetched from IMAP
- `parse_dmarc_reports_parsed_total` - Successfully parsed reports
- `parse_dmarc_reports_stored_total` - Reports saved to database
- `parse_dmarc_reports_fetch_duration_seconds` - Fetch operation duration

### DMARC Statistics

- `parse_dmarc_dmarc_reports_total` - Total reports in database
- `parse_dmarc_dmarc_messages_total` - Total messages processed
- `parse_dmarc_dmarc_compliance_rate` - Overall compliance rate
- `parse_dmarc_dmarc_messages_by_domain{domain}` - Per-domain message count
- `parse_dmarc_dmarc_compliance_rate_by_domain{domain}` - Per-domain compliance

### HTTP Server

- `parse_dmarc_http_requests_total{method,path,status}` - Request count
- `parse_dmarc_http_request_duration_seconds{method,path}` - Request latency

## Frontend Features

The Vue.js dashboard includes:

- **Dashboard Hero** - Overview statistics with compliance rate
- **Recent Reports** - Paginated list of DMARC reports
- **Report Drawer** - Detailed view of individual reports
- **Top Sources** - Visualization of top sending source IPs
- **DMARC DNS Generator** - Interactive tool to generate DMARC DNS TXT records
- **Settings Modal** - Theme switching (light/dark/system) and custom API endpoint configuration

## Configuration

Config via JSON file or environment variables (using caarlos0/env):

```json
{
  "imap": {
    "host": "imap.example.com",
    "port": 993,
    "username": "dmarc@example.com",
    "password": "your-password",
    "mailbox": "INBOX",
    "use_tls": true,
    "starttls": false,
    "tls_ca_file": "",
    "tls_skip_verify": false,
    "mark_as_seen": true,
    "processed_mailbox": ""
  },
  "database": {
    "path": "~/.parse-dmarc/db.sqlite"
  },
  "server": {
    "host": "0.0.0.0",
    "port": 8080
  }
}
```

Environment variables:

- `DATABASE_PATH`
- `IMAP_HOST`
- `IMAP_MAILBOX`
- `IMAP_MARK_AS_SEEN`
- `IMAP_PASSWORD`
- `IMAP_PORT`
- `IMAP_PROCESSED_MAILBOX`
- `IMAP_STARTTLS` - dial plaintext (port 143) and upgrade via STARTTLS; wins over `IMAP_USE_TLS`
- `IMAP_TLS_CA_FILE` - PEM bundle for an internal CA
- `IMAP_TLS_SKIP_VERIFY` - disable IMAP certificate verification (last resort)
- `IMAP_USE_TLS`
- `IMAP_USERNAME`
- `SERVER_HOST`
- `SERVER_PORT`

## Deployment Options

### Docker

```bash
docker run -d -p 8080:8080 \
  -e IMAP_HOST=imap.example.com \
  -e IMAP_USERNAME=dmarc@example.com \
  -e IMAP_PASSWORD=secret \
  ghcr.io/meysam81/parse-dmarc:latest
```

### Docker Compose

See `compose.yml` for a complete example with persistence.

### Systemd

See `parse-dmarc.service` for systemd service configuration.

### Platform-Specific

- **DigitalOcean**: `deploy/digitalocean/` - Packer template for Marketplace
- **Dokploy**: `deploy/dokploy/` - Docker Compose template
- **Coolify**: `deploy/coolify.yaml`
- **CapRover**: `deploy/captain-definition`
- **Zeabur**: `zeabur.yml` - Zeabur platform template
- **Render**: `render.yaml` - Render.com configuration
- **Northflank**: `Northflank.json` - Northflank configuration

## CI/CD

- GitHub Actions workflow in `.github/workflows/ci.yml`
- Prettier for code formatting (auto-commit in PRs)
- golangci-lint for Go linting
- Docker build with cosign signing
- Kubescape security scanning
- Release automation via release-please and goreleaser
- Multi-platform Docker images (amd64, arm64)
- Lockfile: bun.lock only (Docker + CI use `bun install --frozen-lockfile`). package-lock.json was removed so Renovate updates bun.lock; do not reintroduce it.
- GitHub release notes footer (install/verify commands) lives in `.github/release-notes.md`, a goreleaser template passed via `--release-notes-tmpl` in `build-prod` with `release.mode: append`. release-please `pull-request-footer` only affects the release PR body, never the GitHub release. Templates use `missingkey=error`: an unknown `{{ .Field }}` fails the release, so render-check after editing. `--snapshot` never renders it (changelog pipe is skipped in snapshot mode); to see the real output run a non-snapshot `goreleaser release --skip=validate,announce --release-notes-tmpl ...` with a throwaway `--config` that has `release.disable: true` and a custom `publishers` entry whose `env` carries `NOTES={{ .ReleaseNotes }}`.
- Nix install path is binary, not source: `nix:` in `.goreleaser.yml` pushes `pkgs/parse-dmarc/default.nix` (fetchurl of the release archives, per-platform sha256, shell completions via `extra_install`) to `dmarcguardhq/nur`, whose `flake.nix` wrapper serves `nix run github:dmarcguardhq/nur#parse-dmarc`. The pipe silently skips when `nix-hash` is missing, hence `cachix/install-nix-action` in both goreleaser jobs. A source-built `packages.default` in `flake.nix` is not viable: the Go binary embeds the Vite build and nixpkgs has no bun-lockfile fetcher (`buildNpmPackage` needs the banned package-lock.json), which is why the old one was removed. Local check: generate with a non-snapshot skip-publish run, swap in the published `checksums.txt` hashes (`nix hash convert --hash-algo sha256 --to nix32 <hex>`), `nix build` in a git-tracked copy of the nur flake.
- Local `goreleaser release --snapshot` needs a bun that reads `bun.lock` v2: the flake's bun (1.3.8) fails `--frozen-lockfile` with "Unknown lockfile version"; prepend `$(nix build nixpkgs#bun --no-link --print-out-paths)/bin` to PATH.

## Roadmap

See `ROADMAP.md` for the comprehensive product roadmap including:

- **Phase 1**: Delightful Defaults (dark mode, DNS generator, health score, exports)
- **Phase 2**: Proactive Intelligence (alerting, trends, GeoIP maps, DNS validator)
- **Phase 3**: Enterprise Ready (auth, multi-org, RBAC, Prometheus metrics)
- **Phase 4**: AI-Powered Security (AI assistant, anomaly detection, forensic reports)

## Contributing

See `CONTRIBUTING.md` for development setup and contribution guidelines. Key areas:

- Forensic Reports (RUF support)
- OAuth2 for IMAP authentication
- CSV/JSON export functionality
- Email alerts for compliance issues
- Historical trend analysis
- Test coverage improvements

## Architecture Notes

### Database Schema

- `reports` table: Stores report metadata and raw JSON
- `records` table: Stores individual record data per report
- Build tags (`cgo`/`!cgo`) select SQLite driver at compile time

### Frontend Embedding

The Vue.js frontend is built to `dist/`, copied to `internal/api/dist/`, and embedded via Go's `embed` directive. The binary is self-contained.

### State Management

The frontend uses a custom reactive store pattern (similar to Pinia):

- `theme.js` - Manages light/dark/system theme with localStorage persistence
- `settings.js` - Manages custom API endpoint with validation and connection testing

### MCP Integration

The MCP server uses the official `modelcontextprotocol/go-sdk`. It supports:

- **stdio transport**: For desktop apps like Claude Desktop
- **HTTP/SSE transport**: For web-based MCP clients
- **OAuth2**: Optional authentication via OIDC/token introspection
