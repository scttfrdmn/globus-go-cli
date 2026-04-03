<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors -->

# Feature Parity Roadmap

This document tracks feature parity with the upstream Globus CLI and Python SDK.

**Current Status:** All 7 services complete; CLI at v4.5.0-1 (Python SDK v4.5.0 parity)
**Last Updated:** 2026-04-03

---

## Overview

The globus-go-cli currently implements:
- ✅ **Auth Service** - Complete (9 commands)
- ✅ **Transfer Service** - Complete (6 commands + Streams/Tunnels: tunnel create/list/show/update/delete/events, stream-access-point show)
- ✅ **Groups Service** - Complete (12 commands)
- ✅ **Timers Service** - Complete (8 commands)
- ✅ **Search Service** - Complete (19 commands, including `index reopen`)
- ✅ **Flows Service** - Complete (15 commands, including authentication policy flags)
- ✅ **Compute Service** - Complete (14 commands) — exclusive to Go CLI, not in Python CLI

---

## Implementation Priorities

### Priority 1: Core Services (Most Commonly Used)
1. **Groups** - Collaborative group management
2. **Timers** - Scheduled transfers and tasks

### Priority 2: Advanced Services
3. **Search** - Data discovery and indexing
4. **Flows** - Workflow automation

### Priority 3: Specialized Services
5. **Compute** - Function as a Service (FaaS)

---

## Service Implementation Plans

## 1. Groups Service

**SDK Support:** ✅ Available in SDK v3.65.0-1 (groups package)
**Complexity:** Medium
**Estimated Effort:** 2-3 weeks
**Dependencies:** None

### Command Structure

```
globus group
├── create                    # Create a new group
├── delete <GROUP_ID>         # Delete a group
├── join <GROUP_ID>           # Join a group
│   └── --request            # Request to join (requires approval)
├── leave <GROUP_ID>          # Leave a group
├── list                      # List groups you belong to
├── show <GROUP_ID>           # Show group details
├── update <GROUP_ID>         # Update group policies
│   ├── --name               # Update group name
│   ├── --description        # Update description
│   └── --terms-and-conditions  # Set T&C URL
├── member
│   ├── add <GROUP_ID> <IDENTITY_ID>     # Add member
│   │   └── --role [member|manager|admin]  # Specify role
│   ├── invite <GROUP_ID> <EMAIL>        # Invite member
│   │   └── --provision-identity         # Provision new identity
│   ├── list <GROUP_ID>                  # List group members
│   ├── remove <GROUP_ID> <IDENTITY_ID>  # Remove member
│   ├── approve <GROUP_ID> <IDENTITY_ID> # Approve join request
│   └── reject <GROUP_ID> <IDENTITY_ID>  # Reject join request
└── set-subscription-admin-verified <GROUP_ID>  # Set subscription admin (v3.38.0 feature)
    └── --subscription-id <ID>
```

### Implementation Tasks

#### Phase 1: Core Group Operations (Week 1)
- [ ] Create `cmd/group.go` - Root group command
- [ ] Create `cmd/group/create.go` - Group creation
- [ ] Create `cmd/group/delete.go` - Group deletion
- [ ] Create `cmd/group/list.go` - List groups
- [ ] Create `cmd/group/show.go` - Show group details
- [ ] Create `cmd/group/update.go` - Update group settings
- [ ] Add unit tests for core operations

#### Phase 2: Membership Operations (Week 2)
- [ ] Create `cmd/group/join.go` - Join group
- [ ] Create `cmd/group/leave.go` - Leave group
- [ ] Create `cmd/group/member_add.go` - Add member
- [ ] Create `cmd/group/member_invite.go` - Invite member
- [ ] Create `cmd/group/member_list.go` - List members
- [ ] Create `cmd/group/member_remove.go` - Remove member
- [ ] Create `cmd/group/member_approve.go` - Approve join request
- [ ] Create `cmd/group/member_reject.go` - Reject join request
- [ ] Add unit tests for membership operations

#### Phase 3: Advanced Features & Testing (Week 3)
- [ ] Implement `set-subscription-admin-verified` (v3.38.0 feature)
- [ ] Leverage SDK v3.65.0-1 Groups status filtering
- [ ] Add integration tests with real Globus Groups API
- [ ] Add output formatting (text/JSON/CSV) for all commands
- [ ] Documentation and examples
- [ ] End-to-end testing

