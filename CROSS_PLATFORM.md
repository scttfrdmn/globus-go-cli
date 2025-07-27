# Cross-Platform Compatibility Guide

This document outlines guidelines and best practices for ensuring that the Globus CLI works correctly across different operating systems.

## Supported Platforms

The Globus CLI is designed to work on the following platforms:

- **Linux** - Various distributions (Debian, Ubuntu, CentOS, etc.)
- **macOS** - 10.15 (Catalina) and newer
- **Windows** - Windows 10 and newer

## Ensuring Cross-Platform Compatibility

### File Paths

- Use `filepath.Join()` instead of string concatenation for paths
- Use `os.PathSeparator` when needed instead of hardcoding `/` or `\`
- Use `filepath.Clean()` to normalize paths

```go
// Good
path := filepath.Join(homeDir, ".globus-cli", "config")

// Bad
path := homeDir + "/.globus-cli/config" // Will fail on Windows
```

### Home Directory

Always use the platform-specific way to get the home directory:

```go
homeDir, err := os.UserHomeDir()
if err != nil {
    return fmt.Errorf("error getting home directory: %w", err)
}
```

### File Permissions

Be careful with file permissions as they work differently on Windows:

```go
// For directories (cross-platform safe)
err := os.MkdirAll(dir, 0700)

// For files (cross-platform safe)
err := os.WriteFile(file, data, 0600)
```

### Environment Variables

- Environment variable names are case-sensitive on Unix but case-insensitive on Windows
- Always use consistent casing in your code

### Line Endings

- Git should handle line ending conversions automatically if `.gitattributes` is configured
- When reading files, handle both `\n` and `\r\n` line endings

### File Locking

If you need to implement file locking, be aware that it works differently across platforms:

- Use the `github.com/gofrs/flock` library for cross-platform file locking

### Building for Multiple Platforms

To build for multiple platforms locally:

```bash
# Build for Windows from a Unix-like OS
GOOS=windows GOARCH=amd64 go build -o globus.exe

# Build for macOS from another OS
GOOS=darwin GOARCH=amd64 go build -o globus-macos

# Build for Linux from another OS
GOOS=linux GOARCH=amd64 go build -o globus-linux
```

## Testing Cross-Platform Compatibility

### CI/CD Testing

We use GitHub Actions to test on multiple platforms:

- The `cross-platform.yml` workflow builds and tests on Linux, macOS, and Windows
- Run tests on multiple Go versions to ensure compatibility

### Manual Testing

For thorough testing before a release:

1. Test installation from source on each platform
2. Test all major commands on each platform
3. Test with both relative and absolute paths
4. Test with paths containing spaces and special characters
5. Test with non-ASCII characters in paths and filenames

## Common Cross-Platform Issues

### Path Separators

Windows uses backslashes, while Unix-like systems use forward slashes. Use `filepath.Join()` to handle this automatically.

### Case Sensitivity

- Windows file systems are generally case-insensitive but case-preserving
- Unix-like file systems are typically case-sensitive
- Always use exact case matching for files and directories

### Reserved Filenames

Windows has reserved filenames (CON, PRN, AUX, etc.) that don't exist on Unix-like systems. Avoid these names in your code.

### Maximum Path Length

Windows has a shorter maximum path length than Unix-like systems. Keep paths reasonably short and use `filepath.Abs()` to get absolute paths.

## Reporting Cross-Platform Issues

If you encounter platform-specific issues:

1. Specify the platform and OS version
2. Provide exact steps to reproduce
3. Include relevant error messages and logs
4. If possible, test if the issue occurs on other platforms

## Platform-Specific Code

When you absolutely need platform-specific code, use build tags:

```go
//go:build windows
// +build windows

package mypackage

// Windows-specific code here
```

```go
//go:build !windows
// +build !windows

package mypackage

// Code for non-Windows platforms here
```