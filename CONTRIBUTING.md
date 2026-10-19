# Contributing to Parse DMARC

Issues and pull requests are welcome. This page is the local setup and the house rules.

## Prerequisites

- Go 1.25 or newer (`go.mod` pins the exact version)
- Bun 1.x for the Vue frontend
- `just` for the build recipes
- Docker with Compose, only for the local IMAP fixture

`flake.nix` provides all of them: `nix develop`.

## Setup

```bash
git clone https://github.com/dmarcguardhq/parse-dmarc.git
cd parse-dmarc
just install-deps      # go mod tidy + bun install
just build             # frontend into internal/api/dist, then the Go binary at bin/parse-dmarc
```

## Layout

```
cmd/server/          CLI flags and the run loop
internal/api/        HTTP server, embeds the built frontend from internal/api/dist
internal/config/     config.json and environment parsing
internal/imap/       IMAP client, attachment unwrapping
internal/maildir/    local Maildir intake (new/ -> cur/)
internal/mcp/        MCP server and tools
internal/metrics/    Prometheus metrics
internal/parser/     aggregate report XML parser
internal/storage/    SQLite, pure-Go by default, cgo behind the `cgo` build tag
src/                 Vue 3 dashboard
contrib/dovecot/     local IMAP fixture with a seeded sample report
contrib/nginx/       dashboard behind nginx under /dmarc/; e2e.sh checks it in headless Chrome
deploy/              CapRover, Coolify, Dokploy, DigitalOcean templates
grafana/             dashboard and provisioning
docs/                METRICS.md, MCP.md
```

## Day to day

```bash
just dev               # backend with live reload (air)
just frontend-dev      # Vite dev server for the dashboard
just test              # go test -v ./...
docker compose up      # Dovecot on localhost:11143, user dmarc / dev, one report in INBOX
go run . --config config.dev.json   # fetch from that fixture
```

`config.dev.json` is already pointed at the fixture. `parse-dmarc --gen-config` writes a fresh `config.json` template.

## Pull requests

- One change per PR, with a test when the change has logic in it.
- Conventional Commits in the title (`feat:`, `fix:`, `docs:`); release-please builds the changelog from them.
- `bunx prettier -w .` before pushing; CI checks it and fails on any diff.
- `main` merges through a merge queue: once `ci-gate` is green and the PR is approved, a maintainer queues it, and CI runs again on the exact commit that lands.
- Update the README or `docs/` when you change a flag, an env var, a metric or a tool.

## Open asks

In the order people ask: TLS-RPT reports (#154), whois on sending sources (#143). Then, from ROADMAP.md: failure reports (RUF), CSV and JSON export, alerting.

## License

Apache-2.0. By contributing you agree your contribution is licensed the same way.