---

## 2. Timers Service

**SDK Support:** ✅ Available in SDK v3.65.0-1 (timers package with FlowTimer helpers)
**Complexity:** Medium
**Estimated Effort:** 2-3 weeks
**Dependencies:** None (but enhanced by Flows for flow timers)

### Command Structure

```
globus timer
├── create
│   ├── transfer              # Create recurring transfer timer
│   │   ├── --name           # Timer name
│   │   ├── --interval       # ISO 8601 interval
│   │   ├── --start          # Start time
│   │   ├── --stop           # Stop time
│   │   ├── --include        # Include patterns (v3.38.0)
│   │   └── --exclude        # Exclude patterns (v3.38.0)
│   └── flow <FLOW_ID>        # Create recurring flow timer (v3.39.0)
│       ├── --name           # Timer name
│       ├── --interval       # ISO 8601 interval
│       ├── --cron           # Cron expression
│       └── --input          # Flow input parameters
├── list                      # List your timers
├── show <TIMER_ID>           # Display timer details
│   └── (includes Activity status field - v3.39.0)
├── pause <TIMER_ID>          # Pause a timer
├── resume <TIMER_ID>         # Resume a timer
├── delete <TIMER_ID>         # Delete a timer
└── update <TIMER_ID>         # Update timer settings
```

### Implementation Tasks

#### Phase 1: Core Timer Operations (Week 1)
- [ ] Create `cmd/timer.go` - Root timer command
- [ ] Create `cmd/timer/list.go` - List timers
- [ ] Create `cmd/timer/show.go` - Show timer details (with Activity status)
- [ ] Create `cmd/timer/pause.go` - Pause timer
- [ ] Create `cmd/timer/resume.go` - Resume timer
- [ ] Create `cmd/timer/delete.go` - Delete timer
- [ ] Add unit tests for core operations

#### Phase 2: Timer Creation (Week 2)
- [ ] Create `cmd/timer/create_transfer.go` - Create transfer timer
- [ ] Add `--include` and `--exclude` flags (v3.38.0 feature)
- [ ] Create `cmd/timer/create_flow.go` - Create flow timer (v3.39.0 feature)
- [ ] Leverage SDK v3.65.0-1 FlowTimer helpers:
  - [ ] Use `CreateFlowTimerOnce()` for one-time execution
  - [ ] Use `CreateFlowTimerRecurring()` for ISO 8601 intervals
  - [ ] Use `CreateFlowTimerCron()` for cron-based scheduling
- [ ] Add unit tests for timer creation

#### Phase 3: Advanced Features & Testing (Week 3)
- [ ] Implement timer update functionality
- [ ] Add comprehensive validation for intervals and cron expressions
- [ ] Add integration tests with real Globus Timers API
- [ ] Add output formatting (text/JSON/CSV) for all commands
- [ ] Documentation and examples
- [ ] End-to-end testing

---

## 3. Search Service

**SDK Support:** ✅ Available in SDK v3.65.0-1 (search package)
**Complexity:** Medium-High
**Estimated Effort:** 3-4 weeks
**Dependencies:** None

### Command Structure

```
globus search
├── query <INDEX_ID>          # Query a search index
│   ├── --query              # Query string
│   ├── --filter             # Filter parameters
│   ├── --limit              # Result limit
│   └── --offset             # Result offset
├── ingest <INDEX_ID>         # Ingest data into index
│   ├── --file               # JSON file with documents
│   └── --batch-size         # Batch size for ingestion
├── delete-by-query <INDEX_ID>  # Delete documents by query
│   └── --query              # Query to match documents
├── index
│   ├── create               # Create a new index
│   │   ├── --display-name   # Index display name
│   │   └── --description    # Index description
│   ├── delete <INDEX_ID>    # Delete an index
│   ├── list                 # List accessible indices
│   ├── show <INDEX_ID>      # Show index details
│   ├── update <INDEX_ID>    # Update index settings
│   └── role
│       ├── create <INDEX_ID> <PRINCIPAL> <ROLE>  # Create role
│       ├── delete <INDEX_ID> <ROLE_ID>           # Delete role
│       └── list <INDEX_ID>                       # List index roles
├── subject
│   ├── show <INDEX_ID> <SUBJECT_ID>  # Show subject details
│   └── delete <INDEX_ID> <SUBJECT_ID>  # Delete subject
└── task
    ├── list <INDEX_ID>       # List tasks
    └── show <TASK_ID>        # Show task details
```

