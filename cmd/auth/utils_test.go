// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadToken tests the LoadToken function
func TestLoadToken(t *testing.T) {
	// Create a backup of the current function and restore it after the test
	origLoadTokenFunc := LoadTokenFunc
	defer func() {
		LoadTokenFunc = origLoadTokenFunc
	}()

	// Create a mock TokenInfo
	mockToken := &TokenInfo{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scopes:       []string{"openid", "profile", "email"},
	}

	// Replace the LoadTokenFunc with a mock implementation
	LoadTokenFunc = func(profile string) (*TokenInfo, error) {
		if profile == "test-profile" {
			return mockToken, nil
		}
		return nil, errors.New("test error")
	}

	// Test successful token loading
	token, err := LoadToken("test-profile")
	if err != nil {
		t.Errorf("LoadToken returned an error: %v", err)
	}
	if token.AccessToken != mockToken.AccessToken {
		t.Errorf("Expected access token %s, got %s", mockToken.AccessToken, token.AccessToken)
	}

	// Test token loading failure
	_, err = LoadToken("non-existent-profile")
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

// TestGetTokenFilePath tests the GetTokenFilePathFunc function
func TestGetTokenFilePath(t *testing.T) {
	// Create a backup of the current function and restore it after the test
	origGetTokenFilePathFunc := GetTokenFilePathFunc
	defer func() {
		GetTokenFilePathFunc = origGetTokenFilePathFunc
	}()

	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Error getting home directory: %v", err)
	}

	// Create a mock token file path
	expectedPath := filepath.Join(homeDir, ".globus-cli", "tokens", "test-profile.json")

	// Test the function
	tokenPath, err := GetTokenFilePathFunc("test-profile")
	if err != nil {
		t.Errorf("GetTokenFilePathFunc returned an error: %v", err)
	}
	if tokenPath != expectedPath {
		t.Errorf("Expected token path %s, got %s", expectedPath, tokenPath)
	}
}

// TestLoadTokenImpl tests the loadTokenImpl function
func TestLoadTokenImpl(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "globus-cli-test")
	if err != nil {
		t.Fatalf("Error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock token
	mockToken := &TokenInfo{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scopes:       []string{"openid", "profile", "email"},
	}

	// Create a tokens directory
	tokensDir := filepath.Join(tempDir, ".globus-cli", "tokens")
	if err := os.MkdirAll(tokensDir, 0755); err != nil {
		t.Fatalf("Error creating tokens directory: %v", err)
	}

	// Create a token file
	tokenPath := filepath.Join(tokensDir, "test-profile.json")
	tokenData, err := json.Marshal(mockToken)
	if err != nil {
		t.Fatalf("Error marshaling token: %v", err)
	}
	if err := os.WriteFile(tokenPath, tokenData, 0644); err != nil {
		t.Fatalf("Error writing token file: %v", err)
	}

	// Create a backup of the GetTokenFilePathFunc
	origGetTokenFilePathFunc := GetTokenFilePathFunc
	defer func() {
		GetTokenFilePathFunc = origGetTokenFilePathFunc
	}()

	// Replace GetTokenFilePathFunc with a mock implementation
	GetTokenFilePathFunc = func(profile string) (string, error) {
		if profile == "test-profile" {
			return tokenPath, nil
		}
		if profile == "error-profile" {
			return "", errors.New("test error")
		}
		return filepath.Join(tokensDir, profile+".json"), nil
	}

	// Test successful token loading
	token, err := loadTokenImpl("test-profile")
	if err != nil {
		t.Errorf("loadTokenImpl returned an error: %v", err)
	}
	if token.AccessToken != mockToken.AccessToken {
		t.Errorf("Expected access token %s, got %s", mockToken.AccessToken, token.AccessToken)
	}

	// Test error getting token file path
	_, err = loadTokenImpl("error-profile")
	if err == nil {
		t.Error("Expected an error getting token file path, got nil")
	}

	// Test non-existent token file
	_, err = loadTokenImpl("non-existent-profile")
	if err == nil {
		t.Error("Expected an error for non-existent token file, got nil")
	}
}

// TestUtilsIsTokenValid tests the IsTokenValid function
func TestUtilsIsTokenValid(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		token    *TokenInfo
		expected bool
	}{
		{
			name: "Valid token",
			token: &TokenInfo{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				ExpiresAt:    time.Now().Add(1 * time.Hour),
				Scopes:       []string{"openid", "profile", "email"},
			},
			expected: true,
		},
		{
			name: "Expired token",
			token: &TokenInfo{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				ExpiresAt:    time.Now().Add(-1 * time.Hour),
				Scopes:       []string{"openid", "profile", "email"},
			},
			expected: false,
		},
		{
			name:     "Nil token",
			token:    nil,
			expected: false,
		},
		{
			name: "Almost expired token",
			token: &TokenInfo{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				ExpiresAt:    time.Now().Add(4 * time.Minute), // Less than the 5-minute buffer
				Scopes:       []string{"openid", "profile", "email"},
			},
			expected: false,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsTokenValid(tc.token)
			if result != tc.expected {
				t.Errorf("Expected IsTokenValid to return %v, got %v", tc.expected, result)
			}
		})
	}
}
