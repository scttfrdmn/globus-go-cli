// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import "testing"

// TestWhoamiCmd verifies the whoami command is wired correctly. Whoami now
// resolves identity via the v4 auth client's userinfo endpoint (whoami.go); the
// prior tests reimplemented the deleted single-token logic against a mock
// client and no longer apply.
func TestWhoamiCmd(t *testing.T) {
	cmd := WhoamiCmd()

	if cmd.Use != "whoami" {
		t.Errorf("expected Use %q, got %q", "whoami", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}
