# conf-sync

A tool for synchronizing configuration files across multiple servers using GitHub Gist.

## Overview

conf-sync is designed to manage configuration files across multiple servers:
- A central server (server mode) manages and updates the configurations using GitHub Gist
- Multiple client servers (client mode) automatically sync and apply the configuration changes
- Only the server needs GitHub access; clients only need the Gist ID

```
                                      ┌─────────────┐
                                      │  GitHub     │
                                      │   Gist      │
                                      └─────▲─────┬─┘
                                            │     │
                                     Upload │     │ Read
                                     (token)│     │ (public)
                                            │     ▼
┌─────────────┐                      ┌─────┴─────────┐
│  Client B   │◄────Read Gist───────│   Server A    │
└─────────────┘      (public)        └───────────────┘
                                            ▲
                                            │
                                            │ Read Gist
                                            │ (public)
                                     ┌──────┴────────┐
                                     │   Client C    │
                                     └───────────────┘
```

### Architecture

1. **Server Node (A)**:
   - Manages configuration files
   - Requires GitHub token for Gist access
   - Uploads changes to GitHub Gist
   - Acts as the single source of truth

2. **Client Nodes (B, C)**:
   - Run conf-sync as a systemd service
   - Only need Gist ID (no GitHub token required)
   - Automatically sync from GitHub Gist
   - Execute commands after file updates (e.g., restart services)

3. **GitHub Gist**:
   - Acts as the central configuration store
   - Provides version history
   - Server writes using token
   - Clients read public Gist

## Installation

### Server Installation
```bash
# Just copy the binary
cp build/conf-sync /usr/local/bin/
chmod +x /usr/local/bin/conf-sync

# Set GitHub token
export GIST_TOKEN="your-github-token"
```

### Client Installation
```bash
# Download and run the install script
curl -L https://raw.githubusercontent.com/mryee2023/conf-sync/main/scripts/install.sh | sudo bash

# Edit the configuration
sudo vi /etc/conf-sync/client.yaml

# Start the service
sudo systemctl enable conf-sync
sudo systemctl start conf-sync
```

The install script will:
1. Detect your system architecture
2. Download the appropriate binary
3. Create systemd service
4. Create default configuration
5. Set proper permissions

## Usage

### Server Mode

The server mode is used to manage configurations:

```bash
# Upload configuration files
conf-sync -g YOUR_GIST_ID upload /path/to/config1.yaml /path/to/config2.conf

# List files in Gist
conf-sync -g YOUR_GIST_ID list

# Delete files from Gist
conf-sync -g YOUR_GIST_ID delete old-config.json
```

### Client Mode

Clients run as a service, automatically syncing configurations from the Gist:

1. Edit `/etc/conf-sync/client.yaml`:
```yaml
gist_id: "YOUR_GIST_ID"
check_interval: "30s"
mappings:
  - gist_file: "db.conf"
    local_path: "/etc/myapp/db.conf"
    exec: "docker restart myapp"
  - gist_file: "app.yaml"
    local_path: "/etc/myapp/config.yaml"
    exec: "systemctl restart myapp"
```

2. Run the client:
```bash
# Run manually
conf-sync client --config /etc/conf-sync/client.yaml

# Or use systemd
sudo systemctl start conf-sync
```

### Command Line Options

- `-g, --gist-id`: (Required for server mode) The ID of your Gist
- `-l, --log-level`: Set log level (debug, info, warn, error, default: info)
- `-c, --config`: Path to client config file (default: /etc/conf-sync/client.yaml)

### Log Levels

The following log levels are supported:
- `debug`: Show all log messages, including detailed debugging information
- `info`: Show informational messages (default)
- `warn`: Show only warning and error messages
- `error`: Show only error messages

### Client Configuration

The client configuration file (`client.yaml`) supports:

- `gist_id`: The ID of the Gist to sync from
- `check_interval`: How often to check for updates (e.g., "30s", "1m", "5m")
- `mappings`: List of file mappings
  - `gist_file`: Name of the file in Gist
  - `local_path`: Where to save the file locally
  - `exec`: (Optional) Command to run after file is updated

### Environment Variables

- `GIST_TOKEN`: GitHub personal access token with Gist access (required)

## Security

- The GitHub token should have minimal permissions (just Gist access)
- Client configuration files are created with 600 permissions
- The service runs as root to allow writing to protected directories

## Troubleshooting

1. Check the service status:
```bash
systemctl status conf-sync
```

2. View logs:
```bash
# View service logs
journalctl -u conf-sync

# Run with debug logging
conf-sync -l debug client
```

3. Common issues:
- Ensure GIST_TOKEN is set correctly
- Check file permissions
- Verify the Gist ID is correct
- Ensure the service has network access
