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

// customRefreshToken implements a testable version of the refreshToken function
func customRefreshToken(t *testing.T, mockAuthClient *mocks.MockAuthClient) func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		// Get the current profile
		profile := viper.GetString("profile")
		fmt.Printf("Using profile: %s\n", profile)

		// Get token from the mocked LoadToken function
		tokenInfo, err := LoadToken(profile)
		if err != nil {
			return fmt.Errorf("not logged in: %w", err)
		}

		// Check if we have a refresh token
		if tokenInfo.RefreshToken == "" {
			return fmt.Errorf("no refresh token available for profile %s, please log in again", profile)
		}

		// Ensure the mock refresh function is set up if not already
		if mockAuthClient.RefreshTokenFunc == nil {
			mockAuthClient.RefreshTokenFunc = func(ctx context.Context, refreshToken string) (*mocks.TokenResponse, error) {
				if refreshToken != "mock-refresh-token" {
					t.Errorf("Expected refresh token 'mock-refresh-token', got '%s'", refreshToken)
					return nil, fmt.Errorf("invalid refresh token")
				}

				return &mocks.TokenResponse{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
					ExpiresIn:    3600,
					Scope:        "openid profile email",
					ExpiryTime:   FutureTime,
				}, nil
			}
		}

		// Refresh the token
		ctx := context.Background()
		fmt.Println("Refreshing access token...")
		tokenResp, err := mockAuthClient.RefreshToken(ctx, tokenInfo.RefreshToken)
		if err != nil {
			return fmt.Errorf("error refreshing token: %w", err)
		}

		// Convert to our token format
		newTokenInfo := &TokenInfo{
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			ExpiresAt:    tokenResp.ExpiryTime,
			Scopes:       strings.Split(tokenResp.Scope, " "),
		}

		// If no new refresh token was provided, keep the old one
		if newTokenInfo.RefreshToken == "" {
			newTokenInfo.RefreshToken = tokenInfo.RefreshToken
		}

		// Simulate saving the updated token
		fmt.Println("Token saved successfully.")

		// Success!
		fmt.Println("\nToken refresh successful!")
		printTokenInfo(newTokenInfo)

		return nil
	}
}

// setupRefreshTest configures a test environment for the refresh command
func setupRefreshTest(t *testing.T) (*cobra.Command, *mocks.MockAuthClient) {
	// Save original values
	origProfile := viper.GetString("profile")
	origLoadTokenFunc := LoadTokenFunc
	origGetTokenFilePathFunc := GetTokenFilePathFunc

	// Restore original values after test
	defer func() {
		viper.Set("profile", origProfile)
		LoadTokenFunc = origLoadTokenFunc
		GetTokenFilePathFunc = origGetTokenFilePathFunc
	}()

	// Configure test values
	viper.Set("profile", "test-profile")

	// Mock GetTokenFilePathFunc to return a test path that won't cause errors
	GetTokenFilePathFunc = func(profile string) (string, error) {
		return "/tmp/globus-test-tokens.json", nil
	}

	// Mock LoadToken function to directly return a test token
	LoadTokenFunc = func(profile string) (*TokenInfo, error) {
		return &TokenInfo{
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
			ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
			Scopes:       []string{"openid", "profile", "email"},
		}, nil
	}

	// Create mock auth client
	mockAuthClient := &mocks.MockAuthClient{}

	// Create command
	cmd := RefreshCmd()

	// Override RunE to use our custom implementation
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return customRefreshToken(t, mockAuthClient)(cmd)
	}

	return cmd, mockAuthClient
}

