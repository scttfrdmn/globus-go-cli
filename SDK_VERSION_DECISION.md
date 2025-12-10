<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# SDK Version Decision: v3 vs v4 for Feature Parity

**Decision Date:** 2025-10-25
**Status:** ✅ DECIDED - Proceed with SDK v3.65.0-1
**Decision:** Option 2 - Implement services on v3, migrate to v4 later

**Latest Update (2025-12-09):**
- SDK v3.65.0-1 ✅ In use (production-ready)
- SDK v4.2.0-1 📋 Available but not adopting (minimal CLI benefit)
- Decision to stay on v3 remains valid - v4's benefits (context-first, Close() methods) provide minimal value for CLI use case vs migration cost

---

## Executive Summary

We must decide whether to implement the 5 remaining services (Groups, Timers, Search, Flows, Compute) using:
- **SDK v3.65.0-1** (current, stable, no breaking changes), OR
- **SDK v4.1.0-2** (latest, modern architecture, requires migration)

**Recommendation:** **Migrate to SDK v4.1.0-2 first, then implement new services**

This is the optimal time to migrate because:
1. We're undertaking major development work anyway (~5.5 months)
2. V4 provides 100% Python SDK parity needed for full CLI parity
3. Doing it now avoids migrating both existing + new code later
4. Modern architecture (context-first, explicit scopes) aligns with Go best practices

---

## Current Codebase Statistics

| Metric | Count | Notes |
|--------|-------|-------|
| Implementation files (non-test) | 14 | Auth + Transfer |
| SDK v3 imports | 24 | Across all cmd files |
| Total non-test lines | ~3,229 | Needs context.Context additions |
| Services implemented | 2/7 | Auth, Transfer |
| Services remaining | 5/7 | Groups, Timers, Search, Flows, Compute |

---

## Option 1: Continue with SDK v3.65.0-1

### Approach
Continue using v3.x SDK, implement all 5 services, then migrate to v4 later.

### Pros ✅
- **No breaking changes** - Current code continues working
- **Lower immediate risk** - Proven, stable SDK
- **Faster to start** - Begin implementing services immediately
- **Incremental approach** - Can migrate to v4 later when ready
- **Less refactoring upfront** - ~3,229 lines unchanged initially

### Cons ❌
- **Double work** - Will need to migrate 7 services to v4 eventually (not just 2)
- **Technical debt** - Building on older architecture
- **Not future-proof** - V3 is maintenance mode, v4 is the future
- **Missed improvements** - No context support, older error handling
- **Larger migration later** - ~15,000+ lines to migrate instead of ~3,229
- **Feature limitations** - May lack some v4-only SDK features

### Effort Estimate
- **Immediate:** 0 weeks (no migration)
- **Services:** 22 weeks (5 services implementation)
- **Future v4 migration:** 8-12 weeks (all 7 services)
- **Total:** ~30-34 weeks

---

## Option 2: Migrate to SDK v4.1.0-2 First ⭐ RECOMMENDED

### Approach
Migrate existing Auth/Transfer to v4, then implement remaining 5 services with v4 from the start.

### Pros ✅
- **Modern architecture** - Context-first design (Go best practice)
- **100% Python SDK parity** - Required for full CLI feature parity
- **Better error handling** - Structured errors with request IDs
- **Explicit scopes** - Better security posture
- **Single migration** - Migrate 2 services now, not 7 later
- **Future-proof** - Build on the current SDK architecture
- **Clean slate** - New services built right from day one
- **Smaller initial migration** - Only ~3,229 lines vs ~15,000+ later

### Cons ❌
- **Breaking changes** - Requires refactoring existing code
- **Higher upfront effort** - 2-3 weeks migration before new features
- **Testing overhead** - All existing tests need updates
- **Risk during migration** - Potential for regressions

### Effort Estimate
- **v4 Migration:** 2-3 weeks (migrate Auth/Transfer)
- **Services:** 22 weeks (5 services implementation on v4)
- **Future v4 migration:** 0 weeks (already done)
- **Total:** ~24-25 weeks

---

## Migration Scope: v3 → v4

### Required Changes for Existing Code

#### 1. Import Path Updates (Automated)
```go
// Before (v3)
import "github.com/scttfrdmn/globus-go-sdk/v3/auth"

// After (v4)
import "github.com/scttfrdmn/globus-go-sdk/v4/auth"
```

#### 2. Context.Context Parameters (Manual)
```go
// Before (v3)
func GetUserInfo() (*UserInfo, error) {
    client := auth.NewClient(config)
    return client.GetUserInfo()
}

// After (v4)
func GetUserInfo(ctx context.Context) (*UserInfo, error) {
    client, err := auth.NewClient(ctx, config)
    if err != nil {
        return nil, err
    }
    return client.GetUserInfo(ctx)
}
```

