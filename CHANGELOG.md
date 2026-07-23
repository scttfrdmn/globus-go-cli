# Changelog

All notable changes to the Globus Go CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [4.8.1-6] - 2026-07-23

### Fixed
- **`project list`/`show` now decode projects with an expanded `admins` object
  (SDK #61).** The SDK typed `admins.identities`/`admins.groups` as `[]string`,
  but the Globus Auth Projects API returns them as arrays of objects, so any
  project carrying `admins` failed to decode right after auth — blocking the
  whole `project` command tree (including `project client create`). Fixed by
  pinning SDK **v4.8.1-7** (adds `ProjectAdminIdentity`/`ProjectAdminGroup`) and
  reading admin identity/group IDs from the object fields in `project admin`.

## [4.8.1-5] - 2026-07-23

Makes the `project`/`collection`/`session` consent flows work out of the box.
Builds on the v4 SDK (v4.8.1-6). Version tracks upstream Python globus-sdk
4.8.1; `-5` is this project's patch release.

### Fixed
- **Auth consent now works with the shipped default client (#30, #32).** The
  login flow now uses PKCE (via SDK v4.8.1-6), and the default client is
  globus-go-cli's own registered native/public client
  (`ccc07ea1-…`, registered for the `.../v2/web/auth-code` redirect). The
  previous default could not complete the `manage_projects` consent, so
  `project`/`collection`/`session` commands failed at the browser step out of
  the box; they now succeed without a `GLOBUS_CLIENT_ID` override.
- Clearer error when a consent-escalation login fails (points at the client /
  `GLOBUS_CLIENT_ID` fix).

### Added
- `PRIVACY.md` — privacy policy (the CLI runs locally, talks only to Globus
  services, stores tokens under `~/.globus-cli/`, sends no telemetry).

### CI
- `workflow_dispatch` added to the build/tests/security workflows so PR checks
  can be re-run manually when GitHub Actions scheduling lags.

## [4.8.1-4] - 2026-07-23

Adds Globus Connect Personal (GCP) support, completing coverage of the Python
Globus CLI command surface. Builds on the v4 SDK (v4.8.1-5). Version tracks
upstream Python globus-sdk 4.8.1; `-4` is this project's patch release.

### Added — Globus Connect Personal (`gcp`) + `endpoint local-id`
Matching the Python CLI, `gcp` commands manage GCP endpoints/collections through
the Globus service API — they do not install, start, or stop a local GCP agent.
- `gcp create mapped DISPLAY_NAME` — registers a GCP endpoint (via
  `transfer.CreateEndpoint`) and prints the `globus_connect_setup_key` used to
  configure an installed agent.
- `gcp create guest DISPLAY_NAME HOST_ENDPOINT_ID:PATH` — creates a guest
  collection on a GCP host.
- `gcp set-subscription-id ENDPOINT_ID SUBSCRIPTION_ID`.
- Both `create` subcommands carry the shared endpoint/collection metadata flags
  (description, organization, contact-*, keywords, verify, force-encryption,
  user-message, default-directory; `mapped` adds public/private/subscription-id).
- `endpoint local-id` — prints the local Globus Connect Personal endpoint ID by
  reading `~/.globusonline/lta/client-id.txt` (no network).

### Notes
- Requires v4 SDK v4.8.1-5 (`transfer.CreateEndpoint`).

## [4.8.1-3] - 2026-07-23

Flag parity with the Python Globus CLI across the command tree. Builds on the v4
SDK (v4.8.1-4). Version tracks upstream Python globus-sdk 4.8.1; `-3` is this
project's patch release.

### Added — per-command flags matching the Python `globus` CLI
Audited every command's flags against the reference Python CLI and added the
missing ones wherever the v4 SDK backs the underlying request field. Highlights:
- **transfer file ops**: `ls` `--filter`/`--orderby`/`--local-user`; `transfer`
  `--encrypt-data`, `--notify`, `--skip-source-errors`, `--fail-on-quota-errors`,
  `--delete-destination-extra`, `--external-checksum`, `--checksum-algorithm`,
  `--source-local-user`/`--dest-local-user`; `rm`/`delete` `--label`,
  `--deadline`, `--notify`, `--ignore-missing`, `--enable-globs`, `--local-user`;
  `mkdir`/`rename`/`stat` `--local-user`; `task list` `--filter-status`/
  `--orderby`; `task wait` `--polling-interval`/`-H,--heartbeat`.
- **endpoint / collection / gcs**: endpoint search filters; the full
  `endpoint update` field set; role/permission `--identity`/`--group`/
  `--all-authenticated`/`--anonymous`/notify/expiration; the full
  `CollectionDocument` field set on `collection create`/`update`.
- **groups/search/flows/timer/auth**: group `--parent-id`/`--request`/policy
  fields; `search query --query-document`; flows administrators/starters/viewers/
  run-managers/monitors + subtitle/subscription/auth-policy/owner; timer
  `--stop-after-runs`/`--label`; `get-identities --provision`;
  `whoami --linked-identities`; `session update --scope`.

### Notes
- Flags the v4 SDK cannot express are intentionally **not** added as no-ops; the
  SDK-blocked flags are listed in `docs/PYTHON_CLI_PARITY.md`.

## [4.8.1-2] - 2026-07-22

Adds Globus Auth project/console management and the first GCS data-plane
command. Builds on the v4 SDK (v4.8.1-4). Version tracks upstream Python
globus-sdk 4.8.1; `-2` is this project's patch release.

### Added — project / client / credential management (Go-only extension)
Ports the standalone `globus-project-manager` tool into the CLI. The Python
`globus-cli` has no project-management commands (it is web-console only).
- `project list/show/create/update/delete` and `project admin list/add/remove`
  (admin add resolves a username to an identity).
- `project client list/show/create/update/delete` plus `update-redirect-uris`
  and `update-metadata` — manage a project's registered clients (service
  accounts / app registrations).
- `project credential list/create/delete` plus rotation: `rotate`, `list-age`,
  `process-deletions`. Rotation creates a replacement credential with a
  transition window; state (linked credentials + scheduled-deletion dates) is
  tracked in a per-profile local file `~/.globus-cli/credential-state-<profile>.json`.
- These require the `manage_projects` scope (not in the default login scope
  set); the CLI escalates a one-time consent stored under a dedicated token
  namespace so it never collides with the login token.

### Added — GCS data-plane file read
- `collection cat ENDPOINT_ID COLLECTION_ID PATH` reads a file over an
  HTTPS-enabled collection's data plane, using the collection's `https` scope (a
  separate per-collection data-access consent, escalated on first use).

## [4.8.1-1] - 2026-07-21

Drop-in replacement release: the CLI now matches the Python Globus CLI's command
surface and behavior, built entirely on the v4 Go SDK (v4.8.1-4). Version tracks
the upstream Python globus-sdk the CLI builds against (4.8.1); `-1` is this
project's patch release.

### Changed — flat command structure (breaking)
- Auth and transfer commands are now **top-level**, matching the Python CLI:
  `globus login/logout/whoami/get-identities`, `globus ls/mkdir/rm/rename/stat/
  delete/transfer`, `globus task ...`, `globus endpoint ...`. Nested invocations
  (`globus auth login`, `globus transfer ls`) no longer exist.
- `-F/--format unix|json|text`, `--jmespath/--jq`, `--map-http-status`,
  `--quiet`, and `GLOBUS_PROFILE` now match the Python CLI's global flags.
- `-F json` on list commands emits the enveloped `{"DATA_TYPE":...,"DATA":[...]}`
  document, matching the Python CLI.

### Added — v4 SDK migration + GlobusApp auth
- All services build on the v4 SDK with per-resource-server GlobusApp tokens.
  `login` uses the OAuth2 authorization-code flow; the legacy single-token
  bridge was removed. `device` runs a real device-code flow.
- **auth**: `session show/update/consent`, real `get-identities`.
- **transfer**: `rename`, `stat`, `delete`, `task event-list/pause-info/update`;
  `endpoint update/delete/role/permission`, `set-subscription-id`,
  `my-shared-endpoint-list`; `bookmark` group; `endpoint-manager` admin group;
  real `tunnel`/`stream-access-point` (Streams) commands.
- **groups**: real `join/leave/invite`, `member accept/decline/approve/reject`,
  `policies show/set`.
- **GCS**: `collection list/show/create/update/delete` and `gcs info` +
  `storage-gateway`/`role` subcommands, with dynamic per-endpoint consent
  escalation.
- **api**: raw passthrough (`globus api <service> METHOD PATH`).
- **meta**: `list-commands`, `version`.
- **search**: `index role list/create/delete`, `task list`. **flows**: `validate`,
  `run delete`, `run resume` (previously placeholders).

### Notes
- Out of scope: GCP (Globus Connect Personal) — a local agent with no SDK API.
- `docs/PYTHON_CLI_PARITY.md` tracks command coverage vs Python globus-cli 3.42.0.

## [4.5.0-1] - 2026-04-03

### Added
- **Transfer: Globus Streams Tunnels** (`globus transfer tunnel`) — new subcommand group tracking Python SDK v4.3.0/v4.4.0:
  - `tunnel create` — create a new streaming tunnel (`--name`, `--source-endpoint`, `--source-path`, `--expires-in`)
  - `tunnel list` — list tunnels owned by the current user
  - `tunnel show TUNNEL_ID` — show tunnel details
  - `tunnel update TUNNEL_ID` — update tunnel display name
  - `tunnel delete TUNNEL_ID` — permanently delete a tunnel
  - `tunnel events TUNNEL_ID` — list events for a tunnel (Python SDK v4.4.0)
- **Transfer: Globus Streams Access Points** (`globus transfer stream-access-point`) — new subcommand group:
  - `stream-access-point show ACCESS_POINT_ID` — show stream access point details and URL
- **Search: Index Reopen** (`globus search index reopen INDEX_ID`) — reopen a previously deleted index (Python SDK v4.0.0)
- **Flows: Authentication Policy** — `globus flows create` and `globus flows update` now accept:
  - `--high-assurance` — require high-assurance authentication for flow runs
  - `--required-mfa` — require multi-factor authentication for flow runs
  - `--session-policies` — specify named session policies (Python SDK v4.1.0)

### Changed
- CLI version bumped to `4.5.0-1`, aligned with Go SDK tracking of Python SDK v4.5.0
- Copyright year updated to 2025-2026 across all source files
- `go.work` updated to reference local SDK for workspace-mode development

### Technical
- SDK dependency: `github.com/scttfrdmn/globus-go-sdk/v3` — local workspace at Python SDK v4.5.0 parity
- New files: `cmd/transfer/tunnel.go`, `cmd/transfer/stream_access_point.go`, `cmd/search/index_reopen.go`

## [3.39.0-1] - 2025-10-25

### Added - Major Feature Release
- **Complete Groups Service** (12 commands, 80% coverage)
  - Group management: create, list, show, update, delete
  - Membership management: add/remove members, list members
  - Policy management: view and update group policies
  - Role management commands as placeholders (pending SDK support)
- **Complete Timers Service** (8 commands, 100% coverage)
  - Timer job management: create, list, show, update, delete
  - Support for one-time, recurring, and cron-based schedules
  - Pause and resume timer functionality
- **Complete Search Service** (18 commands, 100% coverage)
  - Index management: create, list, show, update, delete indices
  - Document operations: query and ingest documents
  - Subject management: show and delete subjects
  - Task monitoring: show task status
  - Index role management placeholders (pending SDK support)
- **Complete Flows Service** (15 commands, 100% coverage)
  - Flow management: create, list, show, update, delete flows
  - Flow execution: start flows with input and monitoring
  - Run management: list, show, cancel, update runs
  - Run log viewing and flow definition inspection
  - Validation and resume commands as placeholders (pending SDK support)
- **Complete Compute Service** (14 commands, 100% coverage) **NEW - Not in Python CLI!**
  - Endpoint management: list and view compute endpoints
  - Function management: register, list, show, update, delete functions
  - Task execution: run functions, monitor status, cancel tasks
  - List task history
- Enhanced shell completion with all new services
- Comprehensive help documentation for all commands

### Changed
- Updated service coverage from 29% (2/7) to 100% (7/7)
- Total commands increased from ~30 to ~79 commands
- **Now exceeds Python Globus CLI feature parity with exclusive Compute support**

### Technical Improvements
- All services follow consistent command structure and patterns
- Unified error handling and authentication flows
- Support for text, JSON, and CSV output formats across all services
- Comprehensive SDK integration with Globus Go SDK v3.65.0-1

## [3.39.0-1] - 2025-10-25

### Changed
- Updated version alignment to match upstream Globus CLI v3.39.0
- Updated to Globus Go SDK v3.65.0-1 with latest improvements
- Maintained backward compatibility with existing functionality

### Added (SDK-level)
- FlowTimer helper methods for simplified timer creation
  - `CreateFlowTimerOnce()` for one-time flow executions
  - `CreateFlowTimerRecurring()` for ISO 8601 interval-based scheduling
  - `CreateFlowTimerCron()` for cron-based flow scheduling
- Groups status filtering support in `ListGroups()` method
- Enhanced SDK capabilities for future Groups and Timers command implementation

### Technical Improvements
- All unit tests continue to pass with new SDK version
- Integration tests compile successfully with SDK v3.65.0-1
- Enhanced compatibility with upstream project versioning
- SDK update enables future implementation of Groups and Timers commands

### Notes
- This release tracks upstream CLI v3.39.0 for versioning alignment
- Auth and Transfer commands remain fully implemented and tested
- Groups and Timers commands remain in planned stage for future releases
- SDK v3.65.0-1 provides all necessary capabilities for future service implementations

## [3.37.0-1] - 2025-09-18

### Changed
- Updated version alignment to match upstream Globus CLI v3.37.0
- Updated to Globus Go SDK v3.63.0-1 with latest improvements
- Maintained backward compatibility with existing functionality

### Technical Improvements
- All unit tests continue to pass with new SDK version
- Integration tests compile successfully with SDK v3.63.0-1
- Enhanced compatibility with upstream project versioning

## [3.36.0-1] - 2025-09-01

### Changed
- Updated version alignment to match upstream Globus CLI v3.36.0
- Changed version to v3.36.0-1 following established versioning pattern
- Updated SDK to v3.62.0-3 with improved integration testing support

### Fixed
- Resolved Integration Test compilation errors for SDK v3.62.0-3
- Fixed auth client initialization using separate WithClientID/WithClientSecret options
- Updated GetClientCredentialsToken to use individual string parameters
- Fixed transfer client initialization using proper authorizer pattern
- Updated ListDirectory and SubmitTransfer method signatures for SDK v3.62.0-3

### Technical Improvements
- Applied comprehensive code formatting with go fmt
- All unit tests pass with no regressions
- All integration tests compile successfully
- Enhanced GitHub Actions CI/CD pipeline compatibility

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