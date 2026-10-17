#!/bin/bash
# Parse DMARC - DigitalOcean Droplet Setup
# Script 03: Configure systemd services and first-boot
set -euo pipefail

echo "==> Creating systemd service..."
cat > /etc/systemd/system/parse-dmarc.service <<'EOF'
[Unit]
Description=Parse DMARC - DMARC Report Analyzer
Documentation=https://github.com/dmarcguardhq/parse-dmarc
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=parse-dmarc
Group=parse-dmarc
ExecStart=/opt/parse-dmarc/parse-dmarc --config=/etc/parse-dmarc/config.json
Restart=always
RestartSec=10
WorkingDirectory=/var/lib/parse-dmarc

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/var/lib/parse-dmarc

# Environment
Environment=DATABASE_PATH=/var/lib/parse-dmarc/db.sqlite

[Install]
WantedBy=multi-user.target
EOF

echo "==> Creating first-boot configuration service..."
cat > /etc/systemd/system/parse-dmarc-setup.service <<'EOF'
[Unit]
Description=Parse DMARC First Boot Setup
ConditionPathExists=!/var/lib/parse-dmarc/.configured
Before=parse-dmarc.service

[Service]
Type=oneshot
ExecStart=/opt/parse-dmarc/first-boot.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF

echo "==> Creating first-boot script..."
cat > /opt/parse-dmarc/first-boot.sh <<'SCRIPT'
#!/bin/bash
# Parse DMARC - First Boot Configuration Script
set -euo pipefail

CONFIG_FILE="/etc/parse-dmarc/config.json"
MARKER_FILE="/var/lib/parse-dmarc/.configured"

# Check if already configured
if [ -f "$MARKER_FILE" ]; then
    echo "Parse DMARC is already configured."
    exit 0
fi

# Function to check if config is valid
check_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        return 1
    fi

    # Check required fields
    local host=$(jq -r '.imap.host // empty' "$CONFIG_FILE" 2>/dev/null)
    local username=$(jq -r '.imap.username // empty' "$CONFIG_FILE" 2>/dev/null)
    local password=$(jq -r '.imap.password // empty' "$CONFIG_FILE" 2>/dev/null)

    if [ -z "$host" ] || [ -z "$username" ] || [ -z "$password" ]; then
        return 1
    fi

    return 0
}

# Check environment variables for configuration
configure_from_env() {
    if [ -n "${IMAP_HOST:-}" ] && [ -n "${IMAP_USERNAME:-}" ] && [ -n "${IMAP_PASSWORD:-}" ]; then
        echo "Configuring from environment variables..."

        cat > "$CONFIG_FILE" <<EOF
{
  "imap": {
    "host": "${IMAP_HOST}",
    "port": ${IMAP_PORT:-993},
    "username": "${IMAP_USERNAME}",
    "password": "${IMAP_PASSWORD}",
    "mailbox": "${IMAP_MAILBOX:-INBOX}",
    "use_tls": ${IMAP_USE_TLS:-true}
  },
  "database": {
    "path": "/var/lib/parse-dmarc/db.sqlite"
  },
  "server": {
    "port": ${SERVER_PORT:-8080},
    "host": "${SERVER_HOST:-0.0.0.0}"
  }
}
EOF
        chown parse-dmarc:parse-dmarc "$CONFIG_FILE"
        chmod 600 "$CONFIG_FILE"
        return 0
    fi
    return 1
}

# Main logic
echo "========================================"
echo "  Parse DMARC - First Boot Setup"
echo "========================================"

# Try to configure from environment
if configure_from_env; then
    echo "Configuration created from environment variables."
elif check_config; then
    echo "Configuration file already exists and is valid."
else
    # Create a sample config file for the user
    cat > "$CONFIG_FILE" <<EOF
{
  "imap": {
    "host": "CHANGE_ME",
    "port": 993,
    "username": "CHANGE_ME",
    "password": "CHANGE_ME",
    "mailbox": "INBOX",
    "use_tls": true
  },
  "database": {
    "path": "/var/lib/parse-dmarc/db.sqlite"
  },
  "server": {
    "port": 8080,
    "host": "0.0.0.0"
  }
}
EOF
    chown parse-dmarc:parse-dmarc "$CONFIG_FILE"
    chmod 600 "$CONFIG_FILE"

    echo ""
    echo "IMPORTANT: Parse DMARC requires configuration before use."
    echo ""
    echo "Please edit: $CONFIG_FILE"
    echo ""
    echo "Required settings:"
    echo "  - imap.host     : Your IMAP server (e.g., imap.gmail.com)"
    echo "  - imap.username : Your email address"
    echo "  - imap.password : Your email password or app password"
    echo ""
    echo "After configuring, run:"
    echo "  sudo systemctl restart parse-dmarc"
    echo ""

    # Don't mark as configured - user needs to edit the file
    exit 0
fi

# Mark as configured
touch "$MARKER_FILE"
echo "Parse DMARC is configured and ready."
SCRIPT

chmod +x /opt/parse-dmarc/first-boot.sh

echo "==> Creating MOTD..."
cat > /etc/update-motd.d/99-parse-dmarc <<'EOF'
#!/bin/bash

echo ""
echo "  ____                       ____  __  __    _    ____   ____"
echo " |  _ \ __ _ _ __ ___  ___  |  _ \|  \/  |  / \  |  _ \ / ___|"
echo " | |_) / _\` | '__/ __|/ _ \ | | | | |\/| | / _ \ | |_) | |    "
echo " |  __/ (_| | |  \__ \  __/ | |_| | |  | |/ ___ \|  _ <| |___ "
echo " |_|   \__,_|_|  |___/\___| |____/|_|  |_/_/   \_\_| \_\\____|"
echo ""
echo " DMARC Report Analyzer"
echo " https://github.com/dmarcguardhq/parse-dmarc"
echo ""
echo "=========================================================================="
echo ""

CONFIG_FILE="/etc/parse-dmarc/config.json"

if [ -f "$CONFIG_FILE" ]; then
    host=$(jq -r '.imap.host // empty' "$CONFIG_FILE" 2>/dev/null)
    if [ "$host" = "CHANGE_ME" ] || [ -z "$host" ]; then
        echo " STATUS: Not Configured"
        echo ""
        echo " To configure Parse DMARC, edit:"
        echo "   sudo nano /etc/parse-dmarc/config.json"
        echo ""
        echo " Then restart the service:"
        echo "   sudo systemctl restart parse-dmarc"
    else
        echo " STATUS: Configured"
        echo " IMAP Host: $host"
        echo ""
        echo " Service: $(systemctl is-active parse-dmarc 2>/dev/null || echo 'unknown')"
        echo " Dashboard: http://$(hostname -I | awk '{print $1}'):8080"
    fi
else
    echo " STATUS: Configuration file missing"
    echo ""
    echo " Run first-boot setup:"
    echo "   sudo /opt/parse-dmarc/first-boot.sh"
fi

echo ""
echo "=========================================================================="
echo ""
EOF

chmod +x /etc/update-motd.d/99-parse-dmarc

echo "==> Enabling services..."
systemctl daemon-reload
systemctl enable parse-dmarc-setup.service
systemctl enable parse-dmarc.service

echo "==> Services configured successfully"
