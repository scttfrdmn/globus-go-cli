# Changelog

All notable changes to the Globus Go CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.10] - 2025-05-07 (In Progress)

### Changed
- Updated to Globus Go SDK v0.9.10
- Modified DeleteItem handling to be compatible with newer SDK versions
- Updated CLI version to match SDK version
- Created temporary stubs for device authentication while waiting for SDK implementation

### Fixed
- Integration with SDK v0.9.10 which fixes connection pool initialization issues
- Improved compatibility with Globus API v0.10

### Known Issues
- Auth API compatibility issues need to be resolved:
  - Device authentication flow has been substantially changed in the SDK
  - Auth client initialization now returns multiple values
  - Token introspection fields have been renamed
- Some command implementations need to be updated for the new SDK:
  - auth/device.go: Device auth flow needs proper implementation
  - auth/login.go: GetAuthorizationURL method signature changed
  - auth/tokens.go: Token field references need updating
- Partial build success achieved with package tests passing

## [0.9.1] - 2025-05-07

### Changed
- Maintained Globus Go SDK v0.9.1 due to compatibility issues with newer SDK versions
- Submitted bug reports for SDK v0.9.5 issues (github.com/scttfrdmn/globus-go-sdk/issues/9)
- Submitted bug reports for SDK v0.9.6 issues (github.com/scttfrdmn/globus-go-sdk/issues/10)
- Improved code reliability and stability

### Known Issues
- Unable to update to SDK v0.9.7 due to compilation errors
- Import cycle issues in SDK affecting all versions
- Multiple bug reports submitted (github.com/scttfrdmn/globus-go-sdk/issues/8, /issues/9, and /issues/10)
- Waiting for SDK fixes before proceeding with CLI update

## [0.1.0] - 2023-05-01

### Added
- Initial project setup
- Basic CLI framework with Cobra and Viper
- Documentation structure