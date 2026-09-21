# Parse DMARC

[![CI](https://img.shields.io/github/actions/workflow/status/dmarcguardhq/parse-dmarc/ci.yml?branch=main&label=CI&style=flat-square)](https://github.com/dmarcguardhq/parse-dmarc/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/dmarcguardhq/parse-dmarc?style=flat-square)](https://github.com/dmarcguardhq/parse-dmarc/releases)
[![License](https://img.shields.io/github/license/dmarcguardhq/parse-dmarc?style=flat-square)](LICENSE)
[![Stars](https://img.shields.io/github/stars/dmarcguardhq/parse-dmarc?style=flat-square)](https://github.com/dmarcguardhq/parse-dmarc)
[![Docker pulls old](https://img.shields.io/docker/pulls/meysam81/parse-dmarc?style=flat-square&label=meysam81%2Fparse-dmarc%20%28old%29)](https://hub.docker.com/r/meysam81/parse-dmarc)
[![Docker pulls new](https://img.shields.io/docker/pulls/dmarcguard/parse-dmarc?style=flat-square&label=dmarcguard%2Fparse-dmarc%20%28NEW%29)](https://hub.docker.com/r/dmarcguard/parse-dmarc)
[![Image size](https://img.shields.io/docker/image-size/dmarcguard/parse-dmarc/latest?style=flat-square&label=image)](https://hub.docker.com/r/dmarcguard/parse-dmarc/tags)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [What it does](#what-it-does)
- [Run it](#run-it)
- [Get reports flowing](#get-reports-flowing)
- [Deploy anywhere](#deploy-anywhere)
  - [Platform as a Service (PaaS)](#platform-as-a-service-paas)
  - [Self-hosted PaaS](#self-hosted-paas)
  - [Infrastructure](#infrastructure)
- [Configuration](#configuration)
- [Metrics, Grafana and MCP](#metrics-grafana-and-mcp)
- [Parse DMARC or DMARCguard?](#parse-dmarc-or-dmarcguard)
- [Roadmap and contributing](#roadmap-and-contributing)
- [License](#license)

<!-- END doctoc -->

**Read your DMARC aggregate reports in one dashboard. One Go binary, SQLite, no Elasticsearch.**

Parse DMARC is built and maintained by the team behind [DMARCguard](https://dmarcguard.io/?utm_source=github&utm_medium=referral&utm_campaign=parse-dmarc-readme&utm_content=header), a hosted DMARC monitoring service. The two do not share code, and this repository stays Apache-2.0 in full.

![Parse DMARC dashboard showing pass rate, message volume and top sending sources](./assets/screenshots/hero.png)

## What it does

When your DMARC record carries a `rua=` address, mail receivers such as Google, Microsoft and Yahoo send you aggregate reports: gzip or zip XML attachments, at the interval your record asks for and daily by default, listing every IP that sent mail as your domain and whether SPF and DKIM passed ([RFC 7489 section 7.2](https://www.rfc-editor.org/rfc/rfc7489#section-7.2), carried forward by [RFC 9990](https://www.rfc-editor.org/rfc/rfc9990)). Nobody reads those by hand. Parse DMARC does:

- Fetches reports over IMAP from any mailbox. Reports that Exchange or Outlook forward as `message/rfc822` attachments are unwrapped too.
- Parses gzip, zip and raw XML, with a 16 MB decompression cap per report.
- Stores everything in one SQLite file. No database server, no JVM.
- Shows pass rate, message volume and top sending sources, and opens any report down to its raw records.
- Generates your `_dmarc` TXT record from a form.
- Exposes 28 Prometheus metrics and ships a Grafana dashboard.
- Serves an MCP server, so an AI assistant can query your reports.
- Dark mode. One static binary. The Docker image is built `FROM scratch`.

It reads aggregate (RUA) reports only. Failure reports (RUF), TLS-RPT, alerting, login and multi-mailbox intake are not built; see [Roadmap and contributing](#roadmap-and-contributing).

## Run it

Docker, with a named volume for the database:

```bash
docker run -d --name parse-dmarc -p 8080:8080 \
  -e IMAP_HOST=imap.gmail.com \
  -e IMAP_PORT=993 \
  -e IMAP_USERNAME=dmarc@example.com \
  -e IMAP_PASSWORD='your-app-password' \
  -v parse-dmarc:/data \
  dmarcguard/parse-dmarc
```

Homebrew on macOS or Linux:

```bash
brew install dmarcguardhq/tap/parse-dmarc
parse-dmarc --gen-config          # writes a config.json template
parse-dmarc --config config.json
```

Nix, from our [NUR-style repo](https://github.com/dmarcguardhq/nur):

```bash
nix profile install github:dmarcguardhq/nur#parse-dmarc   # or: nix run github:dmarcguardhq/nur#parse-dmarc
```

Or download a release archive, `parse-dmarc_<os>_<arch>.tar.gz`, from the [releases page](https://github.com/dmarcguardhq/parse-dmarc/releases).

Open http://localhost:8080. Gmail needs an [App Password](https://support.google.com/accounts/answer/185833), not the account password. The same image is on Docker Hub as `meysam81/parse-dmarc`; both names track the same builds, and tags such as `v1` or `v1.6.0` pin a release.

## Get reports flowing

Receivers only send reports if your DMARC record asks for them. Publish this TXT record at `_dmarc.example.com`, with your own domain and mailbox:

```
v=DMARC1; p=none; rua=mailto:dmarc@example.com
```

- `p=none` asks receivers to deliver as usual and only report. Move to `p=quarantine`, then `p=reject`, once every legitimate sender in the reports passes ([RFC 9989 section 5.1](https://www.rfc-editor.org/rfc/rfc9989#section-5.1)).
- `rua=` is the mailbox Parse DMARC reads. It has to exist and accept mail before the first report arrives.
- Check the record with `dig +short TXT _dmarc.example.com`. Reports typically start within 24 to 48 hours.

Cloudflare, Route 53 and every other DNS host take the same three fields: name `_dmarc`, type `TXT`, value as above. SPF and DKIM do not have to be set up first; the reports are how you find out what they are doing.

## Deploy anywhere

One click on a platform, or a template for the self-hosted PaaS you already run. Every option needs the IMAP settings from [Configuration](#configuration).

### Platform as a Service (PaaS)

| Provider       | Deploy                                                                                                                                                                             | Notes                                                     |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| **Railway**    | [![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/deploy/4kqQ_I?referralCode=meysam)                                                                      | Recommended for beginners                                 |
| **Render**     | [![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/meysam81/parse-dmarc)                               | Free tier available                                       |
| **Koyeb**      | [![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)][koyeb-1click]                                                                                          | Global edge deployment. Manually mount `/data` as volume. |
| **Zeabur**     | [![Deploy on Zeabur](https://zeabur.com/button.svg)](https://zeabur.com/templates/YB3TS7?referralCode=meysam)                                                                      | Asia-Pacific optimized                                    |
| **Northflank** | [![Deploy to Northflank](https://assets.northflank.com/deploy_to_northflank_smm_36700fb050.svg)](https://app.northflank.com/s/account/templates/new?data=693e394eb41e1e64db65187e) | Developer-focused                                         |

### Self-hosted PaaS

| Provider     | Deploy                                                                                                                                              | Notes                           |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| **CapRover** | [![Deploy to CapRover](https://img.shields.io/badge/Deploy-CapRover-0072CE?style=for-the-badge&logo=docker)](./deploy/captain-definition)           | Self-hosted PaaS                |
| **Coolify**  | [![Deploy to Coolify](https://img.shields.io/badge/Deploy-Coolify-6B46C1?style=for-the-badge&logo=docker)](./deploy/coolify.yaml)                   | Open-source Heroku alternative  |
| **Dokploy**  | [![Deploy to Dokploy](https://img.shields.io/badge/Deploy-Dokploy-00B4D8?style=for-the-badge&logo=docker)](./deploy/dokploy/)                       | Self-hosted deployment platform |
| **Docker**   | [![Docker](https://img.shields.io/badge/Docker-Pull%20Image-2496ED?style=for-the-badge&logo=docker)](https://hub.docker.com/r/meysam81/parse-dmarc) | Run anywhere                    |

### Infrastructure

| Provider                 | Deploy                                                                                                                                             | Notes                |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------- |
| **DigitalOcean Droplet** | [![Deploy to DigitalOcean](https://img.shields.io/badge/Deploy-DigitalOcean-0080FF?style=for-the-badge&logo=digitalocean)](./deploy/digitalocean/) | VM with Packer image |

## Configuration

Every setting is an environment variable or a key in `config.json`. Environment is read first and the file overrides it.

| Setting                       | Environment variable             | Default                                                      |
| ----------------------------- | -------------------------------- | ------------------------------------------------------------ |
| IMAP host and port            | `IMAP_HOST`, `IMAP_PORT`         | required, `993`                                              |
| IMAP username and password    | `IMAP_USERNAME`, `IMAP_PASSWORD` | required                                                     |
| Mailbox to read               | `IMAP_MAILBOX`                   | `INBOX`                                                      |
| Implicit TLS                  | `IMAP_USE_TLS`                   | `true`                                                       |
| STARTTLS on a plaintext port  | `IMAP_STARTTLS`                  | `false`                                                      |
| Internal CA bundle, PEM       | `IMAP_TLS_CA_FILE`               | unset                                                        |
| Skip certificate verification | `IMAP_TLS_SKIP_VERIFY`           | `false`                                                      |
| Mark fetched mail as seen     | `IMAP_MARK_AS_SEEN`              | `true`                                                       |
| Move processed mail to        | `IMAP_PROCESSED_MAILBOX`         | unset                                                        |
| Database file                 | `DATABASE_PATH`                  | `~/.parse-dmarc/db.sqlite`, `/data/parse-dmarc.db` in Docker |
| HTTP listen                   | `SERVER_HOST`, `SERVER_PORT`     | all interfaces, `8080`                                       |
| Seconds between fetches       | `FETCH_INTERVAL`                 | `300`                                                        |
| Log level                     | `LOG_LEVEL`                      | `info`                                                       |

Providers: Gmail is `imap.gmail.com` on 993 with an App Password. Microsoft 365 is `outlook.office365.com` on 993. Anything else is port 993 with TLS unless its documentation says otherwise.

An IMAP server whose certificate comes from your own CA fails with `x509: certificate signed by unknown authority`. Mount the CA bundle and point `IMAP_TLS_CA_FILE` at it; verification stays on. `IMAP_TLS_SKIP_VERIFY=true` also connects and is the last resort, because the session can then be intercepted. `IMAP_STARTTLS=true` dials plaintext, usually port 143, and upgrades; it takes precedence over `IMAP_USE_TLS`. Turning TLS off entirely is plaintext IMAP, and most servers refuse `LOGIN` on it.

| Flag                        | What it does                                                                                  |
| --------------------------- | --------------------------------------------------------------------------------------------- |
| `--fetch-once`              | Fetch, parse, store, exit. For cron.                                                          |
| `--serve-only`              | Serve the dashboard without fetching.                                                         |
| `--fetch-interval 600`      | Seconds between fetch cycles.                                                                 |
| `--metrics=false`           | Turn off `/metrics`.                                                                          |
| `--gen-config`              | Write a `config.json` template and exit.                                                      |
| `--mcp`, `--mcp-http :8081` | Run the MCP server over stdio or HTTP instead of the fetcher. See [docs/MCP.md](docs/MCP.md). |

## Metrics, Grafana and MCP

`/metrics` is on by default: 28 metrics covering fetch cycles, IMAP connections, parse and store errors, compliance rate overall and per domain, SPF and DKIM result counts, and HTTP latency. `grafana/dashboard.json` is a ready dashboard for them. The full list, a Prometheus Operator `ServiceMonitor`, alert rules and a compose stack with Prometheus and Grafana are in [docs/METRICS.md](docs/METRICS.md).

`parse-dmarc --mcp` exposes the same data to an AI assistant over the Model Context Protocol: nine tools, from `get_statistics` to `parse_dmarc_report`, over stdio or HTTP with optional OAuth2. Setup and the tool list are in [docs/MCP.md](docs/MCP.md).

The HTTP API behind the dashboard is four `GET` routes: `/api/statistics`, `/api/reports`, `/api/reports/{id}` and `/api/top-sources`. None of them has authentication, so keep the port behind your reverse proxy or VPN.

## Parse DMARC or DMARCguard?

Two products, one maker, separate codebases. Parse DMARC is the whole of this repository: no enterprise directory, nothing to unlock. DMARCguard is a hosted, proprietary service that reads the reports for you and tells you what to change.

|                 | Parse DMARC                                     | DMARCguard                                                                                                                                                             |
| --------------- | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Who runs it     | You: a binary or container, a mailbox, a volume | We do                                                                                                                                                                  |
| What it reads   | DMARC aggregate reports                         | DMARC aggregate reports, TLS-RPT reports, failure reports on Pro, and DNS checks for 9 protocols: DMARC, SPF, DKIM, BIMI, MTA-STS, TLS-RPT, ARC, DANE, ARF (7 on Free) |
| Sending sources | IP addresses                                    | 170 named senders, Mailchimp or SendGrid rather than an IP, each with the DNS change that fixes it                                                                     |
| Alerts          | Prometheus rules you write                      | Email on Free; Slack, Teams, Discord and webhooks on Pro                                                                                                               |
| Retention       | Whatever your disk holds                        | 30 days on Free, 1 year on Pro                                                                                                                                         |
| Access control  | None; keep it behind your proxy                 | Accounts and 2FA; SAML SSO on Pro                                                                                                                                      |
| Support         | GitHub issues                                   | Email                                                                                                                                                                  |
| Price           | $0, Apache-2.0                                  | Free for 2 domains, no credit card. Pro is $29/month including 2 domains                                                                                               |

Run Parse DMARC when you want the data on your own disk and one mailbox to watch. Use DMARCguard when you want the reports read for you, or you have more domains than evenings.

> I built Parse DMARC first, in late 2025. DMARCguard came out of what people asked for next: a sender's name instead of its IP, and the exact record to change on every alert. Paying for DMARCguard is what funds the hours that go into this repository, and it stays Apache-2.0.
>
> Meysam

[Start free on 2 domains, no credit card](https://dmarcguard.io/pricing/?utm_source=github&utm_medium=referral&utm_campaign=parse-dmarc-readme&utm_content=which-one)

## Roadmap and contributing

The three most-asked additions, in the order people ask: TLS-RPT reports ([#154](https://github.com/dmarcguardhq/parse-dmarc/issues/154)), a Maildir or directory intake for people without IMAP ([#169](https://github.com/dmarcguardhq/parse-dmarc/issues/169)), and whois on sending sources ([#143](https://github.com/dmarcguardhq/parse-dmarc/issues/143)). [ROADMAP.md](ROADMAP.md) has the rest. [CONTRIBUTING.md](CONTRIBUTING.md) covers the toolchain: `just build`, `just dev`, and `docker compose up` for a local Dovecot seeded with a sample report.

## License

Apache-2.0, for everything in this repository. There is no `ee/` directory and no feature that needs a key.

---

If Parse DMARC did the job, a star helps the next person find it. If you would rather have the reports read for you, [DMARCguard](https://dmarcguard.io/?utm_source=github&utm_medium=referral&utm_campaign=parse-dmarc-readme&utm_content=footer) is the hosted product from the same team.

[koyeb-1click]: https://app.koyeb.com/deploy?name=parse-dmarc&type=docker&image=docker.io%2Fmeysam81%2Fparse-dmarc%3Alatest&regions=fra&env%5BDATABASE_PATH%5D=%2Fdata%2Fdb.sqlite&env%5BIMAP_HOST%5D=&env%5BIMAP_MAILBOX%5D=INBOX&env%5BIMAP_PASSWORD%5D=&env%5BIMAP_PORT%5D=993&env%5BIMAP_USERNAME%5D=&env%5BIMAP_USE_TLS%5D=true&env%5BSERVER_PORT%5D=8080&ports=8080%3Bhttp%3B%2F&hc_protocol%5B8080%5D=http&hc_grace_period%5B8080%5D=5&hc_interval%5B8080%5D=30&hc_restart_limit%5B8080%5D=3&hc_timeout%5B8080%5D=5&hc_path%5B8080%5D=%2Fapi%2Fstatistics&hc_method%5B8080%5D=get
