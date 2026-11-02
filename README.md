<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors -->

# Globus Go CLI

A command-line interface for Globus services, built in Go on the v4 [Globus Go SDK](https://github.com/scttfrdmn/globus-go-sdk) (v4.10.0-3, tracking Python globus-sdk 4.10.0). It aims to be a drop-in replacement for the Python [globus-cli](https://github.com/globus/globus-cli), currently compared against **globus-cli 3.44.0**.

Versioning: the CLI version tracks the Python globus-sdk version the Go SDK ports (`4.10.0`), and the `-N` suffix is this project's release number.

## Features

- Modern CLI implementation with Cobra and Viper
- Fast and lightweight with no Python dependencies
- The Python CLI's flat command layout, global flags (`-F json|text|unix`, `--jmespath`, `--map-http-status`) and enveloped JSON output
- Coverage of Globus services. See [docs/PYTHON_CLI_PARITY.md](docs/PYTHON_CLI_PARITY.md) for the command-by-command comparison and the open gaps:
  - Auth: `login`, `logout`, `whoami`, `get-identities`, `session`
  - Transfer: `ls`, `transfer`, `delete`, `task`, `endpoint`, `bookmark`, `endpoint-manager`, `streams`
  - Globus Connect Server (`collection`, `gcs`) and Globus Connect Personal (`gcp`)
  - Groups, Search, Timers
  - Flows, including `flows web-input` and `flows registered-api`
  - `api` raw passthrough
  - Go-only extensions: `compute`, plus `project` (Auth project, client and credential management)
- Multiple output formats (text, JSON, CSV)
- Interactive features with progress visualization
- Multiple configuration profiles
- Cross-platform support (Linux, macOS, Windows)
- Shell completion for Bash, Zsh, Fish, and PowerShell

## Documentation

**📚 [Read the full documentation](https://scttfrdmn.github.io/globus-go-cli/)**

- [Installation Guide](https://scttfrdmn.github.io/globus-go-cli/getting-started/installation/)
- [Quick Start](https://scttfrdmn.github.io/globus-go-cli/getting-started/quickstart/)
- [Command Reference](https://scttfrdmn.github.io/globus-go-cli/reference/)
- [Common Tasks](https://scttfrdmn.github.io/globus-go-cli/guides/common-tasks/)

## Installation

### Using Homebrew (macOS and Linux)

```bash
# Install from Homebrew
brew tap scttfrdmn/tap
brew install globus-go-cli
```

### Using Scoop (Windows)

```powershell
# Add the scoop bucket
scoop bucket add scttfrdmn https://github.com/scttfrdmn/scoop-bucket

# Install globus-go-cli
scoop install globus-go-cli
```

### Using Docker

```bash
# Run using Docker
docker run --rm -it scttfrdmn/globus-go-cli:latest auth whoami
```

### From Binary Releases

Download the latest release for your platform from the [Releases page](https://github.com/scttfrdmn/globus-go-cli/releases).

Binaries are provided for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

### From Source

```bash
# Clone the repository
git clone https://github.com/scttfrdmn/globus-go-cli.git
cd globus-go-cli

# Build the binary
make build
# or
go build -o globus

# Install the binary
mv globus /usr/local/bin/
```

## Quick Start

```bash
# Login to Globus
globus login

# Or login without a browser
globus device

# Show information about the current user
globus whoami

# List your endpoints
globus endpoint list

# List files on an endpoint
globus ls ENDPOINT_ID:/path

# Transfer files between endpoints
globus transfer SOURCE_ENDPOINT:/source/path DEST_ENDPOINT:/dest/path

# Check transfer status
globus task show TASK_ID

# Logout when done
globus logout
```

## Configuration

The CLI stores its configuration and tokens in `~/.globus-cli/`:

- `~/.globus-cli/config.yaml`: General configuration
- `~/.globus-cli/tokens/`: OAuth tokens for different profiles
- `~/.globus-cli/profiles/`: Named configuration profiles

### Configuration Profiles

You can use multiple configuration profiles to work with different Globus accounts:

```bash
# Create a new profile
globus config profile create myprofile

# Use a specific profile
globus --profile=myprofile auth login

# List all profiles
globus config profile list
```

### Output Formats

Most commands support different output formats:

```bash
# Default text format
globus endpoint list

# JSON output for scripting
globus endpoint list --format=json

# CSV output for importing into spreadsheets
globus endpoint list --format=csv
```

## Detailed Command Reference

### Auth Commands

```bash
# Log in using browser
globus login

# Log in with device code (no browser)
globus device

# Show current user info
globus whoami

# List tokens
globus tokens show

# Refresh tokens
globus refresh

# Revoke tokens
globus tokens revoke --type=access

# Look up identities
globus get-identities user@example.com

# Log out
globus logout
```

### Transfer Commands

```bash
# List endpoints
globus endpoint list

# Search for endpoints
globus endpoint search "my data"

# Show endpoint details
globus endpoint show ENDPOINT_ID

# List files on endpoint
globus ls ENDPOINT_ID:/path
globus ls -l ENDPOINT_ID:/path  # long format

# Create directory
globus mkdir ENDPOINT_ID:/new/directory
globus mkdir -p ENDPOINT_ID:/nested/directory  # create parents

# Delete files/directories
globus rm ENDPOINT_ID:/path/to/file
globus rm -r ENDPOINT_ID:/directory  # recursive

# Transfer files
globus transfer SOURCE_EP:/file DEST_EP:/path
globus transfer -r SOURCE_EP:/dir DEST_EP:/path  # recursive

# List tasks
globus task list

# View task details
globus task show TASK_ID

# Wait for task completion
globus task wait TASK_ID

# Cancel task
globus task cancel TASK_ID
```

### Shell Completion

```bash
# Generate shell completion scripts
globus completion bash > ~/.bash_completion.d/globus
globus completion zsh > "${fpath[1]}/_globus"
globus completion fish > ~/.config/fish/completions/globus.fish
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Development

### Integration Testing

For information about setting up integration testing with real Globus credentials, see [INTEGRATION_TESTING.md](INTEGRATION_TESTING.md).

### Cross-Platform Compatibility

For guidelines on ensuring cross-platform compatibility, see [CROSS_PLATFORM.md](CROSS_PLATFORM.md).

## Release Process

For information about the release process, see [RELEASE_PROCESS.md](RELEASE_PROCESS.md).

## Release Notes

- [GitHub Releases](https://github.com/scttfrdmn/globus-go-cli/releases) - Published releases
- [Changelog](CHANGELOG.md) - Full history of changes

## License

Apache License 2.0