// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import "testing"

// TestRefreshCmd verifies the refresh command is wired correctly. Refresh now
// reads the per-resource-server token store and uses the v4 auth client
// (refresh.go); the prior tests reimplemented the deleted single-token refresh
// logic against a mock client and no longer apply.
func TestRefreshCmd(t *testing.T) {
	cmd := RefreshCmd()

	if cmd.Use != "refresh" {
		t.Errorf("expected Use %q, got %q", "refresh", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}
