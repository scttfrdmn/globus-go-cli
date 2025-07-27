# Comprehensive Testing Plan for globus-go-cli

This document outlines the comprehensive testing strategy for the globus-go-cli project, including goals, implementation phases, and specific test categories.

## 1. Testing Goals

### Coverage Targets
- **Initial Target**: 70% overall code coverage, 85% for critical packages (auth, transfer)
- **Long-term Target**: 80%+ overall with 90%+ for core functionality
- **Priority Areas**: Error handling, CLI argument parsing, token management, API interactions

### Test Categories
- **Unit Tests**: Testing individual components with mocked dependencies
- **Command Tests**: Testing CLI commands with input/output validation
- **Integration Tests**: End-to-end tests with actual Globus services
- **Mock Server Tests**: HTTP response simulation for edge cases and error scenarios
- **Benchmark Tests**: Performance validation for critical operations

## 2. Implementation Plan

### Phase 1: Core Unit Tests (1-2 weeks)
- Create test utilities and mocks
- Implement comprehensive tests for token handling and authentication
- Test configuration management
- Add tests for core utility functions

### Phase 2: Command Tests (2-3 weeks)
- Create mock client implementations
- Test CLI command structure and argument parsing
- Verify proper error handling and user feedback
- Test all command variations and flag combinations

### Phase 3: Integration Tests (2-3 weeks)
- Complete the skeleton integration tests
- Add CI-compatible test credential management
- Implement conditional test execution based on credential availability
- Create isolated test endpoints for file operations

### Phase 4: CI Integration (1 week)
- Configure test automation in CI
- Set up coverage reporting and thresholds
- Create test matrix (Go versions, OS platforms)

## 3. Test Infrastructure

### Test Helpers

```go
// pkg/testhelpers/output.go
package testhelpers

import (
    "bytes"
    "io"
    "os"
)

// CaptureOutput captures stdout and stderr during the execution of a function
// Returns captured stdout and stderr as strings
func CaptureOutput(f func()) (string, string) {
    oldStdout, oldStderr := os.Stdout, os.Stderr
    rOut, wOut, _ := os.Pipe()
    rErr, wErr, _ := os.Pipe()
    os.Stdout, os.Stderr = wOut, wErr

    outC, errC := make(chan string), make(chan string)
    go func() {
        var buf bytes.Buffer
        io.Copy(&buf, rOut)
        outC <- buf.String()
    }()
    go func() {
        var buf bytes.Buffer
        io.Copy(&buf, rErr)
        errC <- buf.String()
    }()

    f()

    wOut.Close()
    wErr.Close()
    os.Stdout, os.Stderr = oldStdout, oldStderr
    return <-outC, <-errC
}
```

### Mock Clients

```go
// pkg/testhelpers/mocks/transfer.go
package mocks

import (
    "context"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// MockTransferClient implements a mock transfer client for testing
type MockTransferClient struct {
    // Function fields for mocking responses
    ListEndpointsFunc func(ctx context.Context, options *transfer.ListEndpointsOptions) (*transfer.EndpointList, error)
    GetTaskFunc       func(ctx context.Context, taskID string) (*transfer.Task, error)
    SubmitTransferFunc func(ctx context.Context, sourceEndpointID, sourcePath, 
                             destEndpointID, destPath, label string, 
                             options map[string]interface{}) (*transfer.TaskResponse, error)
    ListDirectoryFunc func(ctx context.Context, options *transfer.ListDirectoryOptions) (*transfer.ListDirectoryResponse, error)
    CreateDirectoryFunc func(ctx context.Context, options *transfer.CreateDirectoryOptions) error
    CancelTaskFunc func(ctx context.Context, taskID string) (*transfer.OperationResult, error)
    GetEndpointFunc func(ctx context.Context, endpointID string) (*transfer.Endpoint, error)
    ListTasksFunc func(ctx context.Context, options *transfer.ListTasksOptions) (*transfer.TaskList, error)
    CreateDeleteTaskFunc func(ctx context.Context, request *transfer.DeleteTaskRequest) (*transfer.TaskResponse, error)
}

// Implement all the interface methods to use the function fields
func (m *MockTransferClient) ListEndpoints(ctx context.Context, options *transfer.ListEndpointsOptions) (*transfer.EndpointList, error) {
    if m.ListEndpointsFunc != nil {
        return m.ListEndpointsFunc(ctx, options)
    }
    return &transfer.EndpointList{}, nil
}

// Implement remaining interface methods...
```

