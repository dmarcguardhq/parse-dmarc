# Parse DMARC

Open-source (Apache-2.0) DMARC aggregate report parser: IMAP or Maildir intake,
SQLite, a Vue dashboard embedded in one Go binary, Prometheus metrics and an MCP
server. User docs: `README.md`, `docs/`; contributor setup: `CONTRIBUTING.md`.

## Hard rules

- The product is "Parse DMARC". Never call it DMARCguard or a self-hosted or
  community edition of it, never imply feature parity or a gated tier. The paid
  product appears only in the README's header line, comparison section and
  footer; each link carries
  `utm_source=github&utm_medium=referral&utm_campaign=parse-dmarc-readme` and a
  `utm_content` naming the slot.
- Copy states only what code proves: counts, caps and defaults come from the
  code. No "real-time", "simple", "easy", image sizes, verification stamps or
  em-dashes.
- Every asset and API URL in `src/` is relative to the page (Vite `base: "./"`,
  static files in `public/`): `e2e-nginx` serves the dashboard under `/dmarc/`
  and fails on any request outside it.
- `bun.lock` is the only lockfile; never add `package-lock.json`.
- `.github/release-notes.md` renders with `missingkey=error`: an unknown field
  fails the release. Render it after editing (the template's header comment).
- Custom text in a release PR body goes below its `## [x.y.z]` heading: above
  it, release-please publishes an empty release and `build-prod` fails.
- Nix ships the released binary through goreleaser's `nix:` pipe; never a
  source-built flake package.
- A changed flag, env var, metric or MCP tool updates `README.md` or `docs/` in
  the same PR.

## Verify

- `go test -race -count=1 ./...`, as CI runs it. The binary embeds
  `internal/api/dist`: run `just frontend` first, or stub its `index.html`.
- `bunx prettier -w .`; CI only checks.

## Process

- Conventional Commits PR title; `main` merges through the merge queue on
  `ci-gate`: squash only, signed commits.
- One canonical doc per topic (`README.md`, `docs/MCP.md`, `docs/METRICS.md`,
  `CONTRIBUTING.md`); no roadmap, plan or spec files in the tree.
- Tracker: beads (`bd`, prefix `parsedmarc-`), local only, never pushed, and
  never cited in commits, PRs or files. GitHub Issues are the public request
  intake.
