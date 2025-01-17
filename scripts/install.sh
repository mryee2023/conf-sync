#!/bin/bash

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root"
  exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
    linux)
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Latest release URL
RELEASE_URL="https://github.com/mryee2023/conf-sync/releases/latest/download/conf-sync-${OS}-${ARCH}"

# Create directories
mkdir -p /etc/conf-sync
mkdir -p /usr/local/bin

# Download binary
echo "Downloading conf-sync binary..."
curl -L -o /usr/local/bin/conf-sync "$RELEASE_URL"
chmod +x /usr/local/bin/conf-sync

# Create service file
echo "Creating systemd service file..."
cat > /etc/systemd/system/conf-sync.service << 'EOL'
[Unit]
Description=Configuration Synchronization Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/conf-sync client --config /etc/conf-sync/client.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOL

# Create default config file
echo "Creating default configuration file..."
cat > /etc/conf-sync/client.yaml << 'EOL'
# Client configuration for conf-sync
gist_id: "YOUR_GIST_ID"  # Replace with your Gist ID
check_interval: "30s"
mappings:
  - gist_file: "app.conf"
    local_path: "/etc/myapp/app.conf"
    exec: "systemctl restart myapp"
EOL

# Set permissions
chmod 644 /etc/systemd/system/conf-sync.service
chmod 600 /etc/conf-sync/client.yaml

# Reload systemd
systemctl daemon-reload

echo "Installation complete!"
echo "Please edit /etc/conf-sync/client.yaml and set your Gist ID"
echo "Then run: systemctl enable conf-sync && systemctl start conf-sync"
echo ""
echo "To view service status: systemctl status conf-sync"
echo "To view logs: journalctl -u conf-sync"
