// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LoadToken loads a token from disk (exported for use in other packages)
func LoadToken(profile string) (*TokenInfo, error) {
	// Get the token file path
	tokenFile, err := getTokenFilePath(profile)
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