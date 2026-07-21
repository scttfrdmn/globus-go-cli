<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors -->

# Parity with the upstream Python Globus CLI

This tracks the Go CLI's command coverage against the reference
[Globus CLI (Python)](https://github.com/globus/globus-cli).

As of **Phase 3**, the Go CLI uses a **flat command structure matching the
Python CLI**: auth and transfer operations are top-level commands, not nested
under `auth`/`transfer` groups. So `globus login`, `globus logout`,
`globus whoami`, `globus get-identities`, `globus ls`, `globus mkdir`,
`globus rm`, `globus transfer SRC DEST`, `globus task show`, and
`globus endpoint search` are all invoked exactly as in the Python CLI.
`group`, `search`, `flows`, and `timer` remain command groups (the Python CLI
groups these too); `compute` is a Go-only extension group.

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

As of **Phase 2c**, the auth package (`login`, `device`, `logout`, `refresh`,
`tokens`, `whoami`, `identities`) and `root.go` are also on v4, and the
transitional legacy-token bridge has been removed. The CLI now depends solely
on the v4 SDK (v3 dropped from `go.mod`). `device` runs a real OAuth2
device-code flow and `identities lookup` performs a real `GetIdentities` call
(both were previously stubs).

## Coverage by area

| Area | Python CLI | Go CLI | Status |
|------|-----------|--------|--------|
| Login / logout / whoami | ✅ | ✅ (`login`/`logout`/`whoami`) | Covered |
| Identity lookup | `get-identities` | ✅ (`get-identities`) | Covered |
| Transfer submit / ls / mkdir / rename / rm / stat | ✅ | ✅ — `transfer`, `ls`, `mkdir`, `rm`, `rename`, `stat`, `delete` | Covered (Phase 4) |
| Tasks (show/list/cancel/wait/event-list/pause-info/update) | ✅ (8) | ✅ — show/list/cancel/wait/event-list/pause-info/update | Covered (Phase 4) |
| Endpoint search / show / update / delete | ✅ | ✅ — search/show/list/update/delete | Covered (Phase 4) |
| Endpoint roles | ✅ (create/delete/list/show) | ✅ (`endpoint role list/show/create/delete`) | Covered (Phase 4) |
| Endpoint permissions (ACLs) | ✅ (5) | ✅ (`endpoint permission list/show/create/update/delete`) | Covered (Phase 4) |
| Bookmarks | ✅ (5) | ✅ (`bookmark list/show/create/rename/delete`) | Covered (Phase 4) |
| Collections / GCS management | ✅ (`collection`, `gcs`, 32 cmds) | ❌ none | Blocked — SDK lacks endpoint GCS-manager URL + per-collection consent (see below) |
| GCP (Connect Personal) | ✅ (6) | ❌ none | Gap (no SDK support) |
| Streams / tunnels | ✅ (8) | ✅ (`tunnel list/show/create/update/delete/events`, `stream-access-point list/show`) | Covered (Phase 5) |
| Search (index/query/ingest/task/role/subject) | ✅ (14) | ✅ comparable | Covered |
| Groups (member/role/policy/invite/join/leave) | ✅ (19) | ✅ — create/delete/list/show/update, member add/invite/list/remove/accept/decline/approve/reject, join/leave, `policies show/set` | Covered (Phase 4) |
| Flows (create/run/list/show/update/validate/logs) | ✅ (17) | ✅ comparable | Covered |
| Timers | ✅ (7) | ✅ (create/list/show/pause/resume/delete) | Covered |
| `api` raw passthrough | ✅ (7 services) | ✅ (`api auth/transfer/groups/search/flows/timer/compute`) | Covered (Phase 5) |
| `session` (consent/show/update) | ✅ (3) | ❌ none | Gap — needs reauth/session-boundary support not in the v4 SDK (only `GetConsents`) |
| Endpoint-manager (admin) | ✅ | ✅ (`endpoint-manager` — monitored-endpoints, task-list/show/cancel/pause/resume, pause-rule, ...) | Covered (Phase 6) |
| Endpoint set-subscription-id / my-shared-endpoint-list | ✅ | ✅ | Covered (Phase 6) |
| `list-commands` / `version` | ✅ | ✅ | Covered (Phase 6) |
| Compute | ❌ (not in Python CLI) | ✅ (endpoint/function/task) | Go-only extension |

**JSON output shape (Phase 6):** list commands emit the enveloped service
document under `-F json` (`{"DATA_TYPE": ..., "DATA": [...]}`), matching the
Python CLI, rather than a bare array. (Groups list is a bare array in both CLIs.)

## Gap classification

Two kinds of gap, important to distinguish:

### A. CLI gaps the Go SDK already supports (exposing them is CLI-only work)

**Phase 4 wired up most of these** — now covered: bookmarks
(`bookmark list/show/create/rename/delete`), endpoint roles
(`endpoint role list/show/create/delete`), endpoint ACLs
(`endpoint permission list/show/create/update/delete`), endpoint
`update`/`delete`, transfer `rename`/`stat`/`delete`, task
`event-list`/`pause-info`/`update`, and group membership actions
(`member accept/decline/approve/reject`, real `join`/`leave`/`invite`) plus
`group policies show/set`.

Still available in the SDK but not yet wired:

- **Endpoint `set-subscription-id`, shared-endpoint list** — SDK:
  `SetSubscriptionID`/`MySharedEndpointList`/`GetSharedEndpointList`.
- **Endpoint-manager admin surface** — SDK: full `EndpointManager*` family
  (monitored endpoints, admin task list/cancel/pause, pause rules).
- **`session` show/update/consent** — the Python CLI's session commands drive a
  high-assurance **reauthentication / session-boundary** flow. The v4 Go SDK
  exposes `GetConsents` but no reauth/session-update endpoint, so a faithful
  `session` is deferred pending SDK support (see Phase 6 / SDK work).

### B. Gaps with no v3 SDK support (need SDK work first, or are out of scope)

- **`api` raw passthrough** — DONE (Phase 5). `globus api <service> METHOD PATH`
  for auth/transfer/groups/search/flows/timer/compute, over the core client.
- **Streams / tunnels** — DONE (Phase 5). The v4 transfer client's
  tunnel/stream-access-point methods now back real `tunnel` and
  `stream-access-point` commands (the "unavailable" stubs are gone).
- **GCS / collections management** (`collection`, `gcs`, ~32 commands) —
  **still blocked**, and NOT merely CLI wiring. The v4 SDK's `gcs.CollectionClient`
  needs two things the rest of the CLI doesn't have:
  1. the endpoint's **GCS Manager URL** (`https://<gcs-host>/api`), which is not
     exposed on the v4 SDK's transfer `Endpoint` struct — there is no field to
     read it from, so the client address can't be constructed from a collection
     ID alone; and
  2. **dynamic per-collection consent** — each collection has its own
     `.../<collection-id>/https` and `.../data_access` scopes, but the CLI's
     GlobusApp login requests only a fixed scope set, so `GetAuthorizer` has no
     token for a given collection's data-access scope.
  Wiring `collection`/`gcs` needs SDK support to surface the manager URL (e.g. on
  the endpoint document) and a consent/scope-escalation path in `pkg/globusauth`.
  Deferred to a future phase (tracked as SDK work).
- **GCP (Globus Connect Personal)** — local-agent management; not an SDK API
  (out of scope).

## Compute is a Go-only extension

The Go CLI exposes a `compute` command tree; the Python CLI has **no** compute
commands (Globus Compute has its own separate `globus-compute` CLI). The Go
compute commands are limited by the Compute web API (no list/cancel/run of tasks
or functions server-side — those require the Globus Compute serialization SDK).

## Summary

The Go CLI covers the **core data-management workflows** (transfer, search,
groups, flows, timers, auth) at parity with the Python CLI's most-used commands,
plus a compute extension. After Phase 4, endpoint administration (roles,
ACLs/permissions, update/delete), bookmarks, transfer `rename`/`stat`/`delete`,
task `event-list`/`pause-info`/`update`, and the full group membership + policy
surface are all wired. The remaining gaps are:

1. **GCS/collections** (`collection`, `gcs`) — Phase 5, via the v4 `gcs` service.
2. **`api` raw passthrough** — Phase 5, a thin HTTP wrapper.
3. **Streams/tunnels** — v4 transfer client has the methods; Phase 5 can replace
   the "unavailable" stubs.
4. **`session`** — needs reauth/session-boundary SDK support not yet available.
5. **GCP (Connect Personal)** — local-agent management; not an SDK API
   (out of scope).