#### 3. Config Struct (Manual)
```go
// Before (v3)
client := auth.NewClient(
    auth.WithClientID(clientID),
    auth.WithClientSecret(clientSecret),
)

// After (v4)
config := core.Config{
    ClientID:     clientID,
    ClientSecret: clientSecret,
    Scopes:       []string{"openid", "profile"},
}
client, err := auth.NewClient(ctx, config)
```

#### 4. Error Handling (Manual)
```go
// Before (v3)
if err != nil {
    return fmt.Errorf("API error: %w", err)
}

// After (v4)
if err != nil {
    if apiErr, ok := err.(*core.APIError); ok {
        return fmt.Errorf("API error [%s] (status %d): %w",
            apiErr.RequestID, apiErr.StatusCode, err)
    }
    return fmt.Errorf("error: %w", err)
}
```

### Files Requiring Changes

| File Category | Count | Change Type |
|--------------|-------|-------------|
| Import statements | 24 | Automated find/replace |
| Client initialization | ~14 | Manual - config struct |
| API method calls | ~50-100 | Manual - add context.Context |
| Error handling | ~50-100 | Manual - use structured errors |
| Test files | ~20 | Manual - update mocks & fixtures |

### Migration Resources Available
- ✅ **V4 Migration Guide** in SDK repository
- ✅ **V4 Quick Start** with examples
- ✅ **Side-by-side comparisons** for common patterns
- ✅ **V3 and v4 can coexist** during gradual migration

---

## Comparison Matrix

| Factor | v3 Approach | v4 Approach |
|--------|-------------|-------------|
| **Immediate effort** | 0 weeks | 2-3 weeks |
| **Total effort** | 30-34 weeks | 24-25 weeks |
| **Lines to migrate eventually** | ~15,000+ | ~3,229 |
| **Architecture** | Older | Modern |
| **Feature completeness** | Good | Excellent (100% parity) |
| **Future maintenance** | Higher (tech debt) | Lower (current arch) |
| **Risk of rework** | High | Low |
| **Go best practices** | Partial | Full (context-first) |

---

## Detailed Analysis

### Why v4 Migration Now Makes Sense

#### 1. **Economies of Scale**
- Migrating 2 services now < Migrating 7 services later
- Current: ~3,229 lines to update
- Later: ~15,000+ lines to update (5x more work)

#### 2. **Natural Breakpoint**
- Just completed version alignment (v3.39.0-1)
- Starting major new development (5 services)
- Team knowledge is fresh from recent SDK update

#### 3. **SDK Maturity**
- V4.1.0-2 includes all services (Transfer, Search, Flows, Timers, Compute)
- Released October 25, 2025 (same day as v3.65.0-1)
- Both are "latest" - v4 is not experimental

#### 4. **Architectural Benefits**
```go
// v3: No cancellation support
client.ListEndpoints()

// v4: Proper Go patterns with context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
client.ListEndpoints(ctx)
```

This enables:
- Request cancellation
- Timeout handling
- Deadline propagation
- Tracing integration

#### 5. **Error Handling Quality**
```go
// v4 structured errors
if apiErr, ok := err.(*core.APIError); ok {
    log.Printf("Request ID: %s", apiErr.RequestID)  // For support tickets
    log.Printf("Status: %d", apiErr.StatusCode)     // For debugging
    log.Printf("Code: %s", apiErr.ErrorCode)        // For handling
}
```

### Why NOT to Stay on v3

#### 1. **Inevitable Migration**
- V3 is maintenance mode
- V4 is the current SDK architecture
- Every month we delay = more code to migrate

#### 2. **Compound Interest Problem**
```
Month 0 (now):     2 services × 1,614 lines = 3,229 lines to migrate
Month 6 (later):   7 services × ~2,200 lines = 15,400 lines to migrate
```
Migration cost grows ~4.7x if we wait.

#### 3. **Mixed Architecture**
If we implement new services on v3, we'll have:
- Auth/Transfer on v3 (migrated later)
- Groups/Timers/Search/Flows/Compute on v3 (also migrated later)

OR:
- Auth/Transfer on v4 (migrated now)
- Groups/Timers/Search/Flows/Compute on v4 (clean from start)

Second option is cleaner.

---

## Migration Plan (If v4 Chosen)

### Phase 1: Preparation (Week 1)
- [ ] Review v4 migration guide thoroughly
- [ ] Set up v4 SDK in go.mod (can coexist with v3)
- [ ] Create migration branch: `feature/sdk-v4-migration`
- [ ] Set up comprehensive test suite for regression detection

