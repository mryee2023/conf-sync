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

### Option 1: Using Docker

#### Client Installation

1. Create your configuration:
```bash
# Create config directory
sudo mkdir -p /etc/conf-sync

# Create and edit config file
sudo tee /etc/conf-sync/client.yaml << 'EOF'
gist_id: "YOUR_GIST_ID"
check_interval: "60s"
mappings:
  - gist_file: "app.conf"
    local_path: "/etc/myapp/app.conf"
    exec: "systemctl restart myapp"
EOF
```

2. Start using docker-compose:
```bash
# Download docker-compose.yml
mkdir -p ~/conf-sync && cd ~/conf-sync
curl -O https://raw.githubusercontent.com/mryee2023/conf-sync/main/docker/client/docker-compose.yml

# Edit docker-compose.yml to mount your config directories
vi docker-compose.yml

# Start the service
docker-compose up -d
```

Or using Docker directly:
```bash
docker run -d \
  --name conf-sync-client \
  --restart unless-stopped \
  -v /etc/conf-sync/client.yaml:/etc/conf-sync/client.yaml:ro \
  -v /etc/myapp:/etc/myapp \
  mryee2023/conf-sync-client:latest
```

#### Server Installation

1. Start using docker-compose:
```bash
# Download docker-compose.yml
mkdir -p ~/conf-sync && cd ~/conf-sync
curl -O https://raw.githubusercontent.com/mryee2023/conf-sync/main/docker/server/docker-compose.yml

# Set your GitHub token
export GIST_TOKEN="your-github-token"

# Start the service
docker-compose up -d
```

Or using Docker directly:
```bash
docker run -d \
  --name conf-sync-server \
  --restart unless-stopped \
  -e GIST_TOKEN="your-github-token" \
  mryee2023/conf-sync-server:latest
```

### Option 2: Using Install Script

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

## Docker Usage

### Client Mode

The client container supports various configurations through docker-compose or command line arguments.

#### Basic Usage

```bash
# Start the client
docker run -d \
  --name conf-sync-client \
  -v /etc/conf-sync/client.yaml:/etc/conf-sync/client.yaml:ro \
  mryee2023/conf-sync-client:latest

# View logs
docker logs -f conf-sync-client

# Manual sync
docker exec conf-sync-client sync
```

#### Advanced Configuration

1. With system commands (requires privileged mode):
```yaml
# docker-compose.yml
services:
  conf-sync-client:
    image: mryee2023/conf-sync-client:latest
    volumes:
      - /etc/conf-sync/client.yaml:/etc/conf-sync/client.yaml:ro
      - /etc/myapp:/etc/myapp
    privileged: true
    restart: unless-stopped
```

2. With Docker container restart capability:
```yaml
services:
  conf-sync-client:
    image: mryee2023/conf-sync-client:latest
    volumes:
      - /etc/conf-sync/client.yaml:/etc/conf-sync/client.yaml:ro
      - /var/run/docker.sock:/var/run/docker.sock
    privileged: true
    restart: unless-stopped
```

3. Multiple configuration directories:
```yaml
services:
  conf-sync-client:
    image: mryee2023/conf-sync-client:latest
    volumes:
      - /etc/conf-sync/client.yaml:/etc/conf-sync/client.yaml:ro
      - /etc/app1:/etc/app1
      - /etc/app2:/etc/app2
      - /var/lib/app3:/var/lib/app3
    restart: unless-stopped
```

### Server Mode

The server container provides a simple interface for managing Gist files.

#### Basic Usage

```bash
# Start the server
docker run -d \
  --name conf-sync-server \
  -e GIST_TOKEN="your-github-token" \
  mryee2023/conf-sync-server:latest

# Upload files
docker exec conf-sync-server upload -g YOUR_GIST_ID /path/to/file

# List files
docker exec conf-sync-server list -g YOUR_GIST_ID

# Delete files
docker exec conf-sync-server delete -g YOUR_GIST_ID filename
```

#### With File Mounting

```yaml
# docker-compose.yml
services:
  conf-sync-server:
    image: mryee2023/conf-sync-server:latest
    environment:
      - GIST_TOKEN=${GIST_TOKEN}
    volumes:
      - ./configs:/configs
    restart: unless-stopped
```

Then use the mounted path:
```bash
docker-compose exec conf-sync-server upload -g YOUR_GIST_ID /configs/app.conf
```

### Building Docker Images

If you want to build the images yourself:

```bash
# Clone the repository
git clone https://github.com/mryee2023/conf-sync.git
cd conf-sync

# Build images
make docker-build

# Or build specific versions
make VERSION=2.0.0 docker-build

# Push to registry (requires authentication)
make docker-push
```

### Docker Security Notes

1. Client Container:
   - Runs as non-root user by default
   - Mount configuration files as read-only
   - Only use privileged mode when necessary
   - Consider using read-only root filesystem

2. Server Container:
   - Runs as non-root user
   - Protect the GIST_TOKEN environment variable
   - No need for privileged mode
   - Mount only necessary directories

3. Network Considerations:
   - Containers only need outbound access to GitHub
   - No ports need to be exposed
   - Can run in isolated networks

### Troubleshooting Docker Deployments

1. Check container status:
```bash
docker ps | grep conf-sync
```

2. View logs:
```bash
# Client logs
docker logs -f conf-sync-client

# Server logs
docker logs -f conf-sync-server
```

3. Check mounted volumes:
```bash
docker inspect conf-sync-client | grep Mounts -A 20
```

4. Common issues:
   - Configuration file not mounted correctly
   - Insufficient permissions on mounted directories
   - Missing or invalid GIST_TOKEN
   - Network connectivity issues to GitHub

## Development

### Docker Hub Release Process

The project automatically builds and publishes Docker images to Docker Hub when a new version tag is pushed. The images are available at:

- Client: [docker.io/mryee2023/conf-sync-client](https://hub.docker.com/r/mryee2023/conf-sync-client)
- Server: [docker.io/mryee2023/conf-sync-server](https://hub.docker.com/r/mryee2023/conf-sync-server)

To release a new version:

1. Tag the release:
```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

2. The GitHub Actions workflow will automatically:
   - Build multi-arch images (amd64, arm64)
   - Tag the images with version numbers (e.g., 1.0.0, 1.0, 1, latest)
   - Push the images to Docker Hub

Available image tags:
- `latest`: Latest stable release
- `x.y.z`: Specific version (e.g., `1.0.0`)
- `x.y`: Minor version (e.g., `1.0`)
- `x`: Major version (e.g., `1`)

All images are available for both amd64 and arm64 architectures.

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
