# Globus Go CLI

Welcome to the Globus Go CLI documentation! This is a command-line interface for Globus services, built in Go.

## Overview

The Globus Go CLI provides a powerful command-line interface for interacting with Globus services including file transfer, search, groups, flows, timers, and compute. Built in Go, it offers cross-platform support, fast performance, and a single binary distribution.

## Key Features

- **🚀 Single Binary Distribution** - No runtime dependencies, just download and run
- **⚡ Fast Performance** - Written in Go for optimal speed and resource usage
- **🌐 Cross-Platform** - Linux, macOS, Windows (including ARM64 architectures)
- **🔄 Full Service Support** - Auth, Transfer, Search, Groups, Flows, Timers, and Compute
- **📦 Multiple Installation Methods** - Homebrew, Scoop, Docker, or direct binary download
- **🎨 Flexible Output** - JSON, text, and formatted output options
- **🔐 Secure Authentication** - OAuth2-based authentication with token management

## Feature Parity

This Go implementation maintains feature parity with the upstream [Python Globus CLI](https://github.com/globus/globus-cli) (v3.39.0), with the addition of exclusive **Compute service** support.

### Supported Services

| Service | Status | Description |
|---------|--------|-------------|
| **Auth** | ✅ Complete | Authentication and identity management |
| **Transfer** | ✅ Complete | File and directory transfer operations |
| **Search** | ✅ Complete | Search index management and queries |
| **Groups** | ✅ Complete | Group membership and policy management |
| **Flows** | ✅ Complete | Workflow automation and management |
| **Timers** | ✅ Complete | Scheduled task execution |
| **Compute** | ✅ Complete | Distributed computing (Go CLI exclusive) |

## Quick Links

- [Installation Guide](getting-started/installation.md) - Get started with installing the CLI
- [Quick Start](getting-started/quickstart.md) - Your first commands
- [Command Reference](reference/index.md) - Complete command documentation
- [GitHub Repository](https://github.com/scttfrdmn/globus-go-cli) - Source code and issues

## Getting Help

If you need help or want to report an issue:

- Check the [Command Reference](reference/index.md) for detailed command documentation
- Review [Common Tasks](guides/common-tasks.md) for practical examples
- Open an issue on [GitHub](https://github.com/scttfrdmn/globus-go-cli/issues)

## About This Project

The Globus Go CLI is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus or the University of Chicago. It is maintained by independent contributors and designed to be compatible with the upstream Python CLI.

## Next Steps

Ready to get started? Head over to the [Installation Guide](getting-started/installation.md) to install the CLI on your system.
