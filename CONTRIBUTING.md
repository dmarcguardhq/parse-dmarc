# Contributing to Parse DMARC

Issues and pull requests are welcome. This page is the local setup and the house rules.

## Prerequisites

- Go 1.26 or newer
- Bun for the Vue frontend, a current release (CI installs the latest)
- `just` for the build recipes
- Docker with Compose, for the local IMAP fixture and `contrib/nginx/e2e.sh`

## Setup

```bash
git clone https://github.com/dmarcguardhq/parse-dmarc.git
cd parse-dmarc
just install-deps      # go mod tidy + bun install
just build             # frontend into internal/api/dist, then the Go binary at bin/parse-dmarc
```

## Day to day

```bash
just dev               # backend with live reload (air)
just frontend-dev      # Vite dev server for the dashboard
just test              # go test -v ./...; CI runs go test -race -count=1 ./...
docker compose up      # Dovecot on localhost:11143, user dmarc / dev, one report in INBOX
go run . --config config.dev.json   # fetch from that fixture
```

`config.dev.json` is already pointed at the fixture. `parse-dmarc --gen-config` writes a fresh `config.json` template.

## Pull requests

- One change per PR, with a test when the change has logic in it.
- Conventional Commits in the title (`feat:`, `fix:`, `docs:`); release-please builds the changelog from them.
- `bunx prettier -w .` before pushing; CI checks it and fails on any diff.
- `main` merges through a merge queue: once `ci-gate` is green, a maintainer queues the PR, and CI runs again on the exact commit that lands.
- Update the README or `docs/` when you change a flag, an env var, a metric or a tool.

## Open asks

In the order people ask: TLS-RPT reports ([#154](https://github.com/dmarcguardhq/parse-dmarc/issues/154)) and whois on sending sources ([#143](https://github.com/dmarcguardhq/parse-dmarc/issues/143)). Then failure reports (RUF), CSV and JSON export, and alerting.

## License

Apache-2.0. By contributing you agree your contribution is licensed the same way.
