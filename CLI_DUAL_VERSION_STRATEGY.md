<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# CLI Dual-Version Strategy: Parallel v3/v4 Development

**Created:** 2025-12-09
**Status:** 📋 PROPOSAL - For future consideration
**Approach:** Adopt SDK's dual-version model for CLI

---

## Executive Summary

Adopt the same dual-version approach as the SDK, maintaining both v3 and v4 CLI implementations in parallel. This provides a low-risk migration path when upstream Globus v4 becomes official.

**Key Benefits:**
- ✅ Low-risk gradual migration (like SDK)
- ✅ Users choose when to migrate
- ✅ v3 stays stable during v4 development
- ✅ Can validate v4 without breaking v3
- ✅ Supports both SDK versions in same repo

---

## Current State vs Proposed State

### Current Approach
```
globus-go-cli/
├── cmd/              # All commands use SDK v3
├── pkg/              # Shared utilities
└── go.mod            # Single SDK v3 dependency
```

**Issues with current approach:**
- When v4 becomes official, we must migrate everything at once
- No gradual transition period
- High risk, large-scale refactoring
- Users forced to migrate immediately

### Proposed Approach (Parallel Development)
```
globus-go-cli/
├── cmd/
│   ├── v3/           # v3 implementation (current code)
│   └── v4/           # v4 implementation (new)
├── pkg/
│   ├── shared/       # Common code (config, output, auth)
│   ├── v3/           # v3-specific utilities
│   └── v4/           # v4-specific utilities
├── main.go           # Entry point (dispatches to v3 or v4)
├── go.mod            # Both SDK v3 and v4 as dependencies
└── go.work           # Workspace for both versions
```

**Benefits:**
- ✅ Gradual command-by-command migration
- ✅ Both versions work simultaneously
- ✅ Easy rollback if issues found
- ✅ Mirrors SDK's proven approach

---

## Architecture Options

### Option A: Single Binary with Version Flag (Recommended)

**Binary:** `globus` (single binary)
**Usage:**
```bash
# Default: uses v3 (during transition)
globus transfer ls ENDPOINT:/path

# Explicit v3
globus --sdk-version=v3 transfer ls ENDPOINT:/path

# Explicit v4
globus --sdk-version=v4 transfer ls ENDPOINT:/path

# Environment variable
GLOBUS_SDK_VERSION=v4 globus transfer ls ENDPOINT:/path
```

**Pros:**
- Single binary to distribute
- Easy for users to test v4
- Gradual migration with fallback
- No changes to existing scripts (defaults to v3)

**Cons:**
- Slightly larger binary (includes both v3 and v4)
- Need routing logic in main.go

### Option B: Separate Binaries

**Binaries:** `globus-v3` and `globus-v4` (or `globus` and `globus-next`)
**Usage:**
```bash
# v3 version
globus-v3 transfer ls ENDPOINT:/path

# v4 version
globus-v4 transfer ls ENDPOINT:/path

# Or: globus (v3) and globus-next (v4)
globus transfer ls ENDPOINT:/path        # v3
globus-next transfer ls ENDPOINT:/path   # v4
```

**Pros:**
- Clean separation
- Smaller individual binaries
- Clear which version you're using

**Cons:**
- Need to install/update two binaries
- Package managers need both
- More confusing for users

### Option C: Build Tags

**Binary:** `globus` (built with v3 or v4)
**Usage:**
```bash
# Build v3 version
go build -tags v3 -o globus-v3

# Build v4 version
go build -tags v4 -o globus-v4
```

**Pros:**
- No runtime overhead
- Clear separation at build time

**Cons:**
- Cannot have both versions in one binary
- Harder for users to switch/test
- More complex release process

---

## Recommended Approach: Option A

**Single binary with `--sdk-version` flag** provides the best user experience and mirrors the SDK's "use v3 in production, test v4" philosophy.

---

## Implementation Plan

### Phase 1: Restructure Repository (2-3 weeks)