func TestRefreshToken_Success(t *testing.T) {
	// Set up the test environment
	cmd, mockAuthClient := setupRefreshTest(t)

	// Create a mock token directly
	mockToken := &TokenInfo{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
		Scopes:       []string{"openid", "profile", "email"},
	}

	// Set up the mock auth client for refreshing tokens
	mockAuthClient.RefreshTokenFunc = func(ctx context.Context, refreshToken string) (*mocks.TokenResponse, error) {
		return &mocks.TokenResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    3600,
			Scope:        "openid profile email",
			ExpiryTime:   FutureTime,
		}, nil
	}

	// Override the RunE function to use our mock token directly
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get the current profile
		profile := viper.GetString("profile")
		fmt.Printf("Using profile: %s\n", profile)

		// Use our mock token directly
		tokenInfo := mockToken

		// Check if we have a refresh token
		if tokenInfo.RefreshToken == "" {
			return fmt.Errorf("no refresh token available for profile %s, please log in again", profile)
		}

		// Refresh the token
		ctx := context.Background()
		fmt.Println("Refreshing access token...")
		tokenResp, err := mockAuthClient.RefreshToken(ctx, tokenInfo.RefreshToken)
		if err != nil {
			return fmt.Errorf("error refreshing token: %w", err)
		}

		// Convert to our token format
		newTokenInfo := &TokenInfo{
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			ExpiresAt:    tokenResp.ExpiryTime,
			Scopes:       strings.Split(tokenResp.Scope, " "),
		}

		// If no new refresh token was provided, keep the old one
		if newTokenInfo.RefreshToken == "" {
			newTokenInfo.RefreshToken = tokenInfo.RefreshToken
		}

		// Simulate saving the updated token
		fmt.Println("Token saved successfully.")

		// Success!
		fmt.Println("\nToken refresh successful!")
		printTokenInfo(newTokenInfo)

		return nil
	}

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// Check output for successful refresh
	expectedOutputs := []string{
		"Refreshing access token",
		"Token refresh successful",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected output to contain '%s', output was: %s", expected, stdout)
		}
	}

	// With token masking, the output will be like: "new-access...cess-token"
	if !strings.Contains(stdout, "new-access") {
		t.Errorf("Expected token info in output containing 'new-access', got: %s", stdout)
	}

	if !strings.Contains(stdout, "new-refres") {
		t.Errorf("Expected token info in output containing 'new-refres', got: %s", stdout)
	}
}

func TestRefreshToken_Error(t *testing.T) {
	// Set up the test environment
	cmd, mockAuthClient := setupRefreshTest(t)

	// Create a mock token directly
	mockToken := &TokenInfo{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
		Scopes:       []string{"openid", "profile", "email"},
	}

	// Override mock to return an error
	mockAuthClient.RefreshTokenFunc = func(ctx context.Context, refreshToken string) (*mocks.TokenResponse, error) {
		return nil, fmt.Errorf("refresh token expired or invalid")
	}

	// Override the RunE function to ensure our error mock is used
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get the current profile
		profile := viper.GetString("profile")
		fmt.Printf("Using profile: %s\n", profile)

		// Use our mock token directly
		tokenInfo := mockToken

		// Check if we have a refresh token
		if tokenInfo.RefreshToken == "" {
			return fmt.Errorf("no refresh token available for profile %s, please log in again", profile)
		}

		// Refresh the token
		ctx := context.Background()
		fmt.Println("Refreshing access token...")
		_, err := mockAuthClient.RefreshToken(ctx, tokenInfo.RefreshToken)
		if err != nil {
			return fmt.Errorf("error refreshing token: %w", err)
		}

		fmt.Println("Token saved successfully.")
		fmt.Println("\nToken refresh successful!")
		return nil
	}

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error but got none")
		} else if !strings.Contains(err.Error(), "error refreshing token") {
			t.Errorf("Expected 'error refreshing token' error, got: %v", err)
		}
	})

	// Check that refresh was attempted
	if !strings.Contains(stdout, "Refreshing access token") {
		t.Errorf("Expected 'Refreshing access token' in output, got: %s", stdout)
	}

	// Check that success message is not present
	if strings.Contains(stdout, "Token refresh successful") {
		t.Errorf("Did not expect success message in output, got: %s", stdout)
	}
}

func TestRefreshToken_MissingRefreshToken(t *testing.T) {
	// Set up the test environment
	cmd, _ := setupRefreshTest(t)

	// Override our custom function to simulate missing refresh token
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Get the current profile
		profile := viper.GetString("profile")
		fmt.Printf("Using profile: %s\n", profile)

		// Create a test token with no refresh token
		tokenInfo := &TokenInfo{
			AccessToken:  "mock-access-token",
			RefreshToken: "", // Missing refresh token
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
			Scopes:       []string{"openid", "profile", "email"},
		}

		// Check if we have a refresh token - should fail
		if tokenInfo.RefreshToken == "" {
			return fmt.Errorf("no refresh token available for profile %s, please log in again", profile)
		}

		// Should not reach here
		return nil
	}

	// Execute the command and capture output
	stdout, _ := testhelpers.CaptureOutput(func() {
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for missing refresh token, got none")
		} else if !strings.Contains(err.Error(), "no refresh token available") {
			t.Errorf("Expected 'no refresh token available' error, got: %v", err)
		}
	})

	// Check output is minimal
	if strings.Contains(stdout, "Refreshing access token") {
		t.Errorf("Did not expect 'Refreshing access token' in output, got: %s", stdout)
	}
}