### Integration Test Utils

```go
// pkg/testhelpers/integration.go
package testhelpers

import (
    "testing"
    "os"
    "path/filepath"
)

// SkipIfNoCredentials skips tests that require credentials if they're not available
func SkipIfNoCredentials(t *testing.T) *TestCredentials {
    creds := LoadTestCredentials(t)
    if creds.ClientID == "" || creds.ClientSecret == "" {
        t.Skip("Skipping test: No credentials available")
    }
    return creds
}

// CreateTemporaryTestFiles creates temporary test files for transfer tests
func CreateTemporaryTestFiles(t *testing.T, fileCount int) (dir string, cleanup func()) {
    dir, err := os.MkdirTemp("", "globus-test-")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    
    for i := 0; i < fileCount; i++ {
        filename := filepath.Join(dir, fmt.Sprintf("test-file-%d", i))
        content := []byte(fmt.Sprintf("Test content for file %d", i))
        if err := os.WriteFile(filename, content, 0644); err != nil {
            os.RemoveAll(dir) // Clean up before failing
            t.Fatalf("Failed to create test file: %v", err)
        }
    }
    
    return dir, func() { os.RemoveAll(dir) }
}
```

## 4. Test Categories and Examples

### Unit Tests

Example unit test for token validation:

```go
func TestTokenValidation(t *testing.T) {
    tests := []struct {
        name        string
        token       *auth.TokenInfo
        expectValid bool
    }{
        {
            name: "valid token",
            token: &auth.TokenInfo{
                AccessToken: "valid-token",
                ExpiresAt:   time.Now().Add(1 * time.Hour),
            },
            expectValid: true,
        },
        {
            name: "expired token",
            token: &auth.TokenInfo{
                AccessToken: "expired-token",
                ExpiresAt:   time.Now().Add(-1 * time.Hour),
            },
            expectValid: false,
        },
        {
            name:        "nil token",
            token:       nil,
            expectValid: false,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            valid := auth.IsTokenValid(tc.token)
            if valid != tc.expectValid {
                t.Errorf("IsTokenValid() = %v, want %v", valid, tc.expectValid)
            }
        })
    }
}
```

### Command Tests

Example command test for `globus transfer ls`:

```go
func TestTransferLsCommand(t *testing.T) {
    tests := []struct {
        name          string
        args          []string
        mockSetup     func(*mocks.MockTransferClient)
        expectOutput  string
        expectError   bool
    }{
        {
            name: "successful directory listing",
            args: []string{"endpoint-id:/path"},
            mockSetup: func(m *mocks.MockTransferClient) {
                m.ListDirectoryFunc = func(ctx context.Context, options *transfer.ListDirectoryOptions) (*transfer.ListDirectoryResponse, error) {
                    return &transfer.ListDirectoryResponse{
                        Path: "/path",
                        Data: []transfer.FileEntry{
                            {Name: "file1", Type: "file", Size: 1024, LastModified: "2023-01-01T12:00:00Z"},
                            {Name: "dir1", Type: "dir", Size: 0, LastModified: "2023-01-01T12:00:00Z"},
                        },
                    }, nil
                }
            },
            expectOutput: "file1",
            expectError:  false,
        },
        {
            name: "error listing directory",
            args: []string{"invalid-endpoint:/path"},
            mockSetup: func(m *mocks.MockTransferClient) {
                m.ListDirectoryFunc = func(ctx context.Context, options *transfer.ListDirectoryOptions) (*transfer.ListDirectoryResponse, error) {
                    return nil, fmt.Errorf("endpoint not found")
                }
            },
            expectOutput: "",
            expectError:  true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            // Create mock client
            mockClient := &mocks.MockTransferClient{}
            if tc.mockSetup != nil {
                tc.mockSetup(mockClient)
            }
            
            // Create command
            cmd := transfer.LsCmd()
            cmd.SetArgs(tc.args)
            
            // Capture output
            stdout, stderr := CaptureOutput(func() {
                err := cmd.Execute()
                if tc.expectError && err == nil {
                    t.Errorf("Expected error but got none")
                } else if !tc.expectError && err != nil {
                    t.Errorf("Unexpected error: %v", err)
                }
            })
            
            // Check output
            if tc.expectOutput != "" && !strings.Contains(stdout, tc.expectOutput) {
                t.Errorf("Expected output to contain %q, got stdout: %q, stderr: %q", 
                         tc.expectOutput, stdout, stderr)
            }
        })
    }
}
```

