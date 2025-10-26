# Auth Commands

Commands for authentication and identity management.

## globus auth whoami

Display information about the currently authenticated user.

```bash
globus auth whoami [flags]
```

**Output:**

```json
{
  "username": "user@example.edu",
  "name": "User Name",
  "email": "user@example.edu",
  "identity_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
}
```

## globus login

Authenticate with Globus.

```bash
globus login [flags]
```

**Flags:**

- `--force` - Force re-authentication even if already logged in
- `--client-credentials` - Use client credentials flow

**Example:**

```bash
globus login
globus login --force
```

## globus logout

Remove stored authentication credentials.

```bash
globus logout
```

## globus session show

Display current session information including all linked identities.

```bash
globus session show
```

## globus session consent

Request consent for additional scopes.

```bash
globus session consent [flags]
```

**Example:**

```bash
globus session consent
```

## See Also

- [Authentication Guide](../getting-started/authentication.md)
- [Quick Start](../getting-started/quickstart.md)