### Implementation Tasks

#### Phase 1: Query & Basic Operations (Week 1)
- [ ] Create `cmd/search.go` - Root search command
- [ ] Create `cmd/search/query.go` - Query index
- [ ] Create `cmd/search/ingest.go` - Ingest documents
- [ ] Create `cmd/search/delete_by_query.go` - Delete by query
- [ ] Add unit tests for query operations

#### Phase 2: Index Management (Week 2)
- [ ] Create `cmd/search/index_create.go` - Create index
- [ ] Create `cmd/search/index_delete.go` - Delete index
- [ ] Create `cmd/search/index_list.go` - List indices
- [ ] Create `cmd/search/index_show.go` - Show index details
- [ ] Create `cmd/search/index_update.go` - Update index
- [ ] Add unit tests for index operations

#### Phase 3: Role & Subject Management (Week 3)
- [ ] Create `cmd/search/role_create.go` - Create role
- [ ] Create `cmd/search/role_delete.go` - Delete role
- [ ] Create `cmd/search/role_list.go` - List roles
- [ ] Create `cmd/search/subject_show.go` - Show subject
- [ ] Create `cmd/search/subject_delete.go` - Delete subject
- [ ] Add unit tests for role/subject operations

#### Phase 4: Task Management & Testing (Week 4)
- [ ] Create `cmd/search/task_list.go` - List tasks
- [ ] Create `cmd/search/task_show.go` - Show task
- [ ] Add integration tests with real Globus Search API
- [ ] Add output formatting (text/JSON/CSV) for all commands
- [ ] Documentation and examples
- [ ] End-to-end testing

---

## 4. Flows Service

**SDK Support:** ✅ Available in SDK v3.65.0-1 (flows package)
**Complexity:** High
**Estimated Effort:** 4-5 weeks
**Dependencies:** None (but complements Timers)

### Command Structure

```
globus flows
├── create                    # Create a new flow
│   ├── --title              # Flow title
│   ├── --definition         # Flow definition (JSON)
│   ├── --input-schema       # Input schema (JSON)
│   └── --subtitle           # Flow subtitle
├── update <FLOW_ID>          # Update flow
│   ├── --title              # Update title
│   └── --definition         # Update definition
├── delete <FLOW_ID>          # Delete a flow
├── list                      # List your flows
├── show <FLOW_ID>            # Show flow details
├── validate                  # Validate flow definition
│   └── --definition         # Flow definition to validate
├── start <FLOW_ID>           # Start a flow run
│   ├── --input              # Input parameters (JSON)
│   ├── --label              # Run label
│   ├── --run-managers       # Run managers (principals)
│   ├── --run-monitors       # Run monitors (principals)
│   ├── --tags               # Tags for the run
│   └── --activity-notification-policy  # Notification policy
├── run
│   ├── list                 # List flow runs
│   │   ├── --flow-id        # Filter by flow ID
│   │   ├── --status         # Filter by status
│   │   └── --role           # Filter by role
│   ├── show <RUN_ID>        # Show run details
│   ├── show-definition <RUN_ID>  # Show run definition
│   ├── cancel <RUN_ID>      # Cancel a run
│   ├── resume <RUN_ID>      # Resume a run
│   ├── release <RUN_ID>     # Release a run
│   ├── log <RUN_ID>         # Show run logs
│   └── update <RUN_ID>      # Update run
│       ├── --label          # Update label
│       └── --tags           # Update tags
└── lint <DEFINITION_FILE>    # Lint flow definition
```

### Implementation Tasks

#### Phase 1: Core Flow Operations (Week 1-2)
- [ ] Create `cmd/flows.go` - Root flows command
- [ ] Create `cmd/flows/create.go` - Create flow
- [ ] Create `cmd/flows/update.go` - Update flow
- [ ] Create `cmd/flows/delete.go` - Delete flow
- [ ] Create `cmd/flows/list.go` - List flows
- [ ] Create `cmd/flows/show.go` - Show flow details
- [ ] Create `cmd/flows/validate.go` - Validate flow definition
- [ ] Create `cmd/flows/lint.go` - Lint flow definition
- [ ] Add unit tests for core operations

