// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FutureTime is a helper for tests creating tokens that shouldn't be expired
var FutureTime = time.Now().Add(24 * time.Hour)

// LoadTokenFunc is a var that allows tests to replace the LoadToken implementation
var LoadTokenFunc = func(profile string) (*TokenInfo, error) {
	return loadTokenImpl(profile)
}

// GetTokenFilePathFunc is a var that allows tests to replace the getTokenFilePath implementation
var GetTokenFilePathFunc = func(profile string) (string, error) {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}

	// Create the token file path
	tokensDir := filepath.Join(homeDir, ".globus-cli", "tokens")
	tokenFile := filepath.Join(tokensDir, profile+".json")

	return tokenFile, nil
}

// LoadToken loads a token from disk (exported for use in other packages)
func LoadToken(profile string) (*TokenInfo, error) {
	return LoadTokenFunc(profile)
}

// loadTokenImpl is the actual implementation of LoadToken
func loadTokenImpl(profile string) (*TokenInfo, error) {
	// Get the token file path
	tokenFile, err := GetTokenFilePathFunc(profile)
	if err != nil {
		return nil, err
	}

	// Check if the token file exists
	if _, err := os.Stat(tokenFile); err != nil {
		return nil, fmt.Errorf("token file does not exist: %w", err)
	}

	// Read the token file
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("error reading token file: %w", err)
	}

	// Unmarshal the token
	var token TokenInfo
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("error unmarshaling token: %w", err)
	}

	return &token, nil
}

// IsTokenValid checks if a token is valid (exported for use in other packages)
func IsTokenValid(token *TokenInfo) bool {
	// Check if the token is expired
	// Add a 5-minute buffer to avoid edge cases
	return token != nil && time.Now().Add(5*time.Minute).Before(token.ExpiresAt)
}
