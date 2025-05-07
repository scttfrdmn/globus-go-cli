# SDK Update Status Report

## Overview
This document summarizes the current status of our attempt to update the Globus Go CLI to use the latest Globus Go SDK versions.

## Current Status
- **Branch**: feature/sdk-0.9.10-update
- **Attempted SDK Version**: v0.9.10
- **Current SDK Version**: v0.9.10 (partial compatibility)
- **Status**: Partial success - package tests pass, but command implementations need updating

## Issues Identified

### 1. SDK v0.9.10 Build Progress
We have successfully updated to SDK v0.9.10, which fixes many of the previous issues:

#### Connection Pool Fix
- The v0.9.10 release fixes the previously reported issue with `httppool.NewHttpConnectionPoolManager`
- Our package tests now pass with the v0.9.10 SDK

#### Remaining API Compatibility Issues
We still have compatibility issues to resolve in the Auth package:

```
cmd/auth/device.go:89:32: authClient.GetDeviceCode undefined (type *"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth".Client has no field or method GetDeviceCode)
cmd/auth/device.go:110:31: authClient.PollDeviceCode undefined (type *"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth".Client has no field or method PollDeviceCode)
cmd/auth/identities.go:106:18: assignment mismatch: 1 variable but sdkConfig.NewAuthClient returns 2 values
cmd/auth/login.go:182:18: assignment mismatch: 2 variables but authClient.GetAuthorizationURL returns 1 value
cmd/auth/login.go:182:49: cannot use context.Background() (value of interface type context.Context) as string value in argument to authClient.GetAuthorizationURL
cmd/auth/login.go:182:78: cannot use scopes (variable of type []string) as string value in argument to authClient.GetAuthorizationURL
cmd/auth/tokens.go:201:48: introspection.Sub undefined (type *"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth".TokenInfo has no field or method Sub)
```

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
1. **Update Auth Package Implementation**: Refactor the auth-related commands to work with the new SDK:
   - Update method signatures for device flow authentication
   - Handle multi-return values from client initialization
   - Update token information structure field references
2. **Integration Testing**: After fixing auth package, perform integration testing with Globus services
3. **Incremental Deployment**: Consider deploying feature branches with partial functionality until all commands are updated
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
- Partial integration of v0.9.10: May 7, 2025

## References
- GitHub Issue #8: Import cycle issues
- GitHub Issue #9: SDK v0.9.5 build errors
- GitHub Issue #10: SDK v0.9.7/v0.9.8 build errors
- GitHub Issue #11: SDK v0.9.9 connection pool errors
- SDK Repository: https://github.com/scttfrdmn/globus-go-sdk