**Goal:** Set up dual-version structure without breaking existing v3

#### Week 1: Directory Restructuring
1. **Create new directory structure:**
   ```bash
   mkdir -p cmd/v3 cmd/v4 pkg/shared pkg/v3 pkg/v4
   ```

2. **Move existing v3 commands:**
   ```bash
   # Move all current cmd/* to cmd/v3/*
   mv cmd/auth cmd/v3/
   mv cmd/transfer cmd/v3/
   mv cmd/group cmd/v3/
   mv cmd/timer cmd/v3/
   mv cmd/search cmd/v3/
   mv cmd/flows cmd/v3/
   mv cmd/compute cmd/v3/
   ```

3. **Extract shared code to pkg/shared:**
   - `pkg/shared/config` - Configuration (works with both versions)
   - `pkg/shared/output` - Output formatters (both versions)
   - `pkg/shared/auth` - Common auth patterns
   - `pkg/shared/common` - Shared utilities

4. **Move v3-specific code to pkg/v3:**
   - Anything that directly uses SDK v3 APIs
   - v3-specific error handling
   - v3-specific client creation

#### Week 2: Update Imports and Build System
1. **Update all import paths:**
   ```go
   // Before
   import "github.com/scttfrdmn/globus-go-cli/pkg/config"

   // After (v3)
   import "github.com/scttfrdmn/globus-go-cli/pkg/shared/config"
   import "github.com/scttfrdmn/globus-go-cli/pkg/v3/client"
   ```

2. **Update go.mod to include both SDKs:**
   ```go
   require (
       github.com/scttfrdmn/globus-go-sdk/v3 v3.65.0-1
       github.com/scttfrdmn/globus-go-sdk/v4 v4.2.0-1
       // ... other deps
   )
   ```

3. **Create main.go dispatcher:**
   ```go
   package main

   import (
       "os"
       v3cmd "github.com/scttfrdmn/globus-go-cli/cmd/v3"
       v4cmd "github.com/scttfrdmn/globus-go-cli/cmd/v4"
   )

   func main() {
       version := os.Getenv("GLOBUS_SDK_VERSION")
       if version == "" {
           version = "v3" // Default during transition
       }

       switch version {
       case "v4":
           v4cmd.Execute()
       default:
           v3cmd.Execute()
       }
   }
   ```

4. **Add version flag to root command:**
   ```go
   rootCmd.PersistentFlags().String("sdk-version", "v3", "SDK version to use (v3 or v4)")
   ```

#### Week 3: Testing and Validation
1. **Verify all v3 commands still work**
2. **Update CI/CD to test both paths**
3. **Update documentation**
4. **Create migration guide**

### Phase 2: Implement v4 Commands (Gradual, On-Demand)

**Goal:** Implement v4 versions of commands as they're needed or requested

#### Prioritization Strategy
Implement v4 commands based on:
1. **User demand** - Commands users ask for in v4
2. **v4-exclusive features** - Commands that benefit from v4 features
3. **Complexity** - Start with simpler commands

#### Implementation Pattern (per service)
1. **Create v4 command structure:**
   ```bash
   mkdir cmd/v4/transfer
   ```

2. **Implement using SDK v4:**
   ```go
   // cmd/v4/transfer/ls.go
   package transfer

   import (
       "context"
       "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
       "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
       "github.com/scttfrdmn/globus-go-cli/pkg/shared/config"
       "github.com/scttfrdmn/globus-go-cli/pkg/shared/output"
   )

   func NewLsCommand() *cobra.Command {
       cmd := &cobra.Command{
           Use:   "ls ENDPOINT_ID:PATH",
           Short: "List directory contents (SDK v4)",
           RunE: func(cmd *cobra.Command, args []string) error {
               ctx := context.Background()

               // Use SDK v4 patterns
               config := &core.Config{
                   AccessToken: getToken(),
                   Scopes:      []string{core.Scopes.TransferAll},
               }

               client, err := transfer.NewClient(ctx, config)
               if err != nil {
                   return err
               }
               defer client.Close() // v4 feature!

               // ... implementation
               return nil
           },
       }
       return cmd
   }
   ```

