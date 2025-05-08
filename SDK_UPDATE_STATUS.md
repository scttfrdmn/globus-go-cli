# SDK Update Status Report

## Overview
This document summarizes the current status of our attempt to update the Globus Go CLI to use the latest Globus Go SDK versions.

## Current Status
- **Branch**: feature/sdk-0.9.10-compatibility
- **SDK Version**: v0.9.10
- **Status**: Success - All tests pass and CLI is functional with SDK v0.9.10

### SDK v0.9.11 Issues
Attempted to update to SDK v0.9.11 but discovered breaking changes:
- Missing functions in the SDK's core package that are still referenced internally
- Created GitHub issue #13 (https://github.com/scttfrdmn/globus-go-sdk/issues/13) to report the problem
- Maintaining v0.9.10 compatibility until the SDK bug is fixed

## Issues Identified

### 1. SDK v0.9.10 Build Progress
We have successfully updated to SDK v0.9.10, which fixes many of the previous issues:

#### Connection Pool Fix
- The v0.9.10 release fixes the previously reported issue with `httppool.NewHttpConnectionPoolManager`
- Our package tests now pass with the v0.9.10 SDK

#### Progress and Remaining Issues

### Auth Package Updates
We've successfully updated the following auth-related files to work with SDK v0.9.10:
- refresh.go - Updated client initialization and token refresh
- tokens.go - Updated field references and client initialization
- whoami.go - Updated field references (Subject instead of Sub, IdentitySet instead of IdentitiesSets)
- logout.go - Updated client initialization
- identities.go - Updated with a temporary stub implementation
- device.go - Added placeholder implementation pending SDK device flow support

### Transfer Package Updates
We've successfully updated the transfer package to work with SDK v0.9.10:

- Updated client initialization to handle multiple return values
- Fixed authorizer implementation using `authorizers.ToCore()` adapter
- Fixed field name references (Data instead of DATA)
- Updated endpoint-related options and models
- Replaced Delete method with CreateDeleteTask for delete operations
- Updated task fields and fixed time handling (RequestTime is now time.Time, CompletionTime is *time.Time)
- Fixed tabular display output formatting

### 2. API Changes Between SDK Versions
The Auth API has undergone significant changes in v0.9.10:
- Method signatures have changed (e.g., GetDeviceCode, PollDeviceCode, GetAuthorizationURL)
- Constructor functions now return multiple values
- Token structure has changed with field renames

The DeleteItem method no longer takes a Recursive field parameter, which we've already updated in our code.

### 3. Import Cycle Issues
When attempting to build with SDK v0.9.1:
```
imports github.com/scttfrdmn/globus-go-sdk/pkg/core/transport from client_with_pool.go
imports github.com/scttfrdmn/globus-go-sdk/pkg/core from transport.go: import cycle not allowed
```
This indicates a circular dependency in the SDK that affects all versions.

## Actions Taken
1. Created GitHub issue #9 for SDK v0.9.5 build errors (https://github.com/scttfrdmn/globus-go-sdk/issues/9)
2. Created GitHub issue #10 for SDK v0.9.7 build errors (https://github.com/scttfrdmn/globus-go-sdk/issues/10)
3. Created GitHub issue #11 for SDK v0.9.9 build errors (https://github.com/scttfrdmn/globus-go-sdk/issues/11)
4. Updated the go.mod file to use SDK v0.9.10 which addresses our latest bug report
5. Updated the CLI code to accommodate the removal of the Recursive field in DeleteItem
6. Successfully built and tested the internal packages (pkg/*)
7. Updated CLI version to v0.9.10 to match SDK version
8. Updated CHANGELOG.md and SDK_UPDATE_STATUS.md to document progress and remaining issues

## Next Steps
1. **Finalize Auth Implementation**: Complete the device flow implementation when SDK support is available
2. **Integration Testing**: Perform integration testing with Globus services
3. **Release Preparation**: Finalize documentation and release v0.9.10
## Recommendations
1. **Create Feature Branch for Auth Updates**: Create a dedicated branch for updating auth-related commands
2. **Documentation**: Document the API changes in a developer guide for future reference
3. **Parallel Development**: Continue development with SDK v0.9.10 in feature branches while maintaining v0.9.1 compatibility in the main branch until all commands are updated

## Timeline
- Issues with v0.9.5 reported: May 7, 2025
- Issues with v0.9.6 reported: May 7, 2025 (after v0.9.6 release)
- Issues with v0.9.7 reported: May 7, 2025 (after v0.9.7 release)
- Issues with v0.9.9 reported: May 7, 2025 (after v0.9.9 release)
- SDK v0.9.10 released with fixes: May 7, 2025
- Full integration of v0.9.10: May 7, 2025
- SDK v0.9.11 released: May 8, 2025
- Issues with v0.9.11 reported: May 8, 2025 (GitHub issue #13)

## References
- GitHub Issue #8: Import cycle issues
- GitHub Issue #9: SDK v0.9.5 build errors
- GitHub Issue #10: SDK v0.9.7/v0.9.8 build errors
- GitHub Issue #11: SDK v0.9.9 connection pool errors
- SDK Repository: https://github.com/scttfrdmn/globus-go-sdk