# Command Reference

Complete reference documentation for all Globus Go CLI commands.

## Command Structure

Most commands follow this structure:

```bash
globus <command> [subcommand] [flags]
```

Example:

```bash
globus task list --limit 10
```

## Available Services

The CLI provides commands for these Globus services:

### [Auth Commands](auth.md)

Authentication and identity management.

```bash
globus whoami
globus login
globus logout
```

### [Transfer Commands](transfer.md)

File and directory transfer operations.

```bash
globus task list
globus ls ENDPOINT_ID:/path
globus transfer --source-endpoint SRC --dest-endpoint DST ...
```

### [Search Commands](search.md)

Search index management and queries.

```bash
globus search index list
globus search query "search terms"
```

### [Groups Commands](groups.md)

Group membership and policy management.

```bash
globus groups list
globus groups show GROUP_ID
globus groups member add GROUP_ID USER_ID
```

### [Flows Commands](flows.md)

Workflow automation and management.

```bash
globus flows list
globus flows run FLOW_ID
globus flows run show RUN_ID
```

### [Timers Commands](timers.md)

Scheduled task execution.

```bash
globus timers list
globus timers create --name "My Timer" --schedule "0 0 * * *"
```

### [Compute Commands](compute.md)

Distributed computing operations (Go CLI exclusive feature).

```bash
globus compute endpoint list
globus compute function register
```

## Global Flags

These flags are available for all commands:

| Flag | Description |
|------|-------------|
| `--format string` | Output format: `json` or `text` (default: `json`) |
| `--help` | Show help for the command |
| `--version` | Show CLI version |

## Output Formats

### JSON Format

Default output format, suitable for parsing:

```bash
globus whoami --format json
```

Output:

```json
{
  "username": "user@example.edu",
  "email": "user@example.edu",
  "identity_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
}
```

### Text Format

Human-readable tabular output:

```bash
globus whoami --format text
```

## Getting Help

Get help for any command:

```bash
# Service-level help
globus transfer --help

# Command-level help
globus task --help

# Subcommand-level help
globus task list --help
```

## Exit Codes

The CLI uses standard exit codes:

- `0` - Success
- `1` - General error
- `2` - Command-line syntax error

## Next Steps

Select a service from the list above to view detailed command documentation, or explore:

- [Common Tasks](../guides/common-tasks.md) - Practical examples
- [Output Formats](../guides/output-formats.md) - Working with output
- [Environment Variables](../guides/environment-variables.md) - Configuration options