3. **Add comprehensive tests**
4. **Update documentation**
5. **Release incrementally**

#### Migration Sequence Example
**Phase 2a: Core Commands (8-12 weeks)**
- Auth commands (login, logout, whoami) - Week 1-2
- Transfer ls, mkdir - Week 3-4
- Transfer cp (basic) - Week 5-6
- Transfer task commands - Week 7-8
- Groups list, show - Week 9-10
- Integration testing - Week 11-12

**Phase 2b: Advanced Commands (12-16 weeks)**
- Remaining Transfer commands
- Search commands
- Flows commands
- Timers commands
- Compute commands

**Phase 2c: Polish and Migration Tools (4 weeks)**
- Migration documentation
- v3 deprecation warnings
- v4 default switch
- Final testing

### Phase 3: Transition Period (6-12 months)

**Goal:** Support both versions, encourage v4 adoption

#### Month 1-3: v4 Beta
- ✅ v3 is default
- ✅ v4 available with `--sdk-version=v4`
- ✅ Documentation shows both versions
- ✅ Gather user feedback

#### Month 4-6: v4 Release Candidate
- ✅ All commands available in v4
- ✅ Production testing with early adopters
- ✅ Performance comparison v3 vs v4
- ✅ Fix any v4 issues

#### Month 7-9: v4 Default
- ⚠️ v4 becomes default
- ⚠️ v3 still available with `--sdk-version=v3`
- ⚠️ Deprecation warnings for v3
- ⚠️ Migration guides published

#### Month 10-12: v3 Deprecation
- ⚠️ v3 marked deprecated
- ⚠️ Announce v3 removal timeline
- ⚠️ Final migration window

### Phase 4: v3 Removal (After 12 months)

**Goal:** Remove v3 code, v4 only

1. Remove `cmd/v3/` directory
2. Remove `pkg/v3/` directory
3. Remove SDK v3 dependency
4. Simplify main.go (no dispatcher)
5. Release v4.0.0 (major version bump)

---

## Directory Structure Detail

### Final Structure
```
globus-go-cli/
├── cmd/
│   ├── v3/                    # v3 implementation
│   │   ├── auth/              # Auth commands (SDK v3)
│   │   ├── transfer/          # Transfer commands (SDK v3)
│   │   ├── group/             # Group commands (SDK v3)
│   │   ├── timer/             # Timer commands (SDK v3)
│   │   ├── search/            # Search commands (SDK v3)
│   │   ├── flows/             # Flows commands (SDK v3)
│   │   ├── compute/           # Compute commands (SDK v3)
│   │   └── root.go            # v3 root command
│   │
│   └── v4/                    # v4 implementation
│       ├── auth/              # Auth commands (SDK v4)
│       ├── transfer/          # Transfer commands (SDK v4)
│       ├── group/             # Group commands (SDK v4)
│       ├── timer/             # Timer commands (SDK v4)
│       ├── search/            # Search commands (SDK v4)
│       ├── flows/             # Flows commands (SDK v4)
│       ├── compute/           # Compute commands (SDK v4)
│       └── root.go            # v4 root command
│
├── pkg/
│   ├── shared/                # Shared code (both versions)
│   │   ├── config/            # Configuration management
│   │   ├── output/            # Output formatters (text/JSON/CSV)
│   │   ├── auth/              # Common auth patterns
│   │   ├── common/            # Shared utilities
│   │   └── version/           # Version detection/routing
│   │
│   ├── v3/                    # v3-specific code
│   │   ├── client/            # v3 client creation helpers
│   │   └── errors/            # v3 error handling
│   │
│   └── v4/                    # v4-specific code
│       ├── client/            # v4 client creation helpers
│       └── errors/            # v4 error handling
│
├── main.go                    # Entry point (version dispatcher)
├── go.mod                     # Both SDK v3 and v4
├── go.work                    # Workspace (if needed)
│
├── docs/
│   ├── v3/                    # v3 documentation
│   ├── v4/                    # v4 documentation
│   └── migration.md           # v3 → v4 migration guide
│
└── CLI_DUAL_VERSION_STRATEGY.md  # This file
```

