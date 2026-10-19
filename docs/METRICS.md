# Prometheus metrics and Grafana

Parse DMARC exposes Prometheus metrics at `/metrics` (on by default, `--metrics=false` turns it off) and ships a Grafana dashboard in `grafana/dashboard.json`. This page is the full catalogue plus scrape, alert and provisioning examples. The README has the summary.

## Available metrics

### Build Information

| Metric                   | Type  | Description                                     |
| ------------------------ | ----- | ----------------------------------------------- |
| `parse_dmarc_build_info` | Gauge | Build information (version, commit, build_date) |

### Report Processing

| Metric                                             | Type      | Description                                 |
| -------------------------------------------------- | --------- | ------------------------------------------- |
| `parse_dmarc_reports_fetched_total`                | Counter   | Total DMARC report emails fetched from IMAP |
| `parse_dmarc_reports_parsed_total`                 | Counter   | Total DMARC reports successfully parsed     |
| `parse_dmarc_reports_stored_total`                 | Counter   | Total DMARC reports stored in database      |
| `parse_dmarc_reports_parse_errors_total`           | Counter   | Total parse errors                          |
| `parse_dmarc_reports_store_errors_total`           | Counter   | Total storage errors                        |
| `parse_dmarc_reports_attachments_total`            | Counter   | Total attachments processed                 |
| `parse_dmarc_reports_fetch_duration_seconds`       | Histogram | Duration of fetch operations                |
| `parse_dmarc_reports_last_fetch_timestamp_seconds` | Gauge     | Unix timestamp of last successful fetch     |
| `parse_dmarc_reports_fetch_cycles_total`           | Counter   | Total fetch cycles executed                 |
| `parse_dmarc_reports_fetch_errors_total`           | Counter   | Total fetch cycle errors                    |

### IMAP Connection

| Metric                                         | Type      | Labels | Description                              |
| ---------------------------------------------- | --------- | ------ | ---------------------------------------- |
| `parse_dmarc_imap_connections_total`           | Counter   | status | IMAP connection attempts (success/error) |
| `parse_dmarc_imap_connection_duration_seconds` | Histogram |        | IMAP connection establishment duration   |

### DMARC Statistics

| Metric                                       | Type  | Description                       |
| -------------------------------------------- | ----- | --------------------------------- |
| `parse_dmarc_dmarc_reports_total`            | Gauge | Total reports in database         |
| `parse_dmarc_dmarc_messages_total`           | Gauge | Total messages across all reports |
| `parse_dmarc_dmarc_compliant_messages_total` | Gauge | Total DMARC-compliant messages    |
| `parse_dmarc_dmarc_compliance_rate`          | Gauge | Overall compliance rate (0-100)   |
| `parse_dmarc_dmarc_unique_source_ips`        | Gauge | Number of unique source IPs       |
| `parse_dmarc_dmarc_unique_domains`           | Gauge | Number of unique domains          |

### Per-Domain/Org Metrics

| Metric                                        | Type  | Labels      | Description                  |
| --------------------------------------------- | ----- | ----------- | ---------------------------- |
| `parse_dmarc_dmarc_messages_by_domain`        | Gauge | domain      | Messages per domain          |
| `parse_dmarc_dmarc_compliance_rate_by_domain` | Gauge | domain      | Compliance rate per domain   |
| `parse_dmarc_dmarc_reports_by_org`            | Gauge | org_name    | Reports per organization     |
| `parse_dmarc_dmarc_messages_by_disposition`   | Gauge | disposition | Messages by disposition type |

### Authentication Results

| Metric                           | Type  | Labels | Description                       |
| -------------------------------- | ----- | ------ | --------------------------------- |
| `parse_dmarc_dmarc_spf_results`  | Gauge | result | SPF authentication result counts  |
| `parse_dmarc_dmarc_dkim_results` | Gauge | result | DKIM authentication result counts |

### HTTP Server

| Metric                                      | Type      | Labels               | Description                |
| ------------------------------------------- | --------- | -------------------- | -------------------------- |
| `parse_dmarc_http_requests_total`           | Counter   | method, path, status | Total HTTP requests        |
| `parse_dmarc_http_request_duration_seconds` | Histogram | method, path         | HTTP request duration      |
| `parse_dmarc_http_requests_in_flight`       | Gauge     |                      | Current in-flight requests |

### Go Runtime (Built-in)

Standard Go runtime metrics are also exposed:

- `go_goroutines` - Number of goroutines
- `go_memstats_*` - Memory statistics
- `go_gc_*` - Garbage collection metrics
- `process_*` - Process metrics (CPU, memory, file descriptors)

## Disabling metrics

To disable the metrics endpoint:

```bash
# Command line
./parse-dmarc --metrics=false

# Environment variable
export PARSE_DMARC_METRICS=false

# Docker
docker run -e PARSE_DMARC_METRICS=false ghcr.io/dmarcguardhq/parse-dmarc:v1
```

## Prometheus configuration

Add Parse DMARC to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: "parse-dmarc"
    static_configs:
      - targets: ["parse-dmarc:8080"]
    scrape_interval: 30s
    metrics_path: /metrics
