# Changelog

All notable changes to the Globus Go CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.9.10+1] - 2025-05-08

### Changed
- Maintained compatibility with Globus Go SDK v0.9.10
- Investigated compatibility issues with SDK v0.9.11, v0.9.12, and v0.9.13
- Created bug report for the SDK (Issue #13)

### Known Issues
- Unable to update to SDK v0.9.11-v0.9.13 due to persistent compilation errors
- Remaining on v0.9.10 until upstream SDK issues are properly resolved
- Despite v0.9.13 claiming to fix the issue, the problem persists

## [0.9.10] - 2025-05-07

### Changed
- Updated to Globus Go SDK v0.9.10
- Modified DeleteItem handling to use CreateDeleteTask instead of Delete
- Updated CLI version to match SDK version
- Refactored auth package for compatibility with SDK v0.9.10:
  - Updated refresh.go - Token refresh compatible with SDK v0.9.10
  - Updated tokens.go - Proper field references and client initialization
  - Updated whoami.go - Fixed Subject field reference (was Sub)
  - Updated logout.go - Updated for new client initialization pattern
  - Updated identities.go - Added temporary stub implementation
  - Updated device.go - Added placeholder implementation
- Refactored transfer package for compatibility with SDK v0.9.10:
  - Updated cp.go - Updated transfer client initialization and authorizer
  - Updated ls.go - Fixed field references from DATA to Data 
  - Updated endpoint.go - Simplified endpoint display formatting
  - Updated mkdir.go - Updated client initialization and authorizer
  - Updated rm.go - Replaced Delete with CreateDeleteTask
  - Updated task.go - Fixed time handling and field references

### Fixed
- Integration with SDK v0.9.10 which fixes connection pool initialization issues
- Improved compatibility with Globus API v0.10
- Fixed token introspection field references (Subject vs Sub)
- Fixed identity set handling in token introspection
- Fixed task time handling (RequestTime is now time.Time, CompletionTime is *time.Time)
- Fixed field name changes in transfer models (SourceEndpointDisplay vs SourceEndpointDisplayName)
- Updated CancelTask to correctly handle multiple return values
- Improved output formatting for tabular display

### Known Issues
- Device authentication flow implementation pending SDK support

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