# Conf Sync

A standard Go project structure created on 2025-01-17, modified to be a command-line tool for synchronizing configuration files with GitHub Gist. Built with Go, this tool allows you to easily upload, download, and monitor changes in your configuration files through Gist.

## Project Structure

```
.
├── api/        # API protocols and swagger/OpenAPI specs
├── cmd/        # Main applications
├── configs/    # Configuration files
├── docs/       # Documentation
├── internal/   # Private application and library code
├── pkg/        # Public library code
└── test/       # Additional test applications and test data
```

## Getting Started

To run the application:

```bash
go run cmd/app/main.go
```

## Requirements

- Go 1.21 or higher

## Features

- Upload multiple configuration files to Gist
- Sync configuration files from Gist to local
- Watch for changes in Gist files
- List all files in a Gist
- Delete files from Gist
- Support for environment variables
- Real-time file monitoring
- Execute commands after file updates (e.g., restart services)
- Support for multiple files synchronization
- Secure token-based authentication

## Installation

### Quick Install (From Release)

#### macOS
```bash
# For Intel Mac
curl -L https://github.com/mryee2023/conf-sync/releases/latest/download/conf-sync-darwin-amd64 -o conf-sync

# For Apple Silicon Mac
curl -L https://github.com/mryee2023/conf-sync/releases/latest/download/conf-sync-darwin-arm64 -o conf-sync

# Make it executable
chmod +x conf-sync
sudo mv conf-sync /usr/local/bin/
```

#### Linux
```bash
# For x64
curl -L https://github.com/mryee2023/conf-sync/releases/latest/download/conf-sync-linux-amd64 -o conf-sync

# For ARM64
curl -L https://github.com/mryee2023/conf-sync/releases/latest/download/conf-sync-linux-arm64 -o conf-sync

# Make it executable
chmod +x conf-sync
sudo mv conf-sync /usr/local/bin/
```

#### Windows
Download the appropriate file from the [releases page](https://github.com/mryee2023/conf-sync/releases/latest):
- Windows x64: `conf-sync-windows-amd64.exe`
- Windows ARM64: `conf-sync-windows-arm64.exe`

### From Source

1. Clone the repository:
```bash
git clone https://github.com/mryee2023/conf-sync.git
cd conf-sync
```

2. Build for your platform:
```bash
# Build for current platform
make build

# Or build for all supported platforms
make build-all

# Or install directly to your $GOPATH/bin
make install
```

The built binaries will be available in the `build` directory:
- macOS (Intel): `build/conf-sync-darwin-amd64`
- macOS (Apple Silicon): `build/conf-sync-darwin-arm64`
- Linux (x64): `build/conf-sync-linux-amd64`
- Linux (ARM64): `build/conf-sync-linux-arm64`
- Windows (x64): `build/conf-sync-windows-amd64.exe`
- Windows (ARM64): `build/conf-sync-windows-arm64.exe`

### Using Go Install

```bash
go install github.com/mryee2023/conf-sync/cmd/app@latest
```

## Configuration

The tool requires a GitHub Personal Access Token with `gist` scope. You can create one at [GitHub Settings](https://github.com/settings/tokens).

Set your GitHub token as an environment variable:
```bash
export GIST_TOKEN="your-github-token"
```

## Usage

### Basic Commands

1. Upload multiple configuration files to Gist:
```bash
./conf-sync -g YOUR_GIST_ID upload config.json .env
```

2. List all files in your Gist:
```bash
./conf-sync -g YOUR_GIST_ID list
```

3. Sync specific files from Gist:
```bash
./conf-sync -g YOUR_GIST_ID sync config.json
# Or sync all files
./conf-sync -g YOUR_GIST_ID sync
```

4. Watch files for changes:
```bash
./conf-sync -g YOUR_GIST_ID watch config.json
# Or watch all files
./conf-sync -g YOUR_GIST_ID watch
```

5. Delete files from Gist:
```bash
./conf-sync -g YOUR_GIST_ID delete old-config.json
```

### Command Line Options

- `-g, --gist-id`: (Required) The ID of your Gist
- `-l, --log-level`: Set log level (debug, info, warn, error, default: info)
- `-i, --interval`: Set check interval for watch command (default: 10s)
- `-e, --exec`: Execute command after files are updated

### Log Levels

The following log levels are supported:
- `debug`: Show all log messages, including detailed debugging information
- `info`: Show informational messages (default)
- `warn`: Show only warning and error messages
- `error`: Show only error messages

Example:
```bash
# Show detailed debug information
conf-sync -g YOUR_GIST_ID -l debug watch db.conf:/etc/myapp/db.conf

# Only show errors
conf-sync -g YOUR_GIST_ID -l error watch db.conf:/etc/myapp/db.conf
```

### Time Units

The following time units are supported for the `-i/--interval` flag:
- `ns` (nanoseconds)
- `us` or `µs` (microseconds)
- `ms` (milliseconds)
- `s` (seconds)
- `m` (minutes)
- `h` (hours)

## Examples

1. Upload multiple configuration files:
```bash
./conf-sync -g YOUR_GIST_ID upload app.yaml db.conf .env
```

2. Sync all configuration files from Gist:
```bash
./conf-sync -g YOUR_GIST_ID sync
```

3. Watch specific configuration files for changes:
```bash
./conf-sync -g YOUR_GIST_ID watch config.json app.yaml
```

4. List all files with their status:
```bash
./conf-sync -g YOUR_GIST_ID list
```

5. Delete multiple files:
```bash
./conf-sync -g YOUR_GIST_ID delete old-config.json backup.yaml
```

## Development

### Build Options

The project includes a Makefile with several useful targets:

- `make build` - Build for current platform
- `make build-all` - Build for all supported platforms
- `make clean` - Clean build directory
- `make test` - Run tests
- `make install` - Install locally to $GOPATH/bin

You can specify a version when building:
```bash
make VERSION=1.2.0 build-all
```

### Release Process

To create a new release:

1. Create and push a new tag:
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

2. GitHub Actions will automatically:
   - Build binaries for all supported platforms
   - Create a new release with the binaries
   - Generate release notes

### Supported Platforms

- darwin/amd64 (macOS Intel)
- darwin/arm64 (macOS Apple Silicon)
- linux/amd64 (Linux x64)
- linux/arm64 (Linux ARM64)
- windows/amd64 (Windows x64)
- windows/arm64 (Windows ARM64)

## Notes

- The watch command checks for changes every 10 seconds
- File operations are atomic - if one file fails, others will still be processed
- The tool uses the base filename when uploading to Gist
- Deleted files in Gist will be marked as deleted in the list command
- Be careful with sensitive information in your configuration files

## License

MIT License
