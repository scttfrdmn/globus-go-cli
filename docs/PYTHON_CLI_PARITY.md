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
| Collections / GCS management | ✅ (`collection`, `gcs`, 32 cmds) | ✅ core set — `collection list/show/create/update/delete`, `gcs info`, `gcs storage-gateway list/show`, `gcs role list/show/create/delete` | Covered (Phase 7) |
| GCP (Connect Personal) | ✅ (6) | ❌ none | Gap (no SDK support) |
| Streams / tunnels | ✅ (8) | ✅ (`tunnel list/show/create/update/delete/events`, `stream-access-point list/show`) | Covered (Phase 5) |
| Search (index/query/ingest/task/role/subject) | ✅ (14) | ✅ comparable — incl. `index role list/create/delete` and `task list` | Covered |
| Groups (member/role/policy/invite/join/leave) | ✅ (19) | ✅ — create/delete/list/show/update, member add/invite/list/remove/accept/decline/approve/reject, join/leave, `policies show/set` | Covered (Phase 4) |
| Flows (create/run/list/show/update/validate/logs) | ✅ (17) | ✅ comparable — incl. `validate`, `run delete`, `run resume` | Covered |
| Timers | ✅ (7) | ✅ (create/list/show/pause/resume/delete) | Covered |
| `api` raw passthrough | ✅ (7 services) | ✅ (`api auth/transfer/groups/search/flows/timer/compute`) | Covered (Phase 5) |
| `session` (consent/show/update) | ✅ (3) | ✅ — `session show` (via `include=session_info`), `session update` (step-up re-auth), `session consent` (scoped consent) | Covered (Phase 8) |
| Endpoint-manager (admin) | ✅ | ✅ (`endpoint-manager` — monitored-endpoints, task-list/show/cancel/pause/resume, pause-rule, ...) | Covered (Phase 6) |
| Endpoint set-subscription-id / my-shared-endpoint-list | ✅ | ✅ | Covered (Phase 6) |
| `list-commands` / `version` | ✅ | ✅ | Covered (Phase 6) |
| Compute | ❌ (not in Python CLI) | ✅ (endpoint/function/task) | Go-only extension |
| Project / client / credential management | ❌ (console-only in Python CLI) | ✅ (`project`, `project client`, `project credential` incl. rotation) | Go-only extension |

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

`session show/update/consent` is now DONE (Phase 8): `show` reads
`include=session_info`; `update` re-runs the login flow with
`session_required_*` params (SDK v4.8.1-4); `consent` runs a scoped login.

### B. Gaps with no v3 SDK support (need SDK work first, or are out of scope)

- **`api` raw passthrough** — DONE (Phase 5). `globus api <service> METHOD PATH`
  for auth/transfer/groups/search/flows/timer/compute, over the core client.
- **Streams / tunnels** — DONE (Phase 5). The v4 transfer client's
  tunnel/stream-access-point methods now back real `tunnel` and
  `stream-access-point` commands (the "unavailable" stubs are gone).
- **GCS / collections management** (`collection`, `gcs`) — DONE (Phase 7). The
  two former blockers were resolved with SDK support:
  1. the endpoint's **GCS Manager URL** is now on the transfer `Endpoint`
     (`gcs_manager_url`, SDK v4.8.1-2), so the CLI resolves the manager address
     from the endpoint ID via the Transfer API; and
  2. **dynamic consent escalation** — `pkg/globusauth.ScopedClientConfig` builds
     a GlobusApp for an endpoint/collection's dynamic scope and runs a consent
     login on first use. Management uses the endpoint's `manage_collections`
     scope (`gcs.EndpointManageCollectionsScope`, SDK v4.8.1-3).
  `collection list/show/create/update/delete` and `gcs info` +
  `storage-gateway`/`role` subcommands are wired. Each takes the owning
  `ENDPOINT_ID` as its first argument. `collection cat ENDPOINT_ID
  COLLECTION_ID PATH` reads a file over the collection's HTTPS data plane,
  using the collection's `https` scope (a separate per-collection data-access
  consent, escalated on first use).
