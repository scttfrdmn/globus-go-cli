// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import "testing"

// TestLogoutCmd verifies the logout command is wired correctly. Logout now
// revokes every stored per-resource-server token and clears the store
// (logout.go); the prior tests reimplemented the deleted single-token
// revocation logic against a mock client and no longer apply.
func TestLogoutCmd(t *testing.T) {
	cmd := LogoutCmd()

	if cmd.Use != "logout" {
		t.Errorf("expected Use %q, got %q", "logout", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}