### Integration Tests

Example integration test for file transfer:

```go
func TestEndToEndFileTransfer(t *testing.T) {
    // Skip if credentials not available
    creds := SkipIfNoCredentials(t)
    creds.RequireTransferEndpoints(t)
    
    // Create temporary test files
    sourceDir, sourceCleanup := CreateTemporaryTestFiles(t, 3)
    defer sourceCleanup()
    
    // Setup transfer command
    cmd := transfer.CpCmd()
    cmd.SetArgs([]string{
        "--recursive",
        creds.SourceEndpoint + ":" + sourceDir,
        creds.DestinationEndpoint + ":" + creds.DestinationPath,
    })
    
    // Execute transfer and capture task ID
    var taskID string
    stdout, _ := CaptureOutput(func() {
        err := cmd.Execute()
        if err != nil {
            t.Fatalf("Failed to execute transfer: %v", err)
        }
    })
    
    // Extract task ID from output
    for _, line := range strings.Split(stdout, "\n") {
        if strings.HasPrefix(line, "Task ID:") {
            taskID = strings.TrimSpace(strings.TrimPrefix(line, "Task ID:"))
            break
        }
    }
    
    if taskID == "" {
        t.Fatal("Failed to get task ID from output")
    }
    
    // Wait for transfer to complete
    waitCmd := transfer.TaskWaitCmd()
    waitCmd.SetArgs([]string{taskID, "--timeout", "300"})
    
    stdout, _ = CaptureOutput(func() {
        err := waitCmd.Execute()
        if err != nil {
            t.Fatalf("Failed to wait for transfer: %v", err)
        }
    })
    
    // Verify transfer succeeded
    if !strings.Contains(stdout, "succeeded") {
        t.Errorf("Transfer did not succeed, output: %s", stdout)
    }
}
```

## 5. CI Integration

The CI process should include:

1. Running all unit tests
2. Running integration tests if credentials are available
3. Calculating and reporting code coverage
4. Setting up matrix testing across multiple Go versions

Example GitHub Actions workflow:

```yaml
name: Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21.x, 1.22.x]

    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests with coverage
      run: go test -v -coverprofile=coverage.out -covermode=atomic ./...
    
    - name: Upload coverage report
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        fail_ci_if_error: false
```

## 6. Implementation Timeline

| Phase | Timeline | Tasks | Status |
|-------|----------|-------|--------|
| 1     | Weeks 1-2 | Core unit tests and test infrastructure | Not started |
| 2     | Weeks 3-5 | Command tests and mocks | Not started |
| 3     | Weeks 6-8 | Integration tests | Not started |
| 4     | Week 9    | CI integration | Not started |

## 7. Success Metrics

- Code coverage percentage meets or exceeds targets
- All critical paths have test coverage
- CI pipeline consistently passes
- Integration tests validate end-to-end functionality
- New PRs must include tests for new functionality