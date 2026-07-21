# Transfer Commands

Commands for file and directory transfer operations.

## Overview

The Transfer service provides commands for:

- Listing files and directories on endpoints
- Initiating file transfers between endpoints
- Managing transfer tasks
- Working with bookmarks

## Common Commands

### globus ls

List contents of a directory on an endpoint.

```bash
globus ls ENDPOINT_ID:/path [flags]
```

**Example:**

```bash
globus ls abc12345-6789-0def-ghij-klmnopqrstuv:/~/
```

### globus transfer

Initiate a file or directory transfer.

```bash
globus transfer [flags]
```

**Required Flags:**

- `--source-endpoint` - Source endpoint UUID
- `--dest-endpoint` - Destination endpoint UUID
- `--source-path` - Source file or directory path
- `--dest-path` - Destination path

**Optional Flags:**

- `--recursive` - Transfer directories recursively
- `--sync-level` - Sync level (exists, size, mtime, checksum)
- `--label` - Task label

**Example:**

```bash
globus transfer \
  --source-endpoint abc12345-6789-0def-ghij-klmnopqrstuv \
  --dest-endpoint xyz67890-abcd-efgh-ijkl-mnopqrstuvwx \
  --source-path /path/to/source \
  --dest-path /path/to/dest \
  --recursive
```

### globus task list

List recent transfer tasks.

```bash
globus task list [flags]
```

**Flags:**

- `--limit` - Maximum number of tasks to return
- `--filter-status` - Filter by status (ACTIVE, INACTIVE, SUCCEEDED, FAILED)

**Example:**

```bash
globus task list --limit 10 --filter-status ACTIVE
```

### globus task show

Show details for a specific transfer task.

```bash
globus task show TASK_ID
```

### globus task cancel

Cancel a transfer task.

```bash
globus task cancel TASK_ID
```

## Endpoint Commands

### globus endpoint list

List endpoints.

```bash
globus endpoint list [flags]
```

### globus endpoint show

Show details for an endpoint.

```bash
globus endpoint show ENDPOINT_ID
```

### globus endpoint search

Search for endpoints.

```bash
globus endpoint search SEARCH_TERMS [flags]
```

## See Also

- [Common Tasks](../guides/common-tasks.md)
- [Quick Start](../getting-started/quickstart.md)
