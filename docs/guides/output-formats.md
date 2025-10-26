# Output Formats

The Globus Go CLI supports multiple output formats for different use cases.

## Available Formats

### JSON (Default)

Machine-readable JSON output, ideal for parsing and automation:

```bash
globus auth whoami --format json
```

Output:

```json
{
  "username": "user@example.edu",
  "name": "User Name",
  "email": "user@example.edu",
  "identity_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
}
```

### Text

Human-readable tabular output:

```bash
globus auth whoami --format text
```

Output:

```
Username      user@example.edu
Name          User Name
Email         user@example.edu
Identity ID   aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
```

## Setting Default Format

### Via Configuration

```bash
globus config set output_format json
```

### Via Environment Variable

```bash
export GLOBUS_CLI_FORMAT=json
```

### Via Flag

```bash
globus auth whoami --format text
```

## Parsing JSON Output

### Using jq

Extract specific fields:

```bash
# Get just the username
globus auth whoami --format json | jq -r '.username'

# Get identity ID
globus auth whoami --format json | jq -r '.identity_id'
```

List processing:

```bash
# Get all endpoint IDs
globus transfer endpoint list --format json | jq -r '.[].id'

# Get endpoint names and IDs
globus transfer endpoint list --format json | \
  jq -r '.[] | "\(.display_name): \(.id)"'
```

### Using Python

```python
import subprocess
import json

result = subprocess.run(
    ['globus', 'auth', 'whoami', '--format', 'json'],
    capture_output=True,
    text=True
)

data = json.loads(result.stdout)
print(f"Username: {data['username']}")
```

## Scripting Best Practices

### Always Specify Format

Don't rely on defaults in scripts:

```bash
# Good
globus auth whoami --format json

# Avoid
globus auth whoami
```

### Check Exit Codes

```bash
if globus auth whoami --format json > /dev/null 2>&1; then
  echo "Authenticated"
else
  echo "Not authenticated"
  exit 1
fi
```

### Handle Errors

```bash
output=$(globus transfer task show $TASK_ID --format json 2>&1)

if [ $? -ne 0 ]; then
  echo "Error: $output"
  exit 1
fi

echo "$output" | jq .
```

## See Also

- [Common Tasks](common-tasks.md)
- [Environment Variables](environment-variables.md)