---

## Benefits of Dual-Version Approach

### 1. Risk Mitigation
- **Low Risk Migration:** Commands can be migrated one at a time
- **Easy Rollback:** If v4 has issues, fall back to v3
- **Gradual Testing:** Real-world v4 testing without breaking v3
- **Production Safety:** v3 stays stable during v4 development

### 2. User Experience
- **No Forced Migration:** Users choose when to migrate
- **Easy Testing:** `--sdk-version=v4` to test without commitment
- **Backward Compatible:** Existing scripts keep working
- **Clear Migration Path:** Documentation shows both versions

### 3. Development Benefits
- **Parallel Development:** Can work on v4 without blocking v3 fixes
- **Code Comparison:** Easy to compare v3 vs v4 implementations
- **Learning Curve:** Developers can learn v4 patterns gradually
- **Best Practices:** Extract common patterns to pkg/shared

### 4. Alignment with SDK
- **Consistent Approach:** CLI mirrors SDK's dual-version strategy
- **Matching Versions:** CLI v3 uses SDK v3, CLI v4 uses SDK v4
- **Synchronized Releases:** Can track both SDK versions
- **Clear Dependencies:** go.mod shows both SDK dependencies

---

## Comparison: Single vs Dual Version

| Aspect | Single Version (Current) | Dual Version (Proposed) |
|--------|-------------------------|-------------------------|
| **Migration Risk** | High (all-at-once) | Low (gradual) |
| **User Choice** | Forced migration | Choose when to migrate |
| **Development** | Must finish all commands | Incremental releases |
| **Testing** | Big-bang testing | Continuous validation |
| **Rollback** | Difficult | Easy (use v3) |
| **Code Size** | Smaller | Larger (temporarily) |
| **Maintenance** | Single codebase | Two codebases (temporary) |
| **Alignment with SDK** | Poor | Excellent |

---

## Cost-Benefit Analysis

### Costs
- **Development Time:** 2-3 weeks restructuring + ongoing dual maintenance
- **Binary Size:** ~20-30% larger (includes both v3 and v4 code)
- **Test Coverage:** Need tests for both versions
- **Documentation:** Maintain docs for both versions
- **Complexity:** More complex main.go and build process

### Benefits
- **Risk Reduction:** Worth 2-3 weeks to avoid catastrophic migration failure
- **User Satisfaction:** Gradual migration = happier users
- **Production Safety:** v3 keeps working while v4 matures
- **Future-Proof:** When upstream v4 is official, we're ready
- **SDK Alignment:** Consistent with SDK's proven approach

**ROI:** Costs are upfront and temporary. Benefits are long-term and substantial.

---

## Decision Criteria

### Proceed with Dual-Version if:
✅ Upstream Globus announces v4 timeline
✅ Users start requesting v4 features
✅ SDK v4 reaches production maturity
✅ CLI v3 is stable and well-tested
✅ Team has bandwidth for restructuring

### Stay Single-Version if:
❌ Upstream v4 is years away
❌ No user demand for v4
❌ SDK v4 is still experimental
❌ Team is resource-constrained

---

## Timeline Summary

| Phase | Duration | Effort | Milestone |
|-------|----------|--------|-----------|
| **Phase 1: Restructure** | 2-3 weeks | High | Dual-version structure |
| **Phase 2a: Core Commands** | 8-12 weeks | Medium | Basic v4 functionality |
| **Phase 2b: Advanced Commands** | 12-16 weeks | Medium | Full v4 parity |
| **Phase 2c: Polish** | 4 weeks | Low | v4 production-ready |
| **Phase 3: Transition** | 6-12 months | Low | v4 becomes default |
| **Phase 4: v3 Removal** | 1 week | Low | v4 only |
| **Total** | ~12-18 months | Variable | Complete migration |

