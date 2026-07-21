<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors -->

# Parity with the upstream Python Globus CLI

This tracks the Go CLI's command coverage against the reference
[Globus CLI (Python)](https://github.com/globus/globus-cli). It is a
functional comparison, not a 1:1 command-path map — the two CLIs organize
commands differently (the Go CLI nests task/endpoint commands under each
service; the Python CLI is flatter, e.g. top-level `task`, `bookmark`,
`collection`, `endpoint`).

- **Python globus-cli compared:** 3.42.0 (latest at time of writing), 164 commands.
- **Go CLI:** ~112 command paths across auth, transfer, search, groups, flows,
  compute, timer, config.

Note the version numbers are independent: the Go CLI tracks the Globus **SDK**,
while the Python **CLI** versions separately (3.42.0).

## SDK backing: v4 migration (Phase 2b)

The backable service command groups — **transfer, groups, search, flows,
compute, timer** — now build on the **v4 Go SDK** (`globus-go-sdk/v4`, tracking
Python globus-sdk 4.8.1) with per-resource-server GlobusApp tokens, replacing
the previous v3 SDK + single-combined-token model. Each package has a
`client.go` `getClient(ctx)` helper over `pkg/globusauth`.

The v4 models differ from v3 in places, so a few flags are retained for
command-surface stability but are currently **no-ops** (they do not map to a v4
request field):

- `flows update --public`; `flows` create/update auth-policy flags
  (`--high-assurance`, `--required-mfa`, `--session-policies`) and run-list
  `--orderby` (v4 run list has no server-side orderby; `--status` is filtered
  client-side).
- `search` index create/update `--monitored` / `--active` (v4 `IndexCreate`/
  `IndexUpdate` carry only display name + description).

The **auth** package (`login` already uses v4 GlobusApp; `device`, `logout`,
`refresh`, `tokens`, `whoami`) and `root.go` still reference the v3 SDK and the
transitional legacy-token bridge — a separate follow-up (Phase 2c).

## Coverage by area

| Area | Python CLI | Go CLI | Status |
|------|-----------|--------|--------|
| Login / logout / whoami | ✅ | ✅ (`auth login/logout/whoami`) | Covered |
| Identity lookup | `get-identities` | ✅ (`auth identities lookup`) | Covered |
| Transfer submit / ls / mkdir / rename / rm / stat | ✅ | Partial — `cp`, `ls`, `mkdir`, `rm`; **no `rename`, no `stat`** | Gap: rename, stat |
| Tasks (show/list/cancel/wait/event-list/pause-info/update) | ✅ (8) | Partial — show/list/cancel/wait; **no event-list, pause-info, update, generate-submission-id** | Gap |
| Endpoint search / show / update / delete | ✅ | Partial — search/show/list; **no update/delete** | Gap |
| Endpoint roles | ✅ (create/delete/list/show) | ❌ none | Gap |
| Endpoint permissions (ACLs) | ✅ (5) | ❌ none | Gap |
| Bookmarks | ✅ (5) | ❌ none | Gap |
| Collections / GCS management | ✅ (`collection`, `gcs`, 32 cmds) | ❌ none | Gap (no SDK support) |
| GCP (Connect Personal) | ✅ (6) | ❌ none | Gap (no SDK support) |
| Streams / tunnels | ✅ (8) | Stubs (report "unavailable") | Gap (not in v3 SDK) |
| Search (index/query/ingest/task/role/subject) | ✅ (14) | ✅ comparable | Covered |
| Groups (member/role/policy/invite/join/leave) | ✅ (19) | Partial — create/delete/list/show/update, member add/invite/list/remove, join/leave; **no group role commands** | Mostly covered |
| Flows (create/run/list/show/update/validate/logs) | ✅ (17) | ✅ comparable | Covered |
| Timers | ✅ (7) | ✅ (create/list/show/pause/resume/delete) | Covered |
| `api` raw passthrough | ✅ (7 services) | ❌ none | Gap |
| `session` (consent/show/update) | ✅ (3) | ❌ none | Gap |
| Compute | ❌ (not in Python CLI) | ✅ (endpoint/function/task) | Go-only extension |

## Gap classification

Two kinds of gap, important to distinguish:

### A. CLI gaps the Go SDK already supports (exposing them is CLI-only work)

The v3 SDK parity audit added the wire methods for most of these; the CLI simply
hasn't wired up commands yet:

- **Bookmarks** — SDK: `BookmarkList`/`CreateBookmark`/`GetBookmark`/
  `UpdateBookmark`/`DeleteBookmark`.
- **Endpoint roles** — SDK: `EndpointRoleList`/`AddEndpointRole`/
  `GetEndpointRole`/`DeleteEndpointRole`.
- **Endpoint ACLs / permissions** — SDK: `EndpointACLList`/`GetEndpointACLRule`/
  `AddEndpointACLRule`/`UpdateEndpointACLRule`/`DeleteEndpointACLRule`.
- **Endpoint update/delete, set-subscription-id, shared-endpoint list** — SDK:
  `UpdateEndpoint`/`DeleteEndpoint`/`SetSubscriptionID`/`MySharedEndpointList`/
  `GetSharedEndpointList`.
- **Transfer `stat`** — SDK: `OperationStat`. **`rename`** — SDK: `Rename`.
- **Task `event-list`/`pause-info`/`update`** — SDK: `TaskEventList`/
  `TaskPauseInfo`/`UpdateTask` (+ `TaskSuccessfulTransfers`/`TaskSkippedErrors`).
- **Endpoint-manager admin surface** — SDK: full `EndpointManager*` family.
- **Group roles**, **session/consents** — SDK: groups membership actions and
  auth `GetConsents`/`GetIdentities` exist.

### B. Gaps with no v3 SDK support (need SDK work first, or are out of scope)

- **GCS / collections management** (`collection`, `gcs`, ~32 commands) — the v4
  SDK module has a `gcs` service. Now that the service commands build on v4
  (Phase 2b), wiring these up is CLI work rather than an SDK blocker.
- **GCP (Globus Connect Personal)** — local-agent management; not an SDK API.
- **Streams / tunnels** — a Python globus-sdk v4.3.0+ feature. The v4 Go SDK
  transfer client now exposes tunnel/stream-access-point methods, so these are
  becoming supportable; the CLI currently ships "unavailable" stubs.
- **`api` raw passthrough** — convenience for arbitrary API calls; no SDK method,
  would be a thin HTTP wrapper.

## Compute is a Go-only extension

The Go CLI exposes a `compute` command tree; the Python CLI has **no** compute
commands (Globus Compute has its own separate `globus-compute` CLI). The Go
compute commands are limited by the Compute web API (no list/cancel/run of tasks
or functions server-side — those require the Globus Compute serialization SDK).

## Summary

The Go CLI covers the **core data-management workflows** (transfer, search,
groups, flows, timers, auth) at rough parity with the Python CLI's most-used
commands, plus a compute extension. The material gaps are:

1. **Endpoint administration** (roles, ACLs/permissions, update/delete) and
   **bookmarks** — all SDK-supported; CLI wiring is the remaining work.
2. **GCS/collections and GCP** — larger, need a v4-gcs path or are out of scope.
3. **`api` passthrough** and **`session`** — convenience features.
