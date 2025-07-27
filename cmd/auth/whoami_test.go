// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers/mocks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Custom whoami implementation that allows us to inject mock behavior
func customWhoami(t *testing.T, mockAuthClient *mocks.MockAuthClient) func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		// Get the current profile
		profile := viper.GetString("profile")
		fmt.Printf("Using profile: %s\n", profile)

		// Create a test token
		tokenInfo := &TokenInfo{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scopes:       []string{"openid", "profile", "email"},
		}

		// Validate token
		if !isTokenValid(tokenInfo) {
			return fmt.Errorf("token is expired, please login again")
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get user info from mock client
		introspection, err := mockAuthClient.IntrospectToken(ctx, tokenInfo.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to get user identity: %w", err)
		}

		// Print user information
		fmt.Println("\nCurrent User:")
		fmt.Printf("  Username: %s\n", introspection.Username)
		fmt.Printf("  Identity ID: %s\n", introspection.Subject)
		fmt.Printf("  Email: %s\n", introspection.Email)
		fmt.Printf("  Name: %s\n", introspection.Name)

		if len(introspection.IdentitySet) > 0 {
			fmt.Println("  Linked Identities:")
			for _, id := range introspection.IdentitySet {
				fmt.Printf("    - %s\n", id)
			}
		}

		// Print token information
		fmt.Println("\nToken Information:")
		fmt.Printf("  Expires At: %s\n", tokenInfo.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("  Expires In: %s\n", time.Until(tokenInfo.ExpiresAt).Round(time.Second))

		return nil
	}
}

// setupWhoamiTest creates a test environment for the whoami command
func setupWhoamiTest(t *testing.T) (*cobra.Command, *mocks.MockAuthClient) {
	// Save original viper state
	origProfile := viper.GetString("profile")
	defer viper.Set("profile", origProfile)

	// Set test profile
	viper.Set("profile", "test-profile")

	// Create a mock auth client
	mockAuthClient := &mocks.MockAuthClient{}

	// Set up the mock client with test data
	mockAuthClient.IntrospectTokenFunc = func(ctx context.Context, token string) (*mocks.TokenIntrospection, error) {
		// Expected token from our mocked loadToken
		if token != "mock-access-token" {
			t.Errorf("Expected token 'mock-access-token', got '%s'", token)
		}

		return &mocks.TokenIntrospection{
			Active:      true,
			Scope:       "openid profile email",
			ClientID:    "test-client-id",
			Username:    "test-user@example.org",
			Email:       "test-user@example.org",
			Name:        "Test User",
			Subject:     "test-subject-id",
			IdentitySet: []string{"id1", "id2"},
		}, nil
	}

	// Create a test whoami command
	cmd := WhoamiCmd()

	// Override the RunE function with our custom implementation
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return customWhoami(t, mockAuthClient)(cmd)
	}

	return cmd, mockAuthClient
}

func TestWhoamiCmd(t *testing.T) {
	cmd, _ := setupWhoamiTest(t)

	// Execute the command and capture output
	stdout, stderr := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check that output contains expected user information
	expectedOutputs := []string{
		"test-user@example.org",
		"test-subject-id",
		"Test User",
		"Linked Identities",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}

	// Stderr should be empty
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

func TestWhoamiCmd_InvalidToken(t *testing.T) {
	// Create the test command
	cmd := WhoamiCmd()

	// Override the command execution with a custom function that simulates an expired token
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Create an expired token
		expiredToken := &TokenInfo{
			AccessToken:  "expired-token",
			RefreshToken: "mock-refresh-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
			Scopes:       []string{"openid", "profile", "email"},
		}

		// This should fail with expired token
		if !isTokenValid(expiredToken) {
			return fmt.Errorf("token is expired, please login again")
		}

		// We shouldn't reach here
		return nil
	}

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for expired token, got nil")
		} else if !strings.Contains(err.Error(), "token is expired") {
			t.Errorf("Expected 'token is expired' error, got: %v", err)
		}
	})

	// Output should be minimal since we errored early
	if strings.Contains(stdout, "Current User") {
		t.Errorf("Didn't expect user output for expired token, got: %s", stdout)
	}
}
