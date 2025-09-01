// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

func TestAuthIntegration(t *testing.T) {
	// Load test credentials
	creds := testhelpers.SkipIfNoCredentials(t)

	// Set up test config
	homeDir, err := os.MkdirTemp("", "globus-cli-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}
	defer os.RemoveAll(homeDir)

	// Set HOME environment variable to the test directory
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", origHome)

	// Create test token directory
	tokenDir := filepath.Join(homeDir, ".globus-cli", "tokens")
	if err := os.MkdirAll(tokenDir, 0700); err != nil {
		t.Fatalf("Failed to create token directory: %v", err)
	}

	t.Run("TestClientCredentialsFlow", func(t *testing.T) {
		// Create auth client with test credentials
		authClient, err := auth.NewClient(
			auth.WithClientID(creds.ClientID),
			auth.WithClientSecret(creds.ClientSecret),
		)
		if err != nil {
			t.Fatalf("Failed to create auth client: %v", err)
		}

		// Test client credentials flow
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get a token using client credentials  
		tokenResp, err := authClient.GetClientCredentialsToken(ctx, "openid", "email", "profile")
		if err != nil {
			t.Fatalf("Failed to get client credentials token: %v", err)
		}

		// Verify we got a valid token
		if tokenResp.AccessToken == "" {
			t.Fatal("Empty access token received")
		}
		if tokenResp.ExpiresIn <= 0 {
			t.Fatal("Invalid token expiration time")
		}

		// Introspect the token to verify it's valid
		introResp, err := authClient.IntrospectToken(ctx, tokenResp.AccessToken)
		if err != nil {
			t.Fatalf("Failed to introspect token: %v", err)
		}

		if !introResp.Active {
			t.Fatal("Token reported as inactive by introspection endpoint")
		}

		t.Logf("Successfully validated client credentials flow with token: %s...", tokenResp.AccessToken[:10])
	})

	// TODO: Re-enable when SDK provides identity lookup functionality
	// GetIdentities method is not available in current SDK v3.62.0-3
	/*
	t.Run("TestIdentityLookup", func(t *testing.T) {
		// Skip if no identity is provided
		if creds.TestIdentity == "" {
			t.Skip("Test identity not configured in .env.test")
		}

		// Create auth client with test credentials
		authClient, err := auth.NewClient(
			auth.WithClientID(creds.ClientID),
			auth.WithClientSecret(creds.ClientSecret),
		)
		if err != nil {
			t.Fatalf("Failed to create auth client: %v", err)
		}

		// Set up context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Look up identity
		identities, err := authClient.GetIdentities(ctx, creds.TestIdentity)
		if err != nil {
			t.Fatalf("Failed to get identity: %v", err)
		}

		// Check if we found any identities
		if len(identities) == 0 {
			t.Fatalf("No identities found for %s", creds.TestIdentity)
		}

		// Verify identity data
		identity := identities[0]
		if identity.Username == "" || identity.ID == "" {
			t.Fatal("Invalid identity returned: missing username or ID")
		}

		t.Logf("Successfully looked up identity: %s (%s)", identity.Username, identity.ID)
	})
	*/

	t.Run("TestDeviceCodeFlow", func(t *testing.T) {
		// Skip this test in CI environments
		if testing.Short() {
			t.Skip("Skipping device code flow test in CI")
		}

		// Skip this test as it requires user interaction
		t.Skip("Skipping device code flow test as it requires user interaction")

		// NOTE: Real device code flow testing requires user interaction
		// In a real scenario, this would:
		// 1. Start the device code flow
		// 2. Display the user code and verification URL
		// 3. Wait for the user to authenticate
		// 4. Verify the token received

		// For an actual automated integration test, this would need to be
		// replaced with a pre-authorized flow or a non-interactive method
	})
}