---

## Success Metrics

### Phase 1 Success
- ✅ All v3 commands work unchanged
- ✅ CI/CD passes for both v3 and v4 paths
- ✅ Binary size increase < 30%
- ✅ No performance degradation

### Phase 2 Success
- ✅ Each v4 command has feature parity with v3
- ✅ v4 commands have comprehensive tests
- ✅ Users report successful v4 testing
- ✅ No critical v4 bugs

### Phase 3 Success
- ✅ 50%+ users migrated to v4
- ✅ v4 performance equal or better than v3
- ✅ No v4 blocker issues
- ✅ Positive user feedback

### Phase 4 Success
- ✅ v3 code removed
- ✅ Binary size reduced
- ✅ All users on v4
- ✅ No migration issues

---

## Risks and Mitigations

### Risk 1: Binary Size Growth
**Impact:** Medium
**Mitigation:** Build tags or separate binaries if size becomes issue

### Risk 2: Maintenance Burden
**Impact:** High
**Mitigation:** Time-box v3 support (12 months), focus on v4

### Risk 3: User Confusion
**Impact:** Medium
**Mitigation:** Clear documentation, version detection, helpful errors

### Risk 4: Divergent Features
**Impact:** Low
**Mitigation:** pkg/shared keeps common code DRY

### Risk 5: Delayed Migration
**Impact:** Medium
**Mitigation:** Set hard deadlines for v4 default and v3 removal

---

## Alternatives Considered

### Alternative 1: Separate Repositories
Create `globus-go-cli-v4` as separate repo.

**Pros:** Clean separation
**Cons:** Duplicated infrastructure, harder to share code, confusing for users

### Alternative 2: Feature Flags
Use feature flags to enable/disable v4 code paths.

**Pros:** Single codebase
**Cons:** Complex, error-prone, hard to test both paths

### Alternative 3: Big Bang Migration
Migrate everything at once in a major version.

**Pros:** Simple, clean
**Cons:** High risk, forced user migration, no fallback

**Conclusion:** Dual-version structure (like SDK) is optimal balance.

---

## Recommendation

✅ **Adopt dual-version structure when upstream Globus v4 timeline is announced**

**Why:**
1. Proven approach (SDK uses it successfully)
2. Low-risk gradual migration
3. Users not forced to migrate immediately
4. Easy to validate v4 in production
5. Clear alignment with SDK architecture

**When to start:**
- When upstream announces v4 beta/RC timeline
- When SDK v4 reaches production maturity (more features beyond Close())
- When users start requesting v4 features
- Estimated: Q2-Q3 2026 (6-9 months from now)

**Next steps when ready:**
1. Review and approve this strategy document
2. Create GitHub project for tracking Phase 1 restructuring
3. Create detailed work breakdown (2-3 weeks estimate)
4. Begin Phase 1: Repository restructuring
5. Announce dual-version approach to users

---

## References

- [SDK Dual-Version Structure](https://github.com/scttfrdmn/globus-go-sdk) - Proven approach
- [SDK_VERSION_DECISION.md](SDK_VERSION_DECISION.md) - Current v3-only decision
- [DEVELOPMENT_ROADMAP.md](../globus-go-sdk/DEVELOPMENT_ROADMAP.md) - SDK's phased approach
- [Python SDK v4 Tracking](../globus-go-sdk/PYTHON_SDK_V4.2.0_TRACKING.md) - v4 features

---

**Status:** This is a planning document for future consideration. Current strategy remains v3-only until upstream Globus v4 becomes official or user demand warrants earlier adoption.

**Review Schedule:** Quarterly (check if conditions warrant starting dual-version work)
