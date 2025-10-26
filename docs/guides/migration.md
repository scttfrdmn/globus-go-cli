# Migration from Python CLI

Guide for users migrating from the Python Globus CLI to the Go CLI.

## Key Differences

### Installation

**Python CLI:**

```bash
pipx install globus-cli
```

**Go CLI:**

```bash
brew install globus-go-cli
```

See [Installation](../getting-started/installation.md) for more options.

### Command Compatibility

The Go CLI maintains command compatibility with the Python CLI. Most commands work identically:

```bash
# These work the same in both CLIs
globus login
globus auth whoami
globus transfer task list
globus search index list
```

### Additional Features

The Go CLI includes the **Compute service**, which is not available in the Python CLI:

```bash
globus compute endpoint list
globus compute function register
```

## Configuration

### Token Storage

Both CLIs store tokens in the same location:

- **Linux/macOS**: `~/.globus/tokens.json`
- **Windows**: `%USERPROFILE%\.globus\tokens.json`

!!! tip
    You can use both CLIs with the same authentication tokens.

### Configuration Files

Configuration format is compatible between both CLIs.

## Output Format

Default output format differs:

- **Python CLI**: Text format by default
- **Go CLI**: JSON format by default

Set JSON as default in Go CLI:

```bash
export GLOBUS_CLI_FORMAT=json
```

## Performance

The Go CLI offers:

- **Faster startup** - No Python interpreter overhead
- **Lower memory usage** - Compiled binary vs interpreted Python
- **Single binary** - No dependency management

## Feature Parity

| Feature | Python CLI | Go CLI |
|---------|------------|--------|
| Auth | ✅ | ✅ |
| Transfer | ✅ | ✅ |
| Search | ✅ | ✅ |
| Groups | ✅ | ✅ |
| Flows | ✅ | ✅ |
| Timers | ✅ | ✅ |
| Compute | ❌ | ✅ |

## Known Differences

### Command Names

All core commands use identical names. No changes needed to existing scripts.

### Output Format

Default output format differs as noted above. Explicitly specify `--format` in scripts for consistency.

### Shell Completion

Both CLIs support shell completion for Bash, Zsh, Fish, and PowerShell.

**Python CLI:**

```bash
eval "$(globus --completion bash)"
```

**Go CLI:**

```bash
source <(globus completion bash)
```

## Migration Checklist

- [ ] Install Go CLI using preferred method
- [ ] Verify authentication works: `globus auth whoami`
- [ ] Test critical commands from your scripts
- [ ] Update scripts to specify `--format` explicitly
- [ ] Update shell completion configuration
- [ ] Test transfers and verify expected behavior
- [ ] Update documentation to reference Go CLI

## Side-by-Side Usage

You can run both CLIs simultaneously:

- Use `globus` for Go CLI (if installed via package manager)
- Use `python -m globus_cli` for Python CLI

## Getting Help

If you encounter issues during migration:

- Check [Command Reference](../reference/index.md)
- Review [Common Tasks](common-tasks.md)
- Open an issue on [GitHub](https://github.com/scttfrdmn/globus-go-cli/issues)

## See Also

- [Quick Start](../getting-started/quickstart.md)
- [Command Reference](../reference/index.md)
- [Common Tasks](common-tasks.md)
