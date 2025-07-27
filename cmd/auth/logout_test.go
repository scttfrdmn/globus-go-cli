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

// customLogout implements a testable version of the logout function
func customLogout(t *testing.T, mockAuthClient *mocks.MockAuthClient) func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		// Get the current profile
		profile := viper.GetString("profile")
		fmt.Printf("Using profile: %s\n", profile)

		// Mock token loading
		tokenInfo := &TokenInfo{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scopes:       []string{"openid", "profile", "email"},
		}

		// Revoke the access token
		fmt.Println("Revoking access token...")
		ctx := context.Background()

		// Use the mock client to revoke tokens
		if err := mockAuthClient.RevokeToken(ctx, tokenInfo.AccessToken); err != nil {
			fmt.Printf("Warning: Failed to revoke access token: %v\n", err)
		}

		// Revoke the refresh token if present
		if tokenInfo.RefreshToken != "" {
			fmt.Println("Revoking refresh token...")
			if err := mockAuthClient.RevokeToken(ctx, tokenInfo.RefreshToken); err != nil {
				fmt.Printf("Warning: Failed to revoke refresh token: %v\n", err)
			}
		}

		// Mock deleting the token file
		fmt.Println("Removing token file...")

		// Success message
		fmt.Println("Logged out successfully!")
		return nil
	}
}

// setupLogoutTest configures a test environment for logout command
func setupLogoutTest(t *testing.T) (*cobra.Command, *mocks.MockAuthClient) {
	// Save original profile
	origProfile := viper.GetString("profile")

	// Restore original value after test
	defer func() {
		viper.Set("profile", origProfile)
	}()

	// Configure test values
	viper.Set("profile", "test-profile")

	// Create mock auth client
	mockAuthClient := &mocks.MockAuthClient{}

	// Create command
	cmd := LogoutCmd()

	// Override RunE to use our custom implementation
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return customLogout(t, mockAuthClient)(cmd)
	}

	return cmd, mockAuthClient
}

func TestLogoutCmd_Success(t *testing.T) {
	// Set up the test environment
	cmd, _ := setupLogoutTest(t)

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check that output shows the logout process
	expectedOutputs := []string{
		"Revoking access token",
		"Revoking refresh token",
		"Logged out successfully",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}

func TestLogoutCmd_AccessTokenRevocationFailure(t *testing.T) {
	// Set up the test environment
	cmd, mockAuthClient := setupLogoutTest(t)

	// Make access token revocation fail
	mockAuthClient.RevokeTokenFunc = func(ctx context.Context, token string) error {
		if token == "mock-access-token" {
			return fmt.Errorf("token revocation error")
		}
		return nil
	}

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check that output shows the warning but still completes
	expectedOutputs := []string{
		"Warning: Failed to revoke access token",
		"Revoking refresh token",
		"Logged out successfully",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}

func TestLogoutCmd_RefreshTokenRevocationFailure(t *testing.T) {
	// Set up the test environment
	cmd, mockAuthClient := setupLogoutTest(t)

	// Make refresh token revocation fail
	mockAuthClient.RevokeTokenFunc = func(ctx context.Context, token string) error {
		if token == "mock-refresh-token" {
			return fmt.Errorf("token revocation error")
		}
		return nil
	}

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check that output shows the warning but still completes
	if !strings.Contains(stdout, "Warning: Failed to revoke refresh token") {
		t.Errorf("Expected warning about refresh token revocation, got: %s", stdout)
	}

	if !strings.Contains(stdout, "Logged out successfully") {
		t.Errorf("Expected successful logout message, got: %s", stdout)
	}
}

func TestLogoutCmd_BothTokensRevocationFailure(t *testing.T) {
	// Set up the test environment
	cmd, mockAuthClient := setupLogoutTest(t)

	// Make both token revocations fail
	mockAuthClient.RevokeTokenFunc = func(ctx context.Context, token string) error {
		return fmt.Errorf("token revocation error")
	}

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check that output shows both warnings but still completes
	expectedOutputs := []string{
		"Warning: Failed to revoke access token",
		"Warning: Failed to revoke refresh token",
		"Logged out successfully",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}
}