### Phase 2: Auth Service Migration (Week 1-2)
- [ ] Update auth package imports to /v4
- [ ] Refactor client initialization with Config struct
- [ ] Add context.Context to all method signatures
- [ ] Update error handling to use APIError
- [ ] Add explicit scopes to configurations
- [ ] Update all unit tests
- [ ] Update integration tests
- [ ] Verify all auth commands work

### Phase 3: Transfer Service Migration (Week 2-3)
- [ ] Update transfer package imports to /v4
- [ ] Refactor client initialization with Config struct
- [ ] Add context.Context to all method signatures
- [ ] Update error handling to use APIError
- [ ] Add explicit scopes to configurations
- [ ] Update all unit tests
- [ ] Update integration tests
- [ ] Verify all transfer commands work

### Phase 4: Infrastructure Updates (Week 3)
- [ ] Update pkg/config for v4 Config patterns
- [ ] Update pkg/output for v4 error types
- [ ] Update main.go and cmd/root.go
- [ ] Remove v3 dependency from go.mod
- [ ] Update documentation
- [ ] Full regression test suite

### Phase 5: Release v4.0.0-1 (Week 3)
- [ ] Create release notes for v4 migration
- [ ] Update CHANGELOG.md
- [ ] Tag release v4.0.0-1
- [ ] Update README with v4 SDK reference

### Phase 6: Implement New Services on v4 (Week 4-25)
- [ ] Follow FEATURE_PARITY_ROADMAP.md
- [ ] All new code uses v4 patterns from day one
- [ ] No migration needed for these services

---

## Risk Assessment

### Risks of v4 Migration Now
| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Regression bugs | Medium | High | Comprehensive test suite |
| Timeline delay | Low | Medium | Well-documented migration guide |
| API compatibility | Low | High | SDK v4 is stable release |
| Team learning curve | Medium | Low | Good documentation + examples |

### Risks of Staying on v3
| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Larger future migration | **High** | **High** | ❌ None effective |
| Missing v4-only features | Medium | Medium | Limited workarounds |
| Technical debt accumulation | **High** | Medium | ❌ Grows over time |
| Mixed SDK versions | High | High | Complex dependency management |

---

## Recommendation Summary

### ⭐ **Migrate to SDK v4.1.0-2 First**

**Rationale:**
1. **Better total economics:** 24-25 weeks vs 30-34 weeks
2. **Smaller migration scope:** ~3,229 lines now vs ~15,000+ later
3. **Modern architecture:** Context-first, explicit scopes, better errors
4. **100% SDK parity:** Required for full CLI feature parity
5. **No future rework:** New services built right from the start
6. **Optimal timing:** Natural breakpoint in development cycle

**Timeline:**
- Weeks 1-3: SDK v4 migration (Auth + Transfer)
- Weeks 4-25: Implement 5 new services on v4
- Total: 25 weeks to full parity

**Outcome:**
- Modern, maintainable codebase
- Full feature parity with upstream CLI
- No future breaking changes needed
- Follows Go best practices

---

## Decision Criteria

Vote **SDK v4** if you value:
- Long-term maintainability
- Modern Go best practices
- Smaller total effort
- Future-proofing

Vote **SDK v3** if you value:
- Absolute minimum immediate change
- Lower perceived risk
- Familiar patterns (temporarily)

---

## Next Steps (Pending Decision)

### If SDK v4 Chosen:
1. Create feature branch `feature/sdk-v4-migration`
2. Begin Phase 1: Preparation (Week 1)
3. Follow migration plan above
4. Release v4.0.0-1 after migration
5. Implement new services on v4

### If SDK v3 Chosen:
1. Begin implementing Groups service on v3
2. Follow FEATURE_PARITY_ROADMAP.md
3. Plan v4 migration for 6+ months out
4. Expect 8-12 week migration effort later

---

## References

- [SDK v4.1.0-2 Release Notes](https://github.com/scttfrdmn/globus-go-sdk/releases/tag/v4.1.0-2)
- [SDK v4.1.0-1 Breaking Changes](https://github.com/scttfrdmn/globus-go-sdk/releases/tag/v4.1.0-1)
- [SDK v3.65.0-1 Release Notes](https://github.com/scttfrdmn/globus-go-sdk/releases/tag/v3.65.0-1)
- [Feature Parity Roadmap](FEATURE_PARITY_ROADMAP.md)

---

**Awaiting Decision:** Please review and decide whether to proceed with v3 or v4 approach.

**Recommended:** SDK v4.1.0-2 migration followed by new service implementation
