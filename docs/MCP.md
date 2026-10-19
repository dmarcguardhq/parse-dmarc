# MCP server

Parse DMARC ships a [Model Context Protocol](https://modelcontextprotocol.io/) server, so an AI assistant can ask questions about your DMARC reports instead of you reading tables. It serves the same SQLite database as the dashboard.

## Run it

MCP mode runs only the MCP server: no IMAP fetching, no dashboard. Point it at the database a normal instance (or a `--fetch-once` cron) fills.

Over stdio, for desktop clients that spawn the process:

```bash
parse-dmarc --mcp --config config.json
```

Over Streamable HTTP, served at the root path of the `--mcp-http` address, for remote clients:

```bash
parse-dmarc --mcp-http :8081 --config config.json
```

Docker, sharing the volume with a running instance:

```bash
docker run --rm -i -v parse-dmarc:/data ghcr.io/dmarcguardhq/parse-dmarc:v1 --mcp
```

Every flag has an environment variable: `PARSE_DMARC_MCP=true`, `PARSE_DMARC_MCP_HTTP=:8081`.

## Client configuration

Claude Desktop, Cursor and other stdio clients take an entry like this:

```json
{
  "mcpServers": {
    "parse-dmarc": {
      "command": "parse-dmarc",
      "args": ["--mcp", "--config", "/path/to/config.json"]
    }
  }
}
```

## Tools

| Tool                 | What it returns                                                                                    |
| -------------------- | -------------------------------------------------------------------------------------------------- |
| `get_statistics`     | Total reports, messages, compliance rate, unique source IPs and unique domains.                    |
| `get_reports`        | Paginated report summaries: ID, organization, domain, date range, message counts, compliance rate. |
| `get_report_by_id`   | One report in full, every record and its authentication results.                                   |
| `get_top_source_ips` | Sending IPs ranked by message count, with pass and fail counts per IP.                             |
| `get_domain_stats`   | Messages, compliant messages and compliance rate per domain.                                       |
| `get_org_stats`      | Report counts per reporting organization (Google, Microsoft, Yahoo and so on).                     |
| `get_spf_stats`      | SPF result counts: pass, fail, softfail, neutral and the rest.                                     |
| `get_dkim_stats`     | DKIM result counts: pass, fail, none and the rest.                                                 |
| `parse_dmarc_report` | Parses a report you hand it as base64: gzip, zip or plain XML. Nothing is stored.                  |

## OAuth2 on the HTTP server

The HTTP transport can require a bearer token. Validation is JWT against the issuer's keys, or token introspection when an endpoint is set.

| Flag                                 | Env                                            | Meaning                                                    |
| ------------------------------------ | ---------------------------------------------- | ---------------------------------------------------------- |
| `--mcp-oauth`                        | `PARSE_DMARC_MCP_OAUTH`                        | Turn authentication on.                                    |
| `--mcp-oauth-issuer`                 | `PARSE_DMARC_MCP_OAUTH_ISSUER`                 | OIDC issuer URL, for example a Keycloak realm.             |
| `--mcp-oauth-audience`               | `PARSE_DMARC_MCP_OAUTH_AUDIENCE`               | Expected `aud` claim, usually the MCP server URL.          |
| `--mcp-oauth-client-id`              | `PARSE_DMARC_MCP_OAUTH_CLIENT_ID`              | Client ID used for introspection.                          |
| `--mcp-oauth-client-secret`          | `PARSE_DMARC_MCP_OAUTH_CLIENT_SECRET`          | Client secret used for introspection.                      |
| `--mcp-oauth-scopes`                 | `PARSE_DMARC_MCP_OAUTH_SCOPES`                 | Required scopes, comma-separated. Default `mcp:tools`.     |
| `--mcp-oauth-introspection-endpoint` | `PARSE_DMARC_MCP_OAUTH_INTROSPECTION_ENDPOINT` | If set, introspection replaces JWT validation.             |
| `--mcp-oauth-resource-name`          | `PARSE_DMARC_MCP_OAUTH_RESOURCE_NAME`          | Name shown in server metadata.                             |
| `--mcp-oauth-insecure`               | `PARSE_DMARC_MCP_OAUTH_INSECURE`               | Skip TLS verification toward the issuer. Development only. |

Example:

```bash
parse-dmarc --mcp-http :8081 --mcp-oauth \
  --mcp-oauth-issuer https://auth.example.com/realms/master \
  --mcp-oauth-audience https://dmarc.example.com/mcp
```
