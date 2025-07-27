// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package testhelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TokenInfo represents token information stored in a token file
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
	TokenType    string    `json:"token_type"`
	Profile      string    `json:"profile"`
}

// SetupTokenFile creates a temporary token file structure for testing with valid tokens
// This is essential for tests that depend on token files to function
func SetupTokenFile(t *testing.T) (string, func()) {
	// Create a temporary home directory for testing
	homeDir, err := os.MkdirTemp("", "globus-cli-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}

	// Create the .globus-cli structure
	configDir := filepath.Join(homeDir, ".globus-cli")
	tokensDir := filepath.Join(configDir, "tokens")
	err = os.MkdirAll(tokensDir, 0700)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to create tokens directory: %v", err)
	}

	// Create the default profile token file
	token := TokenInfo{
		AccessToken:  "mock-access-token-for-testing",
		RefreshToken: "mock-refresh-token-for-testing",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "openid profile email urn:globus:auth:scope:transfer.api.globus.org:all",
		TokenType:    "Bearer",
		Profile:      "default",
	}

	// Marshal the token to JSON
	tokenJSON, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to marshal token JSON: %v", err)
	}

	// Write the token file
	tokenFilePath := filepath.Join(tokensDir, "default.json")
	err = os.WriteFile(tokenFilePath, tokenJSON, 0600)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to write token file: %v", err)
	}

	// Save original HOME environment variable
	originalHome := os.Getenv("HOME")

	// Set HOME to our temp directory for testing
	os.Setenv("HOME", homeDir)

	// Return the token file path and a cleanup function
	return tokenFilePath, func() {
		// Restore original HOME
		os.Setenv("HOME", originalHome)
		// Remove the temp directory
		os.RemoveAll(homeDir)
	}
}

// SetupTestTokenWithCustomProfile creates a token file for a custom profile
func SetupTestTokenWithCustomProfile(t *testing.T, profile string) (string, func()) {
	// Create a temporary home directory for testing
	homeDir, err := os.MkdirTemp("", "globus-cli-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}

	// Create the .globus-cli structure
	configDir := filepath.Join(homeDir, ".globus-cli")
	tokensDir := filepath.Join(configDir, "tokens")
	err = os.MkdirAll(tokensDir, 0700)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to create tokens directory: %v", err)
	}

	// Create the custom profile token file
	token := TokenInfo{
		AccessToken:  "mock-access-token-for-" + profile,
		RefreshToken: "mock-refresh-token-for-" + profile,
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "openid profile email urn:globus:auth:scope:transfer.api.globus.org:all",
		TokenType:    "Bearer",
		Profile:      profile,
	}

	// Marshal the token to JSON
	tokenJSON, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to marshal token JSON: %v", err)
	}

	// Write the token file
	tokenFilePath := filepath.Join(tokensDir, profile+".json")
	err = os.WriteFile(tokenFilePath, tokenJSON, 0600)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to write token file: %v", err)
	}

	// Save original HOME environment variable
	originalHome := os.Getenv("HOME")

	// Set HOME to our temp directory for testing
	os.Setenv("HOME", homeDir)

	// Return the token file path and a cleanup function
	return tokenFilePath, func() {
		// Restore original HOME
		os.Setenv("HOME", originalHome)
		// Remove the temp directory
		os.RemoveAll(homeDir)
	}
}

// SetupExpiredTestToken creates a token file with an expired token for testing token refresh
func SetupExpiredTestToken(t *testing.T) (string, func()) {
	// Create a temporary home directory for testing
	homeDir, err := os.MkdirTemp("", "globus-cli-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}

	// Create the .globus-cli structure
	configDir := filepath.Join(homeDir, ".globus-cli")
	tokensDir := filepath.Join(configDir, "tokens")
	err = os.MkdirAll(tokensDir, 0700)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to create tokens directory: %v", err)
	}

	// Create the default profile token file with expired token
	token := TokenInfo{
		AccessToken:  "mock-expired-access-token",
		RefreshToken: "mock-refresh-token-for-testing",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		Scope:        "openid profile email urn:globus:auth:scope:transfer.api.globus.org:all",
		TokenType:    "Bearer",
		Profile:      "default",
	}

	// Marshal the token to JSON
	tokenJSON, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to marshal token JSON: %v", err)
	}

	// Write the token file
	tokenFilePath := filepath.Join(tokensDir, "default.json")
	err = os.WriteFile(tokenFilePath, tokenJSON, 0600)
	if err != nil {
		os.RemoveAll(homeDir)
		t.Fatalf("Failed to write token file: %v", err)
	}

	// Save original HOME environment variable
	originalHome := os.Getenv("HOME")

	// Set HOME to our temp directory for testing
	os.Setenv("HOME", homeDir)

	// Return the token file path and a cleanup function
	return tokenFilePath, func() {
		// Restore original HOME
		os.Setenv("HOME", originalHome)
		// Remove the temp directory
		os.RemoveAll(homeDir)
	}
}