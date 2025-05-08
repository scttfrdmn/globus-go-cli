# Bug Report: Missing Functions in SDK v0.9.11

## Description
In SDK v0.9.11, the functions `SetConnectionPoolManager` and `EnableDefaultConnectionPool` were removed, but are still referenced in `pkg/core/transport_init.go`. This causes compilation errors when building projects that depend on the SDK.

## Steps to Reproduce
1. Create a Go project that depends on `github.com/scttfrdmn/globus-go-sdk`
2. Update the dependency to `v0.9.11`
3. Attempt to build the project

## Error Message
```
# github.com/scttfrdmn/globus-go-sdk/pkg/core
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.11/pkg/core/transport_init.go:33:2: undefined: SetConnectionPoolManager
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.11/pkg/core/transport_init.go:37:3: undefined: EnableDefaultConnectionPool
```

## Expected Behavior
The project should compile successfully. Either the functions should be restored, or the references to them should be removed from `transport_init.go`.

## Actual Behavior
The compilation fails because the referenced functions are not defined anywhere.

## Investigation
The functions `SetConnectionPoolManager` and `EnableDefaultConnectionPool` were previously defined in `pkg/core/client_with_pool.go` in SDK v0.9.10, but this file was removed in v0.9.11. However, the references to these functions in `pkg/core/transport_init.go` were not updated.

## Environment
- Go version: 1.21+
- SDK version: v0.9.11
- OS: macOS, Linux

## Proposed Fix
Either:
1. Add the missing functions back to the codebase (simplest solution)
2. Update `transport_init.go` to remove the references to these functions
3. Move the connection pool initialization logic to a different location that doesn't require these specific functions

## Impact
This issue blocks any projects from upgrading to SDK v0.9.11.