```

For Kubernetes with ServiceMonitor (Prometheus Operator):

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: parse-dmarc
  labels:
    app: parse-dmarc
spec:
  selector:
    matchLabels:
      app: parse-dmarc
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
```

## Grafana dashboard

A production-ready Grafana dashboard is included in `grafana/dashboard.json`.

### Import Manually

1. In Grafana, go to **Dashboards** > **Import**
2. Upload `grafana/dashboard.json` or paste its contents
3. Select your Prometheus datasource
4. Click **Import**

### Provision Automatically (Recommended for Production)

```bash
# Copy dashboard to Grafana dashboards directory
cp grafana/dashboard.json /var/lib/grafana/dashboards/parse-dmarc/

# Copy provisioning config
cp grafana/provisioning.yaml /etc/grafana/provisioning/dashboards/parse-dmarc.yaml

# Restart Grafana or wait for it to pick up changes
systemctl restart grafana-server
```

### Dashboard Variables

| Variable     | Purpose                        |
| ------------ | ------------------------------ |
| `datasource` | Prometheus datasource to query |
| `job`        | Filter by Prometheus job label |
| `instance`   | Filter by instance(s)          |
| `domain`     | Filter by monitored domain(s)  |

### Dashboard Sections

| Section                            | What It Shows                                                             |
| ---------------------------------- | ------------------------------------------------------------------------- |
| **Overview - Golden Signals**      | Compliance rate, total messages, reports count, time since last fetch     |
| **DMARC Authentication Results**   | SPF/DKIM pass rates, disposition breakdown, per-domain compliance         |
| **Report Sources & Organizations** | Top reporting organizations (Google, Microsoft, etc.), messages by domain |
| **IMAP & Fetch Operations**        | Connection health, fetch cycle monitoring, latency heatmaps               |
| **Error Tracking**                 | Parse errors, storage errors, fetch failures                              |
| **HTTP Server**                    | Request rates, latency percentiles, error rates                           |
| **Go Runtime**                     | Goroutines, memory usage, GC stats, CPU usage                             |

### Example Grafana Panels

**Compliance Rate Gauge:**

```promql
parse_dmarc_dmarc_compliance_rate
```

**Messages Over Time:**

```promql
rate(parse_dmarc_dmarc_messages_total[5m])
```

**Compliance Rate by Domain:**

```promql
parse_dmarc_dmarc_compliance_rate_by_domain
```

**SPF/DKIM Pass Rate:**

```promql
# SPF Pass Rate
parse_dmarc_dmarc_spf_results{result="pass"} / ignoring(result) sum(parse_dmarc_dmarc_spf_results) * 100

# DKIM Pass Rate
parse_dmarc_dmarc_dkim_results{result="pass"} / ignoring(result) sum(parse_dmarc_dmarc_dkim_results) * 100
```

**Fetch Success Rate:**

```promql
1 - (rate(parse_dmarc_reports_fetch_errors_total[1h]) / rate(parse_dmarc_reports_fetch_cycles_total[1h]))
```

**IMAP Connection Health:**

```promql
rate(parse_dmarc_imap_connections_total{status="success"}[5m]) /
(rate(parse_dmarc_imap_connections_total{status="success"}[5m]) + rate(parse_dmarc_imap_connections_total{status="error"}[5m]))
```

**HTTP Request Latency (p95):**

```promql
histogram_quantile(0.95, rate(parse_dmarc_http_request_duration_seconds_bucket[5m]))
```

**Reports by Organization:**

```promql
topk(10, parse_dmarc_dmarc_reports_by_org)
```

### Alerting Rules

Example Prometheus alerting rules:

```yaml
groups:
  - name: parse-dmarc
    rules:
      - alert: DMARCComplianceLow
        expr: parse_dmarc_dmarc_compliance_rate < 90
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "DMARC compliance rate is below 90%"
          description: "Current compliance rate: {{ $value }}%"

      - alert: DMARCFetchFailures
        expr: rate(parse_dmarc_reports_fetch_errors_total[15m]) > 0
        for: 30m
        labels:
          severity: critical
        annotations:
          summary: "Parse DMARC fetch failures detected"
          description: "IMAP fetch operations are failing"

      - alert: IMAPConnectionErrors
        expr: rate(parse_dmarc_imap_connections_total{status="error"}[5m]) > 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "IMAP connection errors detected"
          description: "Check IMAP credentials and server connectivity"

      - alert: NoRecentFetch
        expr: time() - parse_dmarc_reports_last_fetch_timestamp_seconds > 600
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "No recent DMARC report fetch"
          description: "Last fetch was {{ humanizeDuration $value }} ago"
```

## Docker Compose with Prometheus and Grafana

Complete monitoring stack:

```yaml
services:
  parse-dmarc:
    image: ghcr.io/dmarcguardhq/parse-dmarc:v1
    ports:
      - "8080:8080"
    volumes:
      - ./config.json:/app/config.json
      - ./data:/data

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana

volumes:
  grafana-data:
```

With `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "parse-dmarc"
    static_configs:
      - targets: ["parse-dmarc:8080"]
```

Access:

- Parse DMARC Dashboard: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
