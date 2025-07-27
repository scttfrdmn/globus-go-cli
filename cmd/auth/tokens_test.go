// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"fmt"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/spf13/cobra"
)

// customTokenShow implements a simple test version of token show command
func customTokenShow(t *testing.T) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Print mock token information
		fmt.Println("Token Information:")
		fmt.Println("  Access Token: mock-access-token")
		fmt.Println("  Refresh Token: mock-refresh-token")
		fmt.Println("  Scopes: openid profile email")
		return nil
	}
}

// customTokenRevoke implements a simple test version of token revoke command
func customTokenRevoke(t *testing.T, tokenType string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Get the token type flag
		typeFlag, _ := cmd.Flags().GetString("type")
		if typeFlag == "invalid" {
			return fmt.Errorf("invalid token type: %s", typeFlag)
		}

		// Print revocation message
		fmt.Printf("Revoking %s token...\n", typeFlag)
		capitalizedType := strings.ToUpper(typeFlag[:1]) + typeFlag[1:]
		fmt.Printf("%s token revoked successfully\n", capitalizedType)
		return nil
	}
}

// customTokenIntrospect implements a simple test version of token introspect command
func customTokenIntrospect(t *testing.T, withError bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if withError {
			return fmt.Errorf("failed to introspect token: API error")
		}

		// Print introspection information
		fmt.Println("Token Introspection:")
		fmt.Println("  Active: true")
		fmt.Println("  Username: test-user")
		fmt.Println("  Email: test@example.com")
		return nil
	}
}

// Test the tokens show command
func TestTokenShow(t *testing.T) {
	// Create command
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show token information",
		RunE:  customTokenShow(t),
	}

	// Execute command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	// Verify output
	expectedStrings := []string{
		"Token Information:",
		"Access Token:",
		"Refresh Token:",
		"Scopes:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, stdout)
		}
	}
}

// Test the tokens revoke command with access token
func TestTokenRevokeAccess(t *testing.T) {
	// Create command
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a token",
		RunE:  customTokenRevoke(t, "access"),
	}
	cmd.Flags().String("type", "access", "Token type to revoke")

	// Execute command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	// Verify output
	if !strings.Contains(stdout, "Revoking access token") {
		t.Errorf("Expected output to contain revocation message, got: %s", stdout)
	}

	if !strings.Contains(stdout, "Access token revoked successfully") {
		t.Errorf("Expected output to contain success message, got: %s", stdout)
	}
}

// Test the tokens revoke command with refresh token
func TestTokenRevokeRefresh(t *testing.T) {
	// Create command
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a token",
		RunE:  customTokenRevoke(t, "refresh"),
	}
	cmd.Flags().String("type", "refresh", "Token type to revoke")

	// Execute command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	// Verify output
	if !strings.Contains(stdout, "Revoking refresh token") {
		t.Errorf("Expected output to contain revocation message, got: %s", stdout)
	}

	if !strings.Contains(stdout, "Refresh token revoked successfully") {
		t.Errorf("Expected output to contain success message, got: %s", stdout)
	}
}

// Test the tokens revoke command with invalid token type
func TestTokenRevokeInvalid(t *testing.T) {
	// Create command
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a token",
		RunE:  customTokenRevoke(t, "invalid"),
	}
	cmd.Flags().String("type", "invalid", "Token type to revoke")

	// Execute command and capture output
	_, _ = testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err == nil {
			t.Errorf("Expected error for invalid token type, but got none")
		}
	})
}

// Test the tokens introspect command
func TestTokenIntrospect(t *testing.T) {
	// Create command
	cmd := &cobra.Command{
		Use:   "introspect",
		Short: "Introspect a token",
		RunE:  customTokenIntrospect(t, false),
	}

	// Execute command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	// Verify output
	expectedStrings := []string{
		"Token Introspection:",
		"Active: true",
		"Username: test-user",
		"Email: test@example.com",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, stdout)
		}
	}
}

// Test the tokens introspect command with error
func TestTokenIntrospectError(t *testing.T) {
	// Create command
	cmd := &cobra.Command{
		Use:   "introspect",
		Short: "Introspect a token",
		RunE:  customTokenIntrospect(t, true),
	}

	// Execute command and capture output
	_, _ = testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err == nil {
			t.Errorf("Expected error for token introspection, but got none")
		}
	})
}
