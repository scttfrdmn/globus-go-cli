# Changelog

All notable changes to the Globus Go CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [3.62.0-1] - 2025-08-09

### Changed
- Updated to Globus Go SDK v3.62.0-1
- Aligned with Python SDK v3.62.0 feature additions
- Maintained full backward compatibility with existing CLI functionality

### Added (SDK-level)
- Groups service subscription_id support
- SetSubscriptionAdminVerifiedID() method
- GetGroupSubscription() method  
- GroupSubscription type

### Fixed
- No code changes required - CLI benefits from enhanced Groups service features
- All tests pass with zero breaking changes
- Seamless upgrade from v3.61.0-1

## [3.61.0-1] - 2025-08-09

### Changed
- Updated to Globus Go SDK v3.61.0-1 
- Aligned with Python SDK v3.61.0 deprecation timeline
- Maintained full backward compatibility with existing CLI functionality

### Deprecated (SDK-level)
- Globus Connect Server v4 support deprecated in SDK
- ComputeClient alias deprecated in favor of compute.Client
- Legacy GCS v4 server methods deprecated

### Fixed
- No code changes required - CLI does not use deprecated APIs
- All tests pass with zero breaking changes
- Seamless upgrade from v3.60.0-1

## [3.60.0-1] - 2025-07-27

### Changed
- Updated to Globus Go SDK v3.60.0-1 with major version bump
- Migrated to v3 module path: github.com/scttfrdmn/globus-go-sdk/v3
- Aligned with Python SDK v3.60.0 versioning using hybrid format
- All SDK packages now marked as STABLE API

### Added
- Support for Globus Auth Requirements Error (GARE) for dependent consent handling
- Enhanced error handling matching Python SDK behavior
- Comprehensive stability indicators across all components
- Full Python SDK v3.x compatibility patterns

### Fixed
- Zero breaking changes - seamless migration from v0.9.17
- Maintained all existing CLI functionality and test coverage
- Preserved backward compatibility with all commands and options

## [0.9.17] - 2025-05-10

### Changed
- Updated to Globus Go SDK v0.9.17
- Successfully preserved compatibility with API stability changes
- Updated CLI version to match SDK version
- Significantly improved test coverage across all packages
- Enhanced cross-platform compatibility with explicit handling for Windows, macOS, and Linux

### Added
- Support for SDK stability indicators with clear component compatibility
- Improved error handling based on SDK v0.9.17 enhancements
- Comprehensive integration testing with real Globus credentials
- Proper mock implementations for all service clients
- Cross-platform test workflows in GitHub Actions
- Detailed documentation for integration testing setup
- Cross-platform compatibility guide for developers

### Fixed
- Maintained backwards compatibility with all SDK v0.9.15 functionality
- Fixed file path handling for cross-platform compatibility
- Improved test helpers for better test isolation
- Updated linting configuration to use staticcheck

## [0.9.15] - 2025-05-09

### Changed
- Updated to Globus Go SDK v0.9.15
- Successfully resolved SDK compatibility issues reported in GitHub issue #13
- Updated CLI version to match SDK version

### Fixed
- Fixed connection pool initialization issues with EnableDefaultConnectionPool function
- Maintained compatibility with all API changes from v0.9.10 to v0.9.15
- Ensured all tests pass with the updated SDK

## [0.9.10+1] - 2025-05-08

### Changed
- Maintained compatibility with Globus Go SDK v0.9.10
- Investigated compatibility issues with SDK v0.9.11, v0.9.12, and v0.9.13
- Created bug report for the SDK (Issue #13)

### Known Issues
- Unable to update to SDK v0.9.11-v0.9.14 due to persistent compilation errors
- Remaining on v0.9.10 until upstream SDK issues are properly resolved
- Despite v0.9.13 and v0.9.14 claiming to fix the issue, the problem persists
- Verified that v0.9.14 tag improperly points to the same commit as v0.9.11
- Reported detailed findings to upstream project for resolution

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