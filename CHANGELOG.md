# Changelog

All notable changes to the Globus Go CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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