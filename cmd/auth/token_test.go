// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"testing"
	"time"
)

func TestIsTokenValid(t *testing.T) {
	tests := []struct {
		name        string
		token       *TokenInfo
		expectValid bool
	}{
		{
			name: "valid token",
			token: &TokenInfo{
				AccessToken: "valid-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			expectValid: true,
		},
		{
			name: "expired token",
			token: &TokenInfo{
				AccessToken: "expired-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			expectValid: false,
		},
		{
			name: "token expiring soon (within buffer)",
			token: &TokenInfo{
				AccessToken: "expiring-soon-token",
				ExpiresAt:   time.Now().Add(4 * time.Minute), // 5 minute buffer is used
			},
			expectValid: false,
		},
		{
			name: "token expiring soon (outside buffer)",
			token: &TokenInfo{
				AccessToken: "expiring-later-token",
				ExpiresAt:   time.Now().Add(6 * time.Minute), // 5 minute buffer is used
			},
			expectValid: true,
		},
		{
			name:        "nil token",
			token:       nil,
			expectValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := isTokenValid(tc.token)
			if valid != tc.expectValid {
				t.Errorf("isTokenValid() = %v, want %v", valid, tc.expectValid)
			}
		})
	}
}

// Test exported IsTokenValid too
func TestIsTokenValidExported(t *testing.T) {
	tests := []struct {
		name        string
		token       *TokenInfo
		expectValid bool
	}{
		{
			name: "valid token",
			token: &TokenInfo{
				AccessToken: "valid-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			expectValid: true,
		},
		{
			name: "expired token",
			token: &TokenInfo{
				AccessToken: "expired-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			expectValid: false,
		},
		{
			name:        "nil token",
			token:       nil,
			expectValid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := IsTokenValid(tc.token)
			if valid != tc.expectValid {
				t.Errorf("IsTokenValid() = %v, want %v", valid, tc.expectValid)
			}
		})
	}
}
