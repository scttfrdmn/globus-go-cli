# Installation

The Globus Go CLI can be installed using multiple methods depending on your platform and preferences.

## Homebrew (macOS and Linux)

The easiest way to install on macOS or Linux is using Homebrew:

```bash
# Add the tap
brew tap scttfrdmn/tap

# Install the CLI
brew install globus-go-cli
```

To upgrade to the latest version:

```bash
brew upgrade globus-go-cli
```

## Scoop (Windows)

For Windows users, Scoop provides an easy installation method:

```powershell
# Add the bucket
scoop bucket add scttfrdmn https://github.com/scttfrdmn/scoop-bucket

# Install the CLI
scoop install globus-go-cli
```

To upgrade:

```powershell
scoop update globus-go-cli
```

## Docker

Run the CLI using Docker without installing:

```bash
# Run a command
docker run --rm -it scttfrdmn/globus-go-cli:latest auth whoami

# Create an alias for convenience
alias globus='docker run --rm -it -v ~/.globus:/root/.globus scttfrdmn/globus-go-cli:latest'
```

The `-v ~/.globus:/root/.globus` mount preserves your authentication tokens between runs.

## Direct Binary Download

Download pre-built binaries from the [GitHub Releases](https://github.com/scttfrdmn/globus-go-cli/releases) page.

### Linux

```bash
# Download the latest release (replace VERSION with actual version)
wget https://github.com/scttfrdmn/globus-go-cli/releases/download/VERSION/globus-go-cli_Linux_x86_64.tar.gz

# Extract
tar -xzf globus-go-cli_Linux_x86_64.tar.gz

# Move to PATH
sudo mv globus /usr/local/bin/

# Verify installation
globus --version
```

### macOS

```bash
# Download the latest release (replace VERSION with actual version)
curl -LO https://github.com/scttfrdmn/globus-go-cli/releases/download/VERSION/globus-go-cli_Darwin_x86_64.tar.gz

# Extract
tar -xzf globus-go-cli_Darwin_x86_64.tar.gz

# Move to PATH
sudo mv globus /usr/local/bin/

# Verify installation
globus --version
```

### Windows

1. Download the `.zip` file from [GitHub Releases](https://github.com/scttfrdmn/globus-go-cli/releases)
2. Extract the archive
3. Add the directory containing `globus.exe` to your PATH

## Build from Source

If you have Go 1.22+ installed:

```bash
# Clone the repository
git clone https://github.com/scttfrdmn/globus-go-cli.git
cd globus-go-cli

# Build
make build

# Or use go directly
go build -o globus

# Move to PATH
sudo mv globus /usr/local/bin/
```

## Verify Installation

After installation, verify the CLI is working:

```bash
globus --version
```

You should see output like:

```
globus version 3.39.0-2
```

## Shell Completion

The CLI supports shell completion for Bash, Zsh, Fish, and PowerShell.

### Bash

```bash
# Add to your ~/.bashrc
source <(globus completion bash)

# Or install permanently
globus completion bash > /etc/bash_completion.d/globus
```

### Zsh

```bash
# Add to your ~/.zshrc
source <(globus completion zsh)

# Or install permanently
globus completion zsh > "${fpath[1]}/_globus"
```

### Fish

```bash
globus completion fish | source

# Or install permanently
globus completion fish > ~/.config/fish/completions/globus.fish
```

### PowerShell

```powershell
# Add to your profile
globus completion powershell | Out-String | Invoke-Expression
```

## Platform Support

Binaries are provided for:

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

## Next Steps

Now that you have the CLI installed, proceed to the [Quick Start](quickstart.md) guide to authenticate and run your first commands.
