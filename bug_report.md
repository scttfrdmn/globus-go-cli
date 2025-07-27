# Bug Report: SDK v0.9.5 build errors preventing CLI update

## Description
When attempting to update the Globus Go CLI to use SDK v0.9.5, multiple compilation errors are encountered. These errors suggest that the SDK v0.9.5 contains incompatible API changes and possibly internal inconsistencies that prevent successful building.

## Build Environment
- Go version: (go version)
- SDK version: v0.9.5
- OS: macOS

## Error Details

The following errors occur when attempting to build with SDK v0.9.5:

### 1. Missing `VersionCheck` method/field in `Config` struct
```
# github.com/scttfrdmn/globus-go-sdk/pkg/core/config
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/core/config/api_version.go:11:7: c.VersionCheck undefined (type *Config has no field or method VersionCheck)
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/core/config/api_version.go:12:5: c.VersionCheck undefined (type *Config has no field or method VersionCheck)
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/core/config/api_version.go:16:5: c.VersionCheck undefined (type *Config has no field or method VersionCheck)
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/core/config/api_version.go:18:5: c.VersionCheck undefined (type *Config has no field or method VersionCheck)
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/core/config/api_version.go:26:7: c.VersionCheck undefined (type *Config has no field or method VersionCheck)
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/core/config/api_version.go:27:5: c.VersionCheck undefined (type *Config has no field or method VersionCheck)
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/core/config/api_version.go:31:4: c.VersionCheck undefined (type *Config has no field or method VersionCheck)
```

### 2. Auth package errors
```
# github.com/scttfrdmn/globus-go-sdk/pkg/services/auth
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/auth/mfa.go:65:6: declared and not used: mfaErr
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/auth/mfa.go:275:18: method Client.tokenRequest already declared at ../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/auth/client.go:129:18
```

### 3. Transfer package errors
```
# github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client.go:34:27: undefined: Option
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client.go:36:10: undefined: clientConfig
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:27:4: unknown field Owner in struct literal of type Task
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:37:4: unknown field SubmissionTime in struct literal of type Task
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:38:27: cannot use "" (untyped string constant) as *time.Time value in struct literal
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:46:20: undefined: setupMockServer
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:103:20: undefined: setupMockServer
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:174:20: undefined: setupMockServer
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:179:3: unknown field Filter in struct literal of type ListTasksOptions
../../go/pkg/mod/github.com/scttfrdmn/globus-go-sdk@v0.9.5/pkg/services/transfer/client_test_additions.go:240:20: undefined: setupMockServer
```

## Impact
These errors prevent the Globus Go CLI from being updated to use SDK v0.9.5, blocking our ability to release a new version of the CLI that incorporates the latest SDK improvements.

## Potential Causes
1. API changes without proper migration path
2. Missing type definitions or imports
3. Inconsistent interface implementations
4. Test code that doesn't match implementation code

## Suggested Fixes
1. Add the missing `VersionCheck` field/method to the `Config` struct
2. Fix the duplicate `tokenRequest` method declaration in auth package
3. Define the missing `Option` type and `clientConfig` variable
4. Update struct field definitions to match usage in tests
5. Add the missing `setupMockServer` function
6. Fix type conversions in test code

## Steps to Reproduce
1. Clone the Globus Go CLI repository
2. Update go.mod to use SDK v0.9.5
3. Run `make build`
4. Observe the compilation errors

I'd be happy to provide more details or assist with debugging if needed. This is blocking our CLI v0.9.5 release which was intended to incorporate the latest SDK improvements.