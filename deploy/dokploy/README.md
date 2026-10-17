# Parse DMARC - Dokploy Deployment

Deploy Parse DMARC to your Dokploy instance.

## Quick Deploy

### Option 1: Using Template (Recommended)

1. Open your Dokploy dashboard
2. Go to **Templates** or **Projects** → **Create from Template**
3. Search for "Parse DMARC" or use this repository
4. Configure the required environment variables
5. Deploy

### Option 2: Manual Docker Compose

1. In Dokploy, create a new **Docker Compose** project
2. Copy the contents of `docker-compose.yml` to the compose editor
3. Configure environment variables in the Environment section
4. Deploy

## Required Configuration

Before the application can start, you must configure these environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `IMAP_HOST` | Your IMAP server hostname | `imap.gmail.com` |
| `IMAP_USERNAME` | Your email address | `dmarc@yourdomain.com` |
| `IMAP_PASSWORD` | Your email password or app password | `your-app-password` |

## Optional Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `IMAP_PORT` | IMAP server port | `993` |
| `IMAP_MAILBOX` | Mailbox to fetch reports from | `INBOX` |
| `IMAP_USE_TLS` | Use TLS for IMAP connection | `true` |
| `FETCH_INTERVAL` | Seconds between fetch cycles | `300` |
| `METRICS_ENABLED` | Enable Prometheus metrics | `true` |

## Email Provider Setup

### Gmail

1. Enable 2-Step Verification in your Google account
2. Generate an App Password: Google Account → Security → App Passwords
3. Use these settings:
   - `IMAP_HOST`: `imap.gmail.com`
   - `IMAP_PORT`: `993`
   - `IMAP_PASSWORD`: Your generated app password

### Outlook/Office 365

- `IMAP_HOST`: `outlook.office365.com`
- `IMAP_PORT`: `993`

### Custom IMAP Server

Most providers use port `993` with TLS. Check your provider's documentation.

## Data Persistence

The template automatically creates a volume `parse-dmarc-data` mounted at `/data` inside the container. This stores:

- SQLite database with all parsed DMARC reports

## Health Checks

The application includes a health check that monitors `/api/statistics`:
- Interval: 30 seconds
- Timeout: 10 seconds
- Retries: 3

## Accessing the Dashboard

After deployment, access your Parse DMARC dashboard at the domain you configured in Dokploy.

## Prometheus Metrics

Metrics are available at `/metrics` when `METRICS_ENABLED=true`.

## Troubleshooting

### Application won't start

Check the logs for configuration errors. The three required variables must be set:
- `IMAP_HOST`
- `IMAP_USERNAME`
- `IMAP_PASSWORD`

### IMAP connection errors

1. Verify your credentials are correct
2. For Gmail, ensure you're using an App Password
3. Check if your IMAP server requires specific settings

### No reports appearing

1. Verify your DMARC DNS record is set up correctly
2. Wait 24-48 hours for reports to arrive
3. Check if reports are in the configured mailbox

## Links

- [Parse DMARC GitHub](https://github.com/dmarcguardhq/parse-dmarc)
- [Dokploy Documentation](https://docs.dokploy.com/)
