# Docker Support for conf-sync

This directory contains Docker configurations for both client and server components of conf-sync.

## Client Usage

1. Create your configuration file:
```bash
# Copy and edit the example config
cp conf/client.yaml /etc/conf-sync/client.yaml
vi /etc/conf-sync/client.yaml
```

2. Start the client:
```bash
cd docker/client
docker-compose up -d
```

3. View logs:
```bash
docker-compose logs -f
```

### Client Configuration

The client container needs:
1. Configuration file mounted at `/etc/conf-sync/client.yaml`
2. Directories for your application configs mounted at appropriate locations
3. Privileged mode if you need to run commands after file updates

Example docker-compose override:
```yaml
services:
  conf-sync-client:
    volumes:
      - /etc/conf-sync/client.yaml:/etc/conf-sync/client.yaml:ro
      - /etc/myapp:/etc/myapp
      - /var/run/docker.sock:/var/run/docker.sock  # If you need to restart Docker containers
    privileged: true  # If you need to run system commands
```

## Server Usage

1. Set your GitHub token:
```bash
export GIST_TOKEN="your-github-token"
```

2. Start the server:
```bash
cd docker/server
docker-compose up -d
```

3. Use the server:
```bash
# Upload a file
docker-compose exec conf-sync-server upload -g YOUR_GIST_ID /path/to/file

# List files
docker-compose exec conf-sync-server list -g YOUR_GIST_ID

# Delete a file
docker-compose exec conf-sync-server delete -g YOUR_GIST_ID filename
```

### Server Configuration

The server container needs:
1. `GIST_TOKEN` environment variable
2. Files you want to upload (mount them if needed)

Example docker-compose override for mounting files:
```yaml
services:
  conf-sync-server:
    volumes:
      - /path/to/configs:/configs
```

## Building Images

Build both images:
```bash
# Build client
docker build -t conf-sync-client -f docker/client/Dockerfile .

# Build server
docker build -t conf-sync-server -f docker/server/Dockerfile .
```

## Security Notes

1. Client Container:
   - Runs as non-root by default
   - Needs privileged mode only if running system commands
   - Mount config files as read-only when possible

2. Server Container:
   - Runs as non-root
   - Requires `GIST_TOKEN` environment variable
   - No privileged access needed

## Troubleshooting

1. Check container logs:
```bash
# Client logs
docker-compose -f docker/client/docker-compose.yml logs -f

# Server logs
docker-compose -f docker/server/docker-compose.yml logs -f
```

2. Check container status:
```bash
docker ps | grep conf-sync
```

3. Common issues:
   - Missing configuration file
   - Incorrect file permissions
   - Missing GIST_TOKEN
   - Network connectivity issues
