#!/bin/bash
# Parse DMARC - DigitalOcean Droplet Setup
# Script 02: Install Parse DMARC
set -euo pipefail

PARSE_DMARC_VERSION="${PARSE_DMARC_VERSION:-latest}"

echo "==> Installing Parse DMARC (version: ${PARSE_DMARC_VERSION})..."

# Determine the download URL
if [ "$PARSE_DMARC_VERSION" = "latest" ]; then
    DOWNLOAD_URL=$(curl -s https://api.github.com/repos/dmarcguardhq/parse-dmarc/releases/latest | \
        jq -r '.assets[] | select(.name | contains("linux_amd64")) | .browser_download_url')
else
    DOWNLOAD_URL="https://github.com/dmarcguardhq/parse-dmarc/releases/download/${PARSE_DMARC_VERSION}/parse-dmarc_${PARSE_DMARC_VERSION#v}_linux_amd64.tar.gz"
fi

echo "==> Downloading from: ${DOWNLOAD_URL}"
cd /tmp
TARBALL="$(basename "$DOWNLOAD_URL")"
curl -sLo "$TARBALL" "$DOWNLOAD_URL"

# Extract and install
tar -xzf "$TARBALL"
mv parse-dmarc /opt/parse-dmarc/parse-dmarc
chmod +x /opt/parse-dmarc/parse-dmarc
rm -f "$TARBALL"

# Create symlink
ln -sf /opt/parse-dmarc/parse-dmarc /usr/local/bin/parse-dmarc

# Verify installation
/opt/parse-dmarc/parse-dmarc version

echo "==> Parse DMARC installed successfully"
