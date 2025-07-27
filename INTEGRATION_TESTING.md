# Integration Testing with Real Globus Credentials

This document explains how to set up real Globus credentials for integration testing with the Globus CLI.

## Prerequisites

1. A Globus account (register at [globus.org](https://www.globus.org/) if you don't have one)
2. Access to at least two Globus endpoints (source and destination) for transfer testing

## Setting Up Your Test Environment

### 1. Create a Globus App

1. Go to [developers.globus.org](https://developers.globus.org/)
2. Log in with your Globus account
3. Navigate to "Register your app"
4. Create a new app with:
   - Name: "Globus CLI Integration Test"
   - Project: Create a new project or use an existing one
   - Scopes: Request all the scopes you need for testing (auth, transfer, etc.)
   - Redirect URL: Use `https://auth.globus.org/v2/web/auth-code` for non-web apps
5. After creating the app, note down the **Client ID** and **Client Secret**

### 2. Set Up Your .env.test File

1. Copy the `.env.test.example` file to `.env.test` in the project root:
   ```
   cp .env.test.example .env.test
   ```

2. Edit the `.env.test` file and add your test credentials:
   ```
   # Globus Auth Credentials
   GLOBUS_TEST_CLIENT_ID=your_test_client_id
   GLOBUS_TEST_CLIENT_SECRET=your_test_client_secret
   
   # Test User Credentials (for non-interactive testing)
   GLOBUS_TEST_USERNAME=your_globus_username
   GLOBUS_TEST_PASSWORD=your_globus_password
   
   # Test Resource IDs (for transfer tests)
   GLOBUS_TEST_SOURCE_ENDPOINT=source_endpoint_id
   GLOBUS_TEST_DESTINATION_ENDPOINT=destination_endpoint_id
   GLOBUS_TEST_SOURCE_PATH=/source/path/
   GLOBUS_TEST_DESTINATION_PATH=/destination/path/
   
   # Test Identity (for identity lookup tests)
   GLOBUS_TEST_IDENTITY=your_globus_email_or_id
   ```

### 3. Finding Your Endpoint IDs

1. Log in to [app.globus.org](https://app.globus.org/)
2. Navigate to "File Manager" 
3. Find your source and destination endpoints
4. The endpoint ID is the UUID in the URL when you select an endpoint
5. Note these IDs down for your `.env.test` file

### 4. Setting Up Test Paths

1. Ensure your source and destination paths:
   - Exist on their respective endpoints
   - Have the appropriate permissions (you need read access for source, write access for destination)
   - For safety, use dedicated test directories that don't contain important data

## Running Integration Tests

Integration tests are tagged with the `integration` build tag. To run them:

```bash
# Run all integration tests
go test -tags=integration ./...

# Run specific package integration tests
go test -tags=integration ./cmd/auth/...
go test -tags=integration ./cmd/transfer/...
```

## Security Considerations

1. **NEVER** commit your `.env.test` file to version control
2. The `.env.test` file is listed in `.gitignore` to prevent accidental commits
3. Consider using a dedicated test account for integration tests
4. For CI/CD environments, use environment variables instead of the `.env.test` file

## Troubleshooting

1. **Token Expiration**: If tests fail with authentication errors, your tokens may have expired. Run the login command manually to refresh them.

2. **Endpoint Activation**: Ensure your endpoints are activated. If they require activation, you'll need to activate them through the Globus web interface before running tests.

3. **Permission Issues**: Verify you have the correct permissions on the test paths for both source and destination endpoints.

## Implementing More Integration Tests

When adding new integration tests:

1. Use the `integration` build tag at the top of your test file:
   ```go
   //go:build integration
   // +build integration
   ```

2. Use the testing helpers in `pkg/testhelpers` to manage credentials and setup:
   ```go
   // Load test credentials
   creds := testhelpers.LoadTestCredentials(t)
   
   // Skip if credentials not available
   creds.RequireTransferEndpoints(t)
   ```

3. Follow the examples in `cmd/auth/auth_integration_test.go` and `cmd/transfer/transfer_integration_test.go`