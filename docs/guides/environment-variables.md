# Environment Variables

The Globus Go CLI recognizes several environment variables for configuration.

## General Configuration

### GLOBUS_CLI_FORMAT

Set the default output format.

```bash
export GLOBUS_CLI_FORMAT=json  # or 'text'
```

### GLOBUS_CLI_DEBUG

Enable debug output.

```bash
export GLOBUS_CLI_DEBUG=true
```

## Authentication

### GLOBUS_CLIENT_ID

Client ID for client credentials authentication.

```bash
export GLOBUS_CLIENT_ID=your-client-id
```

### GLOBUS_CLIENT_SECRET

Client secret for client credentials authentication.

```bash
export GLOBUS_CLIENT_SECRET=your-client-secret
```

!!! warning
    Never commit client secrets to version control. Use secret management tools in production.

## Configuration File Location

### GLOBUS_CONFIG_DIR

Override the default configuration directory location.

```bash
export GLOBUS_CONFIG_DIR=/custom/path/to/config
```

Default locations:

- Linux/macOS: `~/.globus`
- Windows: `%USERPROFILE%\.globus`

## Transfer Options

### GLOBUS_CLI_SYNC_LEVEL

Set the default sync level for transfers.

```bash
export GLOBUS_CLI_SYNC_LEVEL=mtime
```

Values: `exists`, `size`, `mtime`, `checksum`

## Using in Scripts

### Bash Example

```bash
#!/bin/bash

# Set output to JSON for parsing
export GLOBUS_CLI_FORMAT=json

# Authenticate with client credentials
export GLOBUS_CLIENT_ID=abc123
export GLOBUS_CLIENT_SECRET=secret456

# Run commands
globus auth whoami
globus transfer endpoint list
```

### Python Example

```python
import os
import subprocess

# Set environment
os.environ['GLOBUS_CLI_FORMAT'] = 'json'

# Run command
result = subprocess.run(
    ['globus', 'auth', 'whoami'],
    capture_output=True,
    text=True
)
print(result.stdout)
```

## Priority Order

Configuration values are resolved in this order (highest to lowest):

1. Command-line flags (`--format json`)
2. Environment variables (`GLOBUS_CLI_FORMAT=json`)
3. Configuration file (`~/.globus/config.yml`)
4. Default values

## See Also

- [Configuration](../getting-started/configuration.md)
- [Common Tasks](common-tasks.md)
- [Output Formats](output-formats.md)