- **GCP (Globus Connect Personal)** — local-agent management; not an SDK API
  (out of scope).

## Compute is a Go-only extension

The Go CLI exposes a `compute` command tree; the Python CLI has **no** compute
commands (Globus Compute has its own separate `globus-compute` CLI). The Go
compute commands are limited by the Compute web API (no list/cancel/run of tasks
or functions server-side — those require the Globus Compute serialization SDK).

## Project / client / credential management is a Go-only extension

The Go CLI's `project` command tree manages Globus Auth **projects**, their
registered **clients** (service accounts / app registrations), and client
**secret credentials** — the developer-console administrative surface. The
Python `globus-cli` has **no** commands for this (it is web-console only); the
capability is ported from the standalone `globus-project-manager` tool onto the
v4 SDK. Highlights:

- `project list/show/create/update/delete` and `project admin list/add/remove`.
- `project client list/show/create/update/delete` plus `update-redirect-uris`
  and `update-metadata`.
- `project credential list/create/delete` plus **rotation** (`rotate`,
  `list-age`, `process-deletions`) with a transition period. The Globus Auth API
  has no rotation primitive, so rotation state (linked new/old credentials +
  scheduled-deletion dates) is tracked in a per-profile local file
  `~/.globus-cli/credential-state-<profile>.json`.

These commands require the `manage_projects` scope, which the standard `globus
login` scope set omits; the CLI obtains it via a one-time consent (escalated on
first use), stored under a dedicated token namespace so it never collides with
the login token (both live on `auth.globus.org`).

## Summary

The Go CLI is now a **functional drop-in replacement** for the Python Globus
CLI across the full command surface, all on the v4 SDK: flat command paths,
matching global flags and JSON output (enveloped `DATA`), per-resource-server
GlobusApp auth, and command coverage spanning auth (incl. `session
show/update/consent`), transfer + endpoint administration + endpoint-manager,
streams/tunnels, groups, search, flows, timers, GCS `collection`/`gcs`, `api`
raw passthrough, and the `list-commands`/`version` meta commands — plus a
compute extension the Python CLI lacks.

The only remaining item is **GCP (Globus Connect Personal)** — local-agent
management with no SDK API, out of scope. GCS data-plane file access is started
(`collection cat`); more data-plane operations (e.g. HTTPS directory listing)
could follow using the same per-collection `https`-scope consent.

## Per-command flag parity

Each command's flags were audited against the installed Python `globus` CLI and
the gaps closed wherever the v4 SDK backs the underlying request field (flat
transfer ops, task, endpoint + collection/gcs admin, groups, search, flows,
timer, auth/session). Flags that the v4 SDK cannot express are intentionally
**not** added (no silent no-ops); the notable SDK-blocked flags are:

- **Auth provisioning** — `endpoint role/permission create --provision-identity`
  (needs Auth identity provisioning, not in the Transfer SDK).
- **`endpoint update --managed`** — requires resolving the caller's subscription
  ID; use `--subscription-id`.
- **`session update --all`** and `session/login --no-local-server` — no
  "add every identity" primitive, and the login flow is paste-code only.
- **`login --gcs/--flow/--timer`** and `session consent --timer-data-access` —
  build dynamic dependent scopes the fixed service registry doesn't model
  (the `project`/`collection` trees use scoped consent instead).
- **`timer create transfer --notify/--skip-source-errors/--fail-on-quota-errors`**
  — transfer-body extras not modeled by the timers schedule/body types.
- **`search query --bypass-visible-to/--filter-principal-sets`** and granular
  `task list --filter-task-id/type/label/date` — absent from the SDK options.
- **`flows run resume --skip-inactive-reason-check`**, `flows list --limit`
  (marker-paginated), and cosmetic `logout --yes/--ignore-errors`.

These are tracked so the absence is intentional and discoverable; each becomes
addable if/when the SDK grows the corresponding field.