#### Phase 2: Flow Run Operations (Week 3)
- [ ] Create `cmd/flows/start.go` - Start flow run
- [ ] Create `cmd/flows/run_list.go` - List runs
- [ ] Create `cmd/flows/run_show.go` - Show run details
- [ ] Create `cmd/flows/run_show_definition.go` - Show run definition
- [ ] Create `cmd/flows/run_cancel.go` - Cancel run
- [ ] Create `cmd/flows/run_resume.go` - Resume run
- [ ] Create `cmd/flows/run_release.go` - Release run
- [ ] Add unit tests for run operations

#### Phase 3: Advanced Run Features (Week 4)
- [ ] Create `cmd/flows/run_log.go` - Show run logs
- [ ] Create `cmd/flows/run_update.go` - Update run
- [ ] Implement run managers and monitors support
- [ ] Implement activity notification policy
- [ ] Add comprehensive validation for flow definitions
- [ ] Add unit tests for advanced features

#### Phase 4: Integration & Testing (Week 5)
- [ ] Add integration tests with real Globus Flows API
- [ ] Add output formatting (text/JSON/CSV) for all commands
- [ ] Documentation and examples
- [ ] Flow definition examples and templates
- [ ] End-to-end testing

---

## 5. Compute Service

**SDK Support:** ✅ Available in SDK v3.65.0-1 (compute package)
**Complexity:** Medium-High
**Estimated Effort:** 3-4 weeks
**Dependencies:** Separate globus-compute-endpoint CLI exists

### Command Structure

```
globus compute
├── endpoint
│   ├── list                 # List endpoints
│   ├── show <ENDPOINT_ID>   # Show endpoint details
│   ├── delete <ENDPOINT_ID> # Delete endpoint
│   └── configure            # Configure endpoint with auth policies
├── function
│   ├── register             # Register a function
│   ├── list                 # List registered functions
│   ├── show <FUNCTION_ID>   # Show function details
│   ├── delete <FUNCTION_ID> # Delete function
│   └── run <FUNCTION_ID>    # Run a function
│       ├── --endpoint       # Target endpoint
│       └── --input          # Function input
└── task
    ├── list                 # List tasks
    └── show <TASK_ID>       # Show task details
```

**Note:** The Globus Compute service has a separate `globus-compute-endpoint` CLI for endpoint management (configure, start, stop, etc.). The main CLI integration focuses on function and task management.

### Implementation Tasks

#### Phase 1: Endpoint Operations (Week 1)
- [ ] Create `cmd/compute.go` - Root compute command
- [ ] Create `cmd/compute/endpoint_list.go` - List endpoints
- [ ] Create `cmd/compute/endpoint_show.go` - Show endpoint
- [ ] Create `cmd/compute/endpoint_delete.go` - Delete endpoint
- [ ] Add unit tests for endpoint operations

#### Phase 2: Function Operations (Week 2)
- [ ] Create `cmd/compute/function_register.go` - Register function
- [ ] Create `cmd/compute/function_list.go` - List functions
- [ ] Create `cmd/compute/function_show.go` - Show function
- [ ] Create `cmd/compute/function_delete.go` - Delete function
- [ ] Create `cmd/compute/function_run.go` - Run function
- [ ] Add unit tests for function operations

#### Phase 3: Task Operations & Testing (Week 3-4)
- [ ] Create `cmd/compute/task_list.go` - List tasks
- [ ] Create `cmd/compute/task_show.go` - Show task
- [ ] Add integration tests with real Globus Compute API
- [ ] Add output formatting (text/JSON/CSV) for all commands
- [ ] Documentation and examples
- [ ] Note relationship with globus-compute-endpoint CLI
- [ ] End-to-end testing

---

## Implementation Timeline

### Quarter 1: Core Services (Weeks 1-6)
- **Weeks 1-3:** Groups Service Implementation
- **Weeks 4-6:** Timers Service Implementation

