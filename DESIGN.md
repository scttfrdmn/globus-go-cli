<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors -->

# Globus Go CLI - Design Document

## Overview

The Globus Go CLI is a command-line interface for interacting with Globus services, built using the Globus Go SDK. This CLI aims to provide comparable functionality to the official Python-based Globus CLI while leveraging the performance advantages of Go.

## Design Goals

1. **Comprehensive Coverage**: Support all major Globus services (Auth, Transfer, Search, Groups, Flows, Compute, Timers)
2. **User-Friendly**: Provide intuitive command structure and helpful error messages
3. **Performance**: Leverage Go's performance advantages for large operations
4. **Extensibility**: Design for easy addition of new commands and services
5. **Consistency**: Maintain consistent CLI patterns across all commands
6. **Configuration Management**: Support multiple profiles and flexible configuration options
7. **Interactive Features**: Provide interactive prompts and progress visualization
8. **Cross-Platform**: Work across Linux, macOS, and Windows

## Architecture

### Command Structure

The CLI will use a nested command structure:

```
globus <service> <command> [subcommand] [options] [arguments]
```

Where:
- `<service>` is a Globus service (auth, transfer, groups, etc.)
- `<command>` is an operation within that service
- `[subcommand]` is an optional, more specific operation
- `[options]` are flags like `--verbose` or `--format json`
- `[arguments]` are positional parameters

### Components

1. **CLI Framework**: Using [spf13/cobra](https://github.com/spf13/cobra) for command structure
2. **Configuration System**: Manages CLI settings, using [spf13/viper](https://github.com/spf13/viper)
3. **Authentication Manager**: Handles token acquisition and refresh
4. **Output Formatter**: Supports multiple output formats (text, JSON, CSV)
5. **Service Modules**: One module per Globus service
6. **Interactive Components**: Progress bars, spinners, and interactive prompts
7. **SDK Integration**: Clean integration with the Globus Go SDK

### Service Modules

Each Globus service will have its own module:

1. **Auth**: Login, logout, token management, identity operations
2. **Transfer**: File transfers, endpoint management, task monitoring
3. **Groups**: Group creation and membership management
4. **Search**: Index and query operations
5. **Flows**: Flow creation, execution, and monitoring
6. **Compute**: Function management, execution, and monitoring
7. **Timers**: Timer creation and management

## Configuration System

The CLI will store configuration and credentials in `~/.globus-cli/`:

1. **config.yaml**: General configuration settings
2. **profiles/**: Named configuration profiles
3. **tokens/**: OAuth tokens for different profiles
4. **bookmarks/**: Saved endpoint and collection references

Configuration settings can be set via:
1. Command-line flags
2. Environment variables (prefixed with `GLOBUS_`)
3. Configuration files

## Command Patterns

All commands will follow consistent patterns:

1. **Listing Resources**: `globus <service> list [filters]`
2. **Getting Details**: `globus <service> show <resource-id>`
3. **Creating Resources**: `globus <service> create [options]`
4. **Updating Resources**: `globus <service> update <resource-id> [options]`
5. **Deleting Resources**: `globus <service> delete <resource-id>`
6. **Operations**: `globus <service> <operation-name> [options]`

## Example Command Structure

```
globus
├── auth
│   ├── login
│   ├── logout
│   ├── whoami
│   ├── tokens
│   │   ├── list
│   │   ├── revoke
│   │   └── introspect
│   └── identities
├── transfer
│   ├── endpoint
│   │   ├── list
│   │   ├── show
│   │   ├── update
│   │   └── search
│   ├── ls
│   ├── mkdir
│   ├── cp (transfer)
│   ├── rm (delete)
│   └── task
│       ├── list
│       ├── show
│       ├── cancel
│       └── wait
├── groups
│   ├── list
│   ├── show
│   ├── create
│   ├── delete
│   └── membership
│       ├── list
│       ├── add
│       ├── remove
│       └── update
├── search
│   ├── index
│   │   ├── list
│   │   ├── show
│   │   ├── create
│   │   └── delete
│   ├── query
│   ├── ingest
│   └── delete
├── flows
│   ├── list
│   ├── show
│   ├── run
│   └── status
├── compute
│   ├── endpoints
│   ├── functions
│   ├── containers
│   ├── run
│   └── status
├── timers
│   ├── list
│   ├── create
│   ├── update
│   └── delete
└── config
    ├── show
    ├── init
    ├── profile
    │   ├── list
    │   ├── create
    │   ├── update
    │   └── delete
    └── bookmarks
```

## Implementation Plan

### Phase 1: Core Framework
- CLI framework with Cobra
- Configuration system with Viper
- Auth service commands (login, logout, tokens)
- Basic output formatting

### Phase 2: Transfer Service
- Endpoint management
- File operations (ls, rm, mkdir)
- Transfer operations
- Task management

### Phase 3: Additional Services
- Groups service
- Search service
- Flows service
- Compute service
- Timers service

### Phase 4: Advanced Features
- Interactive prompts
- Progress visualization
- Multiple profiles
- Advanced output formatting
- Tab completion

## Interactive Features

The CLI will support interactive features:

1. **Progress Bars**: For transfers and long-running operations
2. **Spinners**: For operations with indeterminate duration
3. **Interactive Prompts**: For collecting complex inputs
4. **Endpoint Browser**: For interactively selecting files and directories
5. **Confirmation Prompts**: For destructive operations

## Comparison to Python CLI

While aiming for command compatibility with the Python CLI where appropriate, this implementation will focus on:

1. **Performance**: Faster execution, especially for large transfers
2. **Memory Efficiency**: Lower resource usage
3. **Modern CLI Features**: Interactive elements and visualizations
4. **Binary Distribution**: No runtime dependencies

## Requirements

1. Go 1.20 or higher
2. Globus Go SDK
3. Cobra, Viper, and other Go libraries

## Extensibility

The CLI will be designed for extensibility:

1. **Plugin Architecture**: Allow adding commands without modifying core code
2. **Command Discovery**: Automatically discover and register commands
3. **Templating**: Common patterns for adding new commands
4. **Middleware**: For cross-cutting concerns like authentication

## Security Considerations

1. **Token Storage**: Secure storage of access and refresh tokens
2. **Credential Handling**: Avoid exposing credentials in process listings
3. **Local-Only Operations**: No network requests for some operations
4. **Config Permissions**: Enforce appropriate permissions on config files