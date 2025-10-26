# Common Tasks

Practical examples for common operations with the Globus Go CLI.

## File Transfer

### Transfer a Single File

```bash
globus transfer transfer \
  --source-endpoint abc12345-6789-0def-ghij-klmnopqrstuv \
  --dest-endpoint xyz67890-abcd-efgh-ijkl-mnopqrstuvwx \
  --source-path /data/file.txt \
  --dest-path /backup/file.txt \
  --label "Single file transfer"
```

### Transfer a Directory Recursively

```bash
globus transfer transfer \
  --source-endpoint abc12345-6789-0def-ghij-klmnopqrstuv \
  --dest-endpoint xyz67890-abcd-efgh-ijkl-mnopqrstuvwx \
  --source-path /data/mydir/ \
  --dest-path /backup/mydir/ \
  --recursive \
  --label "Directory backup"
```

### Monitor Transfer Progress

```bash
# Get task ID from transfer command output
TASK_ID="01234567-89ab-cdef-0123-456789abcdef"

# Check task status
globus transfer task show $TASK_ID

# List all active tasks
globus transfer task list --filter-status ACTIVE
```

### Sync Directories

Use sync level to avoid re-transferring unchanged files:

```bash
globus transfer transfer \
  --source-endpoint abc12345-6789-0def-ghij-klmnopqrstuv \
  --dest-endpoint xyz67890-abcd-efgh-ijkl-mnopqrstuvwx \
  --source-path /data/ \
  --dest-path /backup/ \
  --recursive \
  --sync-level mtime
```

Sync levels:

- `exists` - Transfer only if file doesn't exist at destination
- `size` - Transfer if size differs
- `mtime` - Transfer if modification time differs
- `checksum` - Transfer if checksum differs (most thorough, slowest)

## Working with Endpoints

### Find an Endpoint

Search for endpoints by name or keyword:

```bash
globus transfer endpoint search "Tutorial"
```

### List Your Endpoints

```bash
globus transfer endpoint list --filter-scope my-endpoints
```

### Get Endpoint Details

```bash
globus transfer endpoint show abc12345-6789-0def-ghij-klmnopqrstuv
```

## Browsing Files

### List Directory Contents

```bash
globus transfer ls abc12345-6789-0def-ghij-klmnopqrstuv:/data/
```

### Show File Details

```bash
globus transfer ls abc12345-6789-0def-ghij-klmnopqrstuv:/data/file.txt --long
```

## Using Output with jq

### Extract Task ID

```bash
TASK_ID=$(globus transfer transfer \
  --source-endpoint SRC \
  --dest-endpoint DST \
  --source-path /src/path \
  --dest-path /dst/path \
  --format json | jq -r '.task_id')

echo "Task ID: $TASK_ID"
```

### Count Active Tasks

```bash
globus transfer task list --filter-status ACTIVE --format json | jq 'length'
```

### Extract Endpoint IDs

```bash
globus transfer endpoint list --format json | jq -r '.[].id'
```

## Automation and Scripting

### Batch Transfers

```bash
#!/bin/bash
FILES="file1.txt file2.txt file3.txt"

for file in $FILES; do
  globus transfer transfer \
    --source-endpoint $SRC_EP \
    --dest-endpoint $DST_EP \
    --source-path "/data/$file" \
    --dest-path "/backup/$file" \
    --label "Backup $file"
done
```

### Check and Wait for Completion

```bash
#!/bin/bash
TASK_ID=$1

while true; do
  STATUS=$(globus transfer task show $TASK_ID --format json | jq -r '.status')

  if [ "$STATUS" = "SUCCEEDED" ]; then
    echo "Transfer completed successfully"
    exit 0
  elif [ "$STATUS" = "FAILED" ]; then
    echo "Transfer failed"
    exit 1
  fi

  echo "Status: $STATUS - waiting..."
  sleep 10
done
```

## See Also

- [Transfer Commands](../reference/transfer.md)
- [Output Formats](output-formats.md)
- [Environment Variables](environment-variables.md)
