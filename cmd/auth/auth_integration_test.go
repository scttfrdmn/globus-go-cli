// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

package auth

import (
	"testing"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
)

func TestAuthIntegration(t *testing.T) {
	// Load test credentials
	creds := testhelpers.LoadTestCredentials(t)

	// This test will be skipped if .env.test is not found or credentials aren't set
	t.Run("TestClientCredentialsFlow", func(t *testing.T) {
		// Example test using client credentials from .env.test
		// This is just a placeholder - replace with actual test logic
		if creds.ClientID == "" || creds.ClientSecret == "" {
			t.Skip("Client credentials not configured in .env.test")
		}

		// Test client credentials flow
		// Actual implementation would use the SDK to test authentication
		// ...

		// Verify the result
		// ...
	})

	t.Run("TestDeviceCodeFlow", func(t *testing.T) {
		// Skip this test in CI environments
		if testing.Short() {
			t.Skip("Skipping device code flow test in CI")
		}

		// Skip if no client credentials
		if creds.ClientID == "" || creds.ClientSecret == "" {
			t.Skip("Client credentials not configured in .env.test")
		}

		// Test device code flow
		// Actual implementation would use the SDK to test device code flow
		// ...

		// Verify the result
		// ...
	})
}