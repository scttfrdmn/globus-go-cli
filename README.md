<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Globus Go CLI

A command-line interface for Globus services, built in Go using the [Globus Go SDK](https://github.com/scttfrdmn/globus-go-sdk).

## Features

- Modern CLI implementation with Cobra and Viper
- Comprehensive coverage of Globus services:
  - Auth
  - Transfer (in development)
  - Groups (planned)
  - Search (planned)
  - Flows (planned)
  - Compute (planned)
  - Timers (planned)
- Multiple output formats (text, JSON, CSV)
- Interactive features with progress visualization
- Multiple configuration profiles
- Cross-platform support (Linux, macOS, Windows)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/scttfrdmn/globus-go-cli.git
cd globus-go-cli

# Build the binary
go build -o globus

# Install the binary
mv globus /usr/local/bin/
```

## Usage

```bash
# Login to Globus
globus auth login

# Show information about the current user
globus auth whoami

# List tokens
globus auth tokens show

# Logout
globus auth logout

# Get help
globus --help
```

## Configuration

The CLI stores its configuration and tokens in `~/.globus-cli/`:

- `~/.globus-cli/config.yaml`: General configuration
- `~/.globus-cli/tokens/`: OAuth tokens for different profiles
- `~/.globus-cli/profiles/`: Named configuration profiles

You can initialize the configuration with:

```bash
globus config init
```

### Configuration Profiles

You can use multiple configuration profiles:

```bash
# Create a new profile
globus config profile create myprofile

# Use a specific profile
globus --profile=myprofile auth login
```

## Commands

### Auth Commands

- `globus auth login`: Log in to Globus
- `globus auth logout`: Log out from Globus
- `globus auth whoami`: Show current user information
- `globus auth tokens show`: Show token information
- `globus auth tokens revoke`: Revoke tokens
- `globus auth tokens introspect`: Introspect token

### Transfer Commands (Coming Soon)

- `globus transfer endpoint list`: List endpoints
- `globus transfer ls <endpoint> <path>`: List files and directories
- `globus transfer mkdir <endpoint> <path>`: Create directory
- `globus transfer cp <source> <destination>`: Transfer files
- `globus transfer rm <endpoint> <path>`: Delete files
- `globus transfer task show <task_id>`: Show task status

### Configuration Commands

- `globus config show`: Show current configuration
- `globus config init`: Initialize configuration

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Apache License 2.0