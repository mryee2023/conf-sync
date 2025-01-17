#!/bin/bash

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root"
  exit 1
fi

# Create directories
mkdir -p /etc/conf-sync
mkdir -p /usr/local/bin

# Copy binary
cp conf-sync /usr/local/bin/
chmod +x /usr/local/bin/conf-sync

# Copy config
cp conf/client.yaml /etc/conf-sync/
chmod 600 /etc/conf-sync/client.yaml

# Install service
cp scripts/conf-sync.service /etc/systemd/system/
chmod 644 /etc/systemd/system/conf-sync.service

# Reload systemd
systemctl daemon-reload

echo "Installation complete!"
echo "Please edit /etc/conf-sync/client.yaml and set your GitHub token in the service file"
echo "Then run: systemctl enable conf-sync && systemctl start conf-sync"
