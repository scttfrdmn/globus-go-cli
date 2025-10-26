# Authentication

The Globus Go CLI uses OAuth2 authentication to securely access Globus services.

## Login

Authenticate with Globus:

```bash
globus login
```

This opens your browser for authentication and stores credentials locally.

## Authentication Flow

1. **Browser Opens** - The CLI opens your default browser
2. **Login** - Sign in with your institutional or Globus identity
3. **Grant Permissions** - Authorize the CLI to access Globus services
4. **Authorization Code** - Receive a code to paste back into the terminal
5. **Token Storage** - Credentials are stored in `~/.globus/`

## Checking Authentication Status

View your current identity:

```bash
globus auth whoami
```

Check session information:

```bash
globus session show
```

## Token Management

### Token Location

Tokens are stored in:

- **Linux/macOS**: `~/.globus/tokens.json`
- **Windows**: `%USERPROFILE%\.globus\tokens.json`

### Token Refresh

Tokens are automatically refreshed when needed. If refresh fails, you'll need to log in again:

```bash
globus login --force
```

## Multiple Identities

Link additional identities to your account:

```bash
globus session consent
```

View all linked identities:

```bash
globus session show
```

## Logout

Remove stored credentials:

```bash
globus logout
```

This removes all tokens from `~/.globus/tokens.json`.

## Client Credentials

For automation and scripts, use client credentials instead of user authentication.

### Setup

1. Register an application at [developers.globus.org](https://developers.globus.org)
2. Note your Client ID and Client Secret
3. Set environment variables:

```bash
export GLOBUS_CLIENT_ID="your-client-id"
export GLOBUS_CLIENT_SECRET="your-client-secret"
```

### Usage

```bash
globus login --client-credentials
```

## Scopes and Consent

The CLI requests these scopes:

- `openid` - Basic identity information
- `profile` - User profile details
- `email` - Email address
- `urn:globus:auth:scope:transfer.api.globus.org:all` - Transfer operations
- `urn:globus:auth:scope:search.api.globus.org:all` - Search operations
- Additional service-specific scopes as needed

## Troubleshooting

### Login Fails

If login fails:

1. Check your internet connection
2. Ensure browser opens correctly
3. Try `globus login --force` to force re-authentication
4. Clear tokens: `rm ~/.globus/tokens.json` and try again

### Token Expired

If you see "token expired" errors:

```bash
globus login --force
```

### Permission Denied

If you see permission errors:

1. Ensure you've consented to all required scopes
2. Try `globus session consent` to re-consent
3. Check that your identity has access to the requested resource

## Security Best Practices

- **Never share tokens** - Keep `~/.globus/tokens.json` private
- **Use client credentials** for automation
- **Logout when done** on shared systems
- **Rotate credentials** periodically for long-running scripts
- **Set file permissions**: `chmod 600 ~/.globus/tokens.json`

## Next Steps

- Configure the CLI in [Configuration](configuration.md)
- Try [Common Tasks](../guides/common-tasks.md)
- Review [Command Reference](../reference/index.md)
