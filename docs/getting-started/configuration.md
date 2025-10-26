# Configuration

The Globus Go CLI can be configured through configuration files, environment variables, and command-line flags.

## Configuration File

The CLI stores configuration in:

- **Linux/macOS**: `~/.globus/config.yml`
- **Windows**: `%USERPROFILE%\.globus\config.yml`

## Viewing Configuration

View current configuration:

```bash
globus config show
```

## Setting Configuration Values

Set a configuration value:

```bash
globus config set KEY VALUE
```

Examples:

```bash
# Set default output format
globus config set output_format json

# Set transfer sync level
globus config set sync_level mtime
```

## Configuration Options

Common configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `output_format` | Default output format (`json`, `text`) | `json` |
| `sync_level` | Transfer sync level | `mtime` |
| `debug` | Enable debug output | `false` |

## Environment Variables

Environment variables override configuration file settings.

### Common Variables

```bash
# Output format
export GLOBUS_CLI_FORMAT=json

# Enable debug mode
export GLOBUS_CLI_DEBUG=true

# Client credentials
export GLOBUS_CLIENT_ID=your-client-id
export GLOBUS_CLIENT_SECRET=your-client-secret
```

See [Environment Variables](../guides/environment-variables.md) for a complete list.

## Command-Line Flags

Flags override both configuration and environment variables:

```bash
globus auth whoami --format text
```

## Priority Order

Configuration is resolved in this order (highest to lowest):

1. **Command-line flags** - `--format json`
2. **Environment variables** - `GLOBUS_CLI_FORMAT=json`
3. **Configuration file** - `~/.globus/config.yml`
4. **Default values**

## Configuration File Format

The configuration file uses YAML format:

```yaml
# ~/.globus/config.yml
output_format: json
sync_level: mtime
debug: false
```

## Resetting Configuration

Remove a configuration value:

```bash
globus config unset KEY
```

Reset all configuration:

```bash
rm ~/.globus/config.yml
```

## Next Steps

- Learn about [Environment Variables](../guides/environment-variables.md)
- Try [Common Tasks](../guides/common-tasks.md)
- Review [Command Reference](../reference/index.md)
