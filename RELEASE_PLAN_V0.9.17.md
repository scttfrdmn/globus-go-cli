# Release Plan for Globus Go CLI v0.9.17

This document outlines the comprehensive plan for preparing and releasing version 0.9.17 of the Globus Go CLI, with a significant focus on improving test coverage and quality.

## Current Status

- SDK has been updated to v0.9.17
- CLI code is compatible with SDK v0.9.17
- Documentation has been updated to reflect SDK changes
- Test coverage is inadequate:
  - cmd: 29.7%
  - cmd/auth: 6.5%
  - cmd/transfer: 9.2%
  - pkg/config: 53.8%
  - pkg/output: 70.5%
  - pkg/testhelpers: 0%
- Multiple test failures present in the test suite
- Linting configuration needs to be updated

## Release Goals

1. Update SDK to v0.9.17 ✓
2. Increase test coverage to at least 80% across all packages
3. Fix all test failures
4. Ensure linting passes without errors
5. Complete cross-platform testing
6. Perform comprehensive integration testing
7. Release a stable v0.9.17 version

## Test Coverage Improvement Plan

### Phase 1: Test Infrastructure Enhancement

1. Review and enhance test mocks
   - Update mock implementations for auth and transfer services
   - Create consistent test fixtures for tokens and credentials
   - Implement test helper utilities for common operations

2. Fix test environment setup
   - Create proper test initialization that doesn't rely on user environment
   - Develop standardized test configuration for auth tokens
   - Implement test cleanup routines to ensure test isolation

### Phase 2: Fix Existing Tests

1. Repair broken tests in cmd/transfer
   - Address token file not found errors in ls_test.go
   - Fix mock expectations in mkdir_test.go
   - Correct command execution context in task_test.go

2. Update test assertions to match current output
   - Review and update all test output expectations
   - Ensure test assertions are robust against minor output changes

### Phase 3: Coverage Expansion

1. cmd/auth package (target: 80%+)
   - Add tests for login.go functionality
   - Implement comprehensive tests for tokens.go
   - Add tests for identity-related operations
   - Test refresh, logout, and whoami functionality thoroughly

2. cmd/transfer package (target: 80%+)
   - Expand tests for ls.go with various input cases
   - Add tests for cp.go with different transfer scenarios
   - Implement tests for mkdir.go with various path structures
   - Add comprehensive tests for rm.go including recursive operations
   - Enhance task.go tests covering all status cases

3. pkg modules (target: 90%+)
   - Expand config package tests
   - Add comprehensive formatter tests including edge cases
   - Implement tests for testhelpers package itself

### Phase 4: Edge Case and Error Handling Testing

1. Add negative test cases for all commands
   - Invalid input handling
   - API error responses
   - Network failure scenarios
   - Authentication failures

2. Test command option combinations
   - Verify all flag combinations work correctly
   - Test invalid flag combinations

## Linting and Code Quality

1. Update linting configuration
   - Update .golangci.yml to include Go version field ✓
   - Configure to use staticcheck for code validation ✓
   - Remove deprecated linters and use a minimal set of reliable linters ✓

2. Perform comprehensive linting passes
   - Fix all identified lint issues
   - Standardize code style across the codebase

## Integration Testing

1. Set up test credentials for integration testing
   - Create dedicated test endpoints and resources
   - Configure credential storage for automated testing
   - **CRITICAL REQUIREMENT**: Test with real Globus credentials before release

2. Implement automated integration tests
   - Test auth workflows against real Globus Auth service
   - Verify transfer operations with real endpoints
   - Test task management with actual transfer tasks
   - Ensure all integration tests pass with real-world credentials

3. Document integration test setup
   - Create detailed instructions for setting up test environment
   - Document required credentials and permissions
   - Provide guide for troubleshooting integration test failures

## Cross-Platform Verification

1. Test on multiple platforms
   - Linux (Ubuntu LTS)
   - macOS (latest)
   - Windows 10/11

2. Verify installation methods
   - Direct binary installation
   - Package manager installation (Homebrew, etc.)
   - Build from source

## Documentation and Release Prep

1. Update documentation
   - Ensure README reflects latest capabilities
   - Update command examples with v0.9.17 features
   - Verify all help text is accurate

2. Prepare changelog
   - Document all changes since v0.9.15
   - Highlight test coverage improvements
   - Note any behavior changes

## Release Process

1. Final verification
   - Run full test suite with coverage report
   - Verify coverage meets 80%+ target
   - Ensure all linting passes
   - Run integration tests

2. Create release
   - Tag v0.9.17
   - Build release binaries
   - Update package manager formulas
   - Publish release notes

## Timeline and Milestones

| Phase | Tasks | Estimated Time | Target Completion |
|-------|-------|----------------|-------------------|
| Infrastructure | Mocks and fixtures | 1 week | Week 1 |
| Fix existing tests | Repair failing tests | 1 week | Week 2 |
| Auth coverage | Increase to 80%+ | 1.5 weeks | Week 3-4 |
| Transfer coverage | Increase to 80%+ | 2 weeks | Week 4-5 |
| Edge cases | Add negative tests | 1 week | Week 6 |
| Linting | Fix all issues | 0.5 week | Week 6 |
| Integration | Perform real-world tests | 1 week | Week 7 |
| Documentation | Update all docs | 0.5 week | Week 7 |
| Release | Final verification and release | 0.5 week | Week 8 |

Total estimated timeline: 8 weeks

## Success Criteria

1. Test coverage above 80% for all packages
2. All tests passing on all platforms
3. No linting errors with staticcheck
4. **REQUIRED**: Successful integration tests with real Globus credentials
5. Well-documented release with clear changelog
6. All functional tests verified against actual Globus services

## Tracking

Progress on this release plan will be tracked in the project management system with weekly status updates.