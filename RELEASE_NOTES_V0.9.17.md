# Globus Go CLI v0.9.17 Release Notes

## Overview

Globus Go CLI v0.9.17 is a significant maintenance release that updates the CLI to use the latest Globus Go SDK v0.9.17. This release focuses on improving stability, test coverage, and cross-platform compatibility.

## Compatibility

- Requires Go 1.20 or newer
- Compatible with Globus Go SDK v0.9.17
- Supports Windows, macOS, and Linux platforms

## Key Features

### SDK Update

- Updated to Globus Go SDK v0.9.17
- Successfully preserved compatibility with all API stability changes
- Leveraged improved error handling from the SDK
- Incorporated SDK stability indicators for better component compatibility

### Testing Improvements

- Significantly increased test coverage across all packages
- Added comprehensive integration testing with real Globus credentials
- Implemented proper mock clients for all Globus services
- Created test helpers for better test isolation and reliability
- Added GitHub Actions workflow for cross-platform testing

### Documentation

- Added comprehensive integration testing documentation
- Created cross-platform compatibility guide for developers
- Updated SDK compatibility status documentation
- Enhanced release process documentation

### Cross-Platform Compatibility

- Improved file path handling for better cross-platform compatibility
- Enhanced Windows support through proper path handling
- Verified functionality on Windows, macOS, and Linux through CI/CD
- Added explicit platform detection for browser launching

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/scttfrdmn/globus-go-cli.git
cd globus-go-cli

# Build the CLI
go build -o globus .

# Install to your PATH
mv globus /usr/local/bin/  # Unix/Linux/macOS
# or
move globus %USERPROFILE%\bin\  # Windows
```

### Using Go Install

```bash
go install github.com/scttfrdmn/globus-go-cli@v0.9.17
```

## Setting Up Integration Testing

For developers who want to run integration tests, we've added comprehensive documentation in the new `INTEGRATION_TESTING.md` file. This document explains how to:

1. Create a Globus App for testing
2. Set up test credentials in the `.env.test` file
3. Configure endpoints for transfer testing
4. Run integration tests with real credentials

## Known Issues

- Device authentication flow implementation is still using a placeholder until SDK support is available
- Some advanced transfer features may require newer SDK versions

## What's Next

Our roadmap for future releases includes:

1. Full device authentication flow implementation once SDK support is available
2. Additional command-line utilities for common Globus operations
3. Improved performance for large file transfers
4. Enhanced documentation with examples for common use cases

## Acknowledgments

Special thanks to:
- The Globus team for continued API improvements
- All contributors who provided feedback and bug reports
- Users who participated in testing and validation