<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors -->

# Privacy Policy

_Last updated: 2026-07-23_

`globus-go-cli` ("the CLI") is a free, open-source command-line client for the
[Globus](https://www.globus.org/) platform. This policy explains what data the
CLI handles. It applies to the CLI software itself; it does not govern the
Globus services the CLI talks to (see the
[Globus Privacy Policy](https://www.globus.org/legal/privacy)).

## Summary

The CLI runs entirely on your own machine. It communicates **only** with Globus
services, using credentials you obtain by logging in. It sends **no telemetry,
analytics, or usage data** to the CLI's authors or any third party.

## What the CLI stores locally

- **OAuth2 tokens.** After you log in, access and refresh tokens are stored on
  your machine under `~/.globus-cli/` (per profile, one entry per Globus
  resource server), with `0600`/`0700` permissions. These let the CLI act on
  your behalf without re-prompting.
- **Configuration.** Optional client ID/secret and preferences you set, under
  `~/.globus-cli/`.
- **Credential-rotation state.** For `project credential` rotation, a local file
  (`~/.globus-cli/credential-state-<profile>.json`) records which credentials
  were rotated and any scheduled-deletion dates. It does **not** store secret
  values.

You can remove all of this at any time with `globus logout` (which also revokes
your tokens with Globus Auth) or by deleting `~/.globus-cli/`.

## What the CLI sends, and to whom

- **Only to Globus service endpoints** — `auth.globus.org`,
  `transfer.api.globus.org`, `groups.api.globus.org`, `search.api.globus.org`,
  `flows.globus.org`, `timer.automate.globus.org`, `compute.api.globus.org`, and
  a Globus Connect Server's own manager/HTTPS hosts when you use `collection`/
  `gcs` commands — to carry out the operations you request.
- The data sent is the request you issue (e.g. a transfer's endpoints and paths,
  a search query, a group edit) plus your OAuth2 authorization.
- Requests carry the CLI's default OAuth2 client identifier so Globus can
  attribute API calls; this identifies the application, not you.

## What the CLI does **not** do

- No analytics, telemetry, crash reporting, or "phone home".
- No transmission of your data to the CLI's authors or any non-Globus third
  party.
- No advertising or tracking.

## Data controlled by Globus

Your identity, tokens, transfers, collections, and other resources are managed
by the Globus services. Their handling of that data is governed by the
[Globus Privacy Policy](https://www.globus.org/legal/privacy) and
[Terms of Service](https://www.globus.org/legal/terms), not by this document.

## Contact

Questions about this policy or the CLI: open an issue at
<https://github.com/scttfrdmn/globus-go-cli/issues>.
