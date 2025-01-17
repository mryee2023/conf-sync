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
   - Support both automatic and manual sync

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

## Configuration

### Client Configuration
```yaml
# Client configuration for conf-sync
gist_id: "YOUR_GIST_ID"  # Replace with your Gist ID
check_interval: "60s"    # Minimum interval is 60s for unauthenticated requests
mappings:
  - gist_file: "app.conf"    # File name in Gist
    local_path: "/etc/myapp/app.conf"  # Where to save locally
    exec: "systemctl restart myapp"     # Optional command to run after update
```

Note: Due to GitHub API rate limits:
- Authenticated requests (server): minimum interval 1s (5000 requests/hour)
- Unauthenticated requests (client): minimum interval 60s (60 requests/hour)
- The client will automatically adjust intervals if rate limited

## Usage

### Client Mode

1. Start the service for automatic sync:
```bash
sudo systemctl start conf-sync
```

2. View service logs:
```bash
sudo journalctl -u conf-sync -f
```

3. Manual sync:
```bash
# Sync all files once
conf-sync sync

# Use a different config file
conf-sync sync -c /path/to/config.yaml
```

### Server Mode

1. Set GitHub token:
```bash
export GIST_TOKEN="your-github-token"
```

2. Upload files:
```bash
# Upload single file
conf-sync -g YOUR_GIST_ID upload config.yaml

# Upload multiple files
conf-sync -g YOUR_GIST_ID upload config1.yaml config2.conf
```

3. List files:
```bash
conf-sync -g YOUR_GIST_ID list
```

4. Delete files:
```bash
conf-sync -g YOUR_GIST_ID delete config.yaml
```

### Command Line Options

```bash
Usage:
  conf-sync [command]

Available Commands:
  client      Run in client mode (automatic sync)
  sync        Sync files once in client mode
  upload      Upload files to Gist (server mode)
  list        List files in Gist (server mode)
  delete      Delete files from Gist (server mode)

Flags:
  -g, --gist-id string    Gist ID
  -l, --log-level string  Log level (debug, info, warn, error) (default "info")
  -h, --help             Help for conf-sync

Client Mode Flags:
  -c, --config string    Path to client config file (default "/etc/conf-sync/client.yaml")
```

## Security Notes

1. Server Security:
   - Keep your GitHub token secure
   - Use environment variables for tokens
   - Consider using read-only tokens when possible

2. Client Security:
   - No GitHub token required
   - Configuration file permissions are set to 600
   - Only root can modify the configuration

3. Network Security:
   - All communication is through HTTPS
   - Clients only need outbound access to GitHub
   - No direct communication between server and clients

## Troubleshooting

1. Rate Limiting:
   - If you see "rate limit exceeded" errors, the client will automatically increase the check interval
   - The interval will gradually return to normal when rate limits allow
   - For clients, ensure check_interval is at least 60s

2. File Updates:
   - Check service logs with `journalctl -u conf-sync`
   - Use `conf-sync sync` to force an immediate update
   - Verify file permissions and paths

3. Command Execution:
   - Commands run as root
   - Use absolute paths in commands
   - Check logs for command output

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
