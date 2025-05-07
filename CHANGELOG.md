# Changelog

All notable changes to the Globus Go CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.3] - 2025-05-07

### Changed
- Updated CLI version to 0.9.3
- Attempted to use SDK v0.9.3 but encountered compilation issues (see SDK issues #6 and #7)
- Maintained dependency on SDK v0.9.1 due to issues in the newer versions
- Improved version tracking and documentation

### Known Issues
- Import cycle identified in the Globus Go SDK (see SDK issue #8)
- Build issues with all SDK versions (v0.9.1 to v0.9.3)
- Awaiting resolution of SDK issues before further updates

## [0.9.2] - 2025-05-07

### Changed
- Updated to include Globus Go SDK v0.9.1 (SDK v0.9.2 introduces breaking interface changes)
- Improved code consistency and documentation
- Fixed configuration loading in test environments

### Added
- Complete Auth service functionality
  - Login command with web browser flow
  - Login command with device code flow
  - Token refresh and revocation
  - Identity lookup
- Basic Transfer service functionality
  - Endpoint listing and search
  - Directory listing
  - File and directory creation/removal
  - File transfers between endpoints
  - Task management (list, show, cancel, wait)
- Multiple output formats (text, JSON, CSV)
- Configuration profiles
- Shell completions for Bash, Zsh, Fish, and PowerShell
- CI/CD pipeline for automated builds and releases
- Homebrew formula for macOS distribution
- Docker image for containerized usage
- Cross-platform support (Linux, macOS, Windows)

### Changed
- Initial project structure based on modern Go practices
- Comprehensive Makefile with build, test, and lint targets

## [0.1.0] - 2023-05-01

### Added
- Initial project setup
- Basic CLI framework with Cobra and Viper
- Documentation structure