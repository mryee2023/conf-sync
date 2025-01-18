# Change List

## 2025-01-18: Major Refactoring - Split Client and Server

### Architecture Changes

1. Project Structure
   - Split `cmd/app` into separate `cmd/client` and `cmd/server` directories
   - Created dedicated `internal/client` and `internal/server` packages
   - Separated client and server binaries for clearer responsibility separation

2. Build System
   - Updated Makefile to build both client (`conf-sync-client`) and server (`conf-sync-server`) binaries
   - Added separate build targets for client and server
   - Maintained cross-platform build support for all architectures

### Client-side Changes

1. Command Structure
   - Renamed `client` command to `watch` for better clarity
   - Added new `sync` command for manual synchronization
   - Both commands share the same configuration file format

2. Configuration
   - Updated default check interval to 60s for unauthenticated requests
   - Added rate limit handling with exponential backoff
   - Improved error messages and logging

3. File Handling
   - Enhanced file watching logic with better error handling
   - Added support for checking file deletion status
   - Improved directory creation handling

### Server-side Changes

1. Command Structure
   - Created dedicated server binary with focused responsibilities
   - Maintained core commands: `upload`, `list`, and `delete`
   - Added authentication checks for write operations

2. Security
   - Added explicit authentication status checking
   - Improved error messages for unauthorized operations
   - Maintained token-based authentication via environment variables

### Internal Improvements

1. Gist Package
   - Refactored `gist.Client` for better encapsulation
   - Added `GistFile` type for better file representation
   - Fixed type conversion issues with GitHub API
   - Improved rate limit detection and handling

2. Installation
   - Updated install script to handle new binary names
   - Improved service configuration
   - Added check for existing configuration
   - Updated default configuration with new interval settings

### Documentation

1. README Updates
   - Added documentation for new commands
   - Updated installation instructions
   - Added rate limit explanations
   - Improved troubleshooting section

2. Configuration Examples
   - Updated example configurations
   - Added comments about rate limits
   - Improved command execution examples

### Breaking Changes

1. Binary Names
   - Changed from single `conf-sync` to `conf-sync-client` and `conf-sync-server`
   - Requires updates to existing installation scripts
   - Service file needs to be updated for new binary name

2. Command Structure
   - `client` command replaced with `watch`
   - Added new `sync` command
   - Server commands now in separate binary

3. Configuration
   - Minimum check interval increased to 60s for unauthenticated requests
   - Changed some configuration field names for clarity
   - Added new fields for rate limit handling

### Migration Guide

1. For Client Users
   ```bash
   # Stop existing service
   systemctl stop conf-sync

   # Update binary and config
   curl -L https://raw.githubusercontent.com/mryee2023/conf-sync/main/scripts/install.sh | sudo bash

   # Start new service
   systemctl start conf-sync
   ```

2. For Server Users
   ```bash
   # Download new server binary
   curl -L -o /usr/local/bin/conf-sync-server "https://github.com/mryee2023/conf-sync/releases/latest/download/conf-sync-server-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')"
   chmod +x /usr/local/bin/conf-sync-server

   # Update commands to use new binary name
   conf-sync-server -g YOUR_GIST_ID [command]
   ```