### Quarter 2: Advanced Services (Weeks 7-17)
- **Weeks 7-10:** Search Service Implementation
- **Weeks 11-15:** Flows Service Implementation
- **Weeks 16-17:** Integration and Testing

### Quarter 3: Specialized Services (Weeks 18-21)
- **Weeks 18-21:** Compute Service Implementation
- **Week 22:** Final testing and documentation

**Total Estimated Effort:** ~22 weeks (5.5 months) for full feature parity

---

## Technical Considerations

### Code Structure
Each service should follow the established pattern from Auth and Transfer:
```
cmd/
├── service.go              # Root command (e.g., groups.go, timer.go)
└── service/
    ├── command1.go         # Individual commands
    ├── command2.go
    ├── command1_test.go    # Unit tests
    ├── command2_test.go
    └── service_integration_test.go  # Integration tests
```

### Common Components

All implementations should include:

1. **Client Initialization**
   - Proper SDK client setup with authentication
   - Connection pool configuration
   - Error handling

2. **Output Formatting**
   - Text format (default, human-readable)
   - JSON format (for scripting)
   - CSV format (for data import)

3. **Testing Strategy**
   - Unit tests with mocks (minimum 70% coverage)
   - Integration tests with real API (where feasible)
   - Table-driven tests for comprehensive coverage

4. **Documentation**
   - Command help text
   - Usage examples
   - Update README.md

5. **Error Handling**
   - Proper error propagation with context
   - User-friendly error messages
   - API error translation

### SDK Feature Utilization

Leverage SDK v3.65.0-1 capabilities:
- **Groups:** Status filtering in ListGroups()
- **Timers:** FlowTimer helpers (CreateFlowTimerOnce, CreateFlowTimerRecurring, CreateFlowTimerCron)
- **All Services:** Consistent error handling and type safety

---

## Success Criteria

For each service implementation:

- ✅ All commands from upstream CLI implemented
- ✅ Command-line interface matches upstream patterns
- ✅ All output formats supported (text, JSON, CSV)
- ✅ Unit test coverage ≥ 70%
- ✅ Integration tests pass with real API
- ✅ Documentation complete
- ✅ Code quality checks pass (go vet, go fmt, golangci-lint)
- ✅ Cross-platform compatibility verified

---

## Risks and Mitigation

### Risk 1: SDK API Changes
- **Mitigation:** Monitor SDK releases closely, maintain compatibility tests

### Risk 2: Upstream CLI Changes
- **Mitigation:** Track upstream releases, update roadmap as needed

### Risk 3: Complex Flow Definitions
- **Mitigation:** Start with examples, provide validation tools, comprehensive error messages

### Risk 4: Time Estimation Accuracy
- **Mitigation:** Iterative development, regular progress reviews, adjust timeline as needed

---

## Progress Tracking

Update this section as services are implemented:

| Service | Status | Start Date | Completion Date | Notes |
|---------|--------|------------|-----------------|-------|
| Auth | ✅ Complete | - | 2025-09-18 | 9 commands |
| Transfer | ✅ Complete | - | 2026-04-03 | + Streams/Tunnels (Python SDK v4.3.0/v4.4.0) |
| Groups | ✅ Complete | - | 2025-10-25 | 12 commands |
| Timers | ✅ Complete | - | 2025-10-25 | 8 commands |
| Search | ✅ Complete | - | 2026-04-03 | 19 commands (+index reopen, Python SDK v4.0.0) |
| Flows | ✅ Complete | - | 2026-04-03 | 15 commands (+auth policy, Python SDK v4.1.0) |
| Compute | ✅ Complete | - | 2025-10-25 | 14 commands (Go CLI exclusive) |

---

## Contributing

To contribute to feature parity implementation:

1. Choose a service or command to implement
2. Follow the established patterns from Auth/Transfer
3. Implement with tests (unit + integration)
4. Update this roadmap with progress
5. Submit PR with comprehensive description

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## References

- [Upstream Globus CLI](https://github.com/globus/globus-cli)
- [Globus CLI Documentation](https://docs.globus.org/cli/)
- [Globus Go SDK v3.65.0-1](https://github.com/scttfrdmn/globus-go-sdk)
- [Globus API Documentation](https://docs.globus.org/api/)

---

**Last Updated:** 2025-10-25
**Version:** 1.0.0
**Status:** Planning Phase
