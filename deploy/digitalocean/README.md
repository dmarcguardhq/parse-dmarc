# Parse DMARC - DigitalOcean Droplet

This directory contains everything needed to create a DigitalOcean Marketplace image for Parse DMARC.

## Quick Deploy (Existing Image)

If a Marketplace image is available, deploy directly from the DigitalOcean Marketplace.

## Building the Image

### Prerequisites

1. [Packer](https://www.packer.io/) v1.8+
2. A DigitalOcean account with API token
3. At least one SSH key in your DigitalOcean account

### Build Steps

```bash
# Set your DigitalOcean API token
export DIGITALOCEAN_TOKEN="your-api-token"

# Initialize Packer plugins
packer init packer.pkr.hcl

# Validate the template
packer validate packer.pkr.hcl

# Build the image
packer build packer.pkr.hcl
```

### Build Options

```bash
# Build with a specific version
packer build -var 'parse_dmarc_version=v1.2.0' packer.pkr.hcl

# Build with custom image name
packer build -var 'image_name=parse-dmarc-custom' packer.pkr.hcl

# Build in a specific region
packer build -var 'droplet_region=fra1' packer.pkr.hcl
```

## What's Included

The image includes:

- **Parse DMARC** - Latest version from GitHub releases
- **Ubuntu 24.04 LTS** - Base operating system
- **UFW Firewall** - Pre-configured (SSH + port 8080)
- **Fail2ban** - SSH brute-force protection
- **Automatic Security Updates** - Unattended upgrades enabled
- **Systemd Service** - Parse DMARC runs as a service

## Directory Structure

```
deploy/digitalocean/
├── README.md           # This file
├── packer.pkr.hcl      # Packer template
├── marketplace.yaml    # Marketplace submission metadata
└── scripts/
    ├── 01-system-setup.sh        # System packages and security
    ├── 02-install-parse-dmarc.sh # Parse DMARC installation
    ├── 03-configure-services.sh  # Systemd and first-boot setup
    ├── 90-cleanup.sh             # Image cleanup
    └── 99-img-check.sh           # Marketplace validation
```

## First Boot Configuration

When a user creates a Droplet from this image:

1. They'll see a welcome MOTD with setup instructions
2. Configuration file is at `/etc/parse-dmarc/config.json`
3. After editing the config, restart the service:
   ```bash
   sudo systemctl restart parse-dmarc
   ```
4. Dashboard is available at `http://<droplet-ip>:8080`

## Settings

The service reads `/etc/parse-dmarc/config.json` (step 2 above); systemd does not pass
`/etc/environment` to it. `parse-dmarc --gen-config -c /tmp/config.json` writes
a file with every key; the [README](../../README.md#configuration) says what
each setting does.

## Service Management

```bash
# Check status
sudo systemctl status parse-dmarc

# View logs
sudo journalctl -u parse-dmarc -f

# Restart service
sudo systemctl restart parse-dmarc

# Stop service
sudo systemctl stop parse-dmarc
```

## Security Notes

- The image follows DigitalOcean Marketplace security guidelines
- UFW firewall is enabled by default
- Fail2ban protects against SSH brute-force attacks
- Automatic security updates are enabled
- Config file permissions are restricted (600)

## Marketplace Submission

To submit to DigitalOcean Marketplace:

1. Build the image using Packer
2. Note the snapshot ID from the build output
3. Submit via the [Vendor Portal](https://marketplace.digitalocean.com/vendors)
4. Include `marketplace.yaml` metadata

## Troubleshooting

### Service won't start

Check if configuration is valid:
```bash
cat /etc/parse-dmarc/config.json
sudo journalctl -u parse-dmarc -n 50
```

### Can't connect to dashboard

Check firewall:
```bash
sudo ufw status
sudo ufw allow 8080/tcp
```

### IMAP connection errors

- Verify IMAP credentials
- For Gmail, use an App Password
- Check if port 993 is not blocked by your network
