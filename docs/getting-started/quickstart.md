# Quick Start

This guide will help you get started with the Globus Go CLI by walking through the essential first steps.

## Prerequisites

Before you begin, ensure you have:

- [Installed the Globus Go CLI](installation.md)
- A Globus account (create one at [globus.org](https://www.globus.org))

## Authentication

The first step is to authenticate with Globus:

```bash
globus login
```

This will:

1. Open your web browser to the Globus authentication page
2. Prompt you to log in with your Globus account
3. Ask you to grant permissions to the CLI
4. Provide an authorization code to paste back into the terminal

After successful authentication, you'll see:

```
You have successfully logged in to the Globus CLI!
```

### Verify Authentication

Check that you're logged in:

```bash
globus auth whoami
```

This displays your identity information:

```json
{
  "username": "yourname@example.edu",
  "name": "Your Name",
  "email": "yourname@example.edu",
  "identity_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
}
```

## Your First Commands

### List Your Endpoints

View endpoints you have access to:

```bash
globus transfer endpoint list
```

### View Endpoint Details

Get details about a specific endpoint:

```bash
globus transfer endpoint show ENDPOINT_ID
```

Replace `ENDPOINT_ID` with an actual endpoint UUID.

### List Files

Browse files on an endpoint:

```bash
globus transfer ls ENDPOINT_ID:/path/to/directory
```

### Transfer Files

Transfer a file between two endpoints:

```bash
globus transfer transfer \
  --source-endpoint SOURCE_ENDPOINT_ID \
  --dest-endpoint DEST_ENDPOINT_ID \
  --source-path /path/to/source/file.txt \
  --dest-path /path/to/destination/file.txt
```

Check transfer status:

```bash
globus transfer task show TASK_ID
```

## Output Formats

The CLI supports multiple output formats:

### JSON Output (Default)

```bash
globus auth whoami --format json
```

### Text Output

```bash
globus auth whoami --format text
```

### Quiet Mode

For scripting, use quiet mode to get minimal output:

```bash
globus transfer task list --format json --limit 1 | jq -r '.[0].task_id'
```

## Common Options

Most commands support these global options:

| Option | Description |
|--------|-------------|
| `--format` | Output format: `json` or `text` |
| `--help` | Show help for the command |
| `--version` | Show CLI version |

## Environment Variables

Set common options via environment variables:

```bash
# Set default output format
export GLOBUS_CLI_FORMAT=json

# Enable debug output
export GLOBUS_CLI_DEBUG=true
```

See [Environment Variables](../guides/environment-variables.md) for a complete list.

## Getting Help

### Command Help

Get help for any command:

```bash
globus --help
globus transfer --help
globus transfer task list --help
```

### Available Commands

List all available commands:

```bash
globus --help
```

Main command groups:

- `auth` - Authentication and identity management
- `transfer` - File transfer operations
- `search` - Search index management
- `groups` - Group membership management
- `flows` - Workflow automation
- `timers` - Scheduled tasks
- `compute` - Distributed computing

## Working with Configuration

### View Configuration

```bash
globus config show
```

### Set Configuration Values

```bash
globus config set output_format json
```

## Logging Out

When you're done, log out to remove stored credentials:

```bash
globus logout
```

## Next Steps

- Learn about [Authentication](authentication.md) in detail
- Explore [Common Tasks](../guides/common-tasks.md) for practical examples
- Review the [Command Reference](../reference/index.md) for complete documentation

## Need Help?

- Check the [Command Reference](../reference/index.md)
- Visit the [GitHub repository](https://github.com/scttfrdmn/globus-go-cli)
- Review [Common Tasks](../guides/common-tasks.md)
