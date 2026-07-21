// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import "testing"

// TestLoginCmd verifies the login command is wired correctly: it uses the v4
// GlobusApp login flow (login.go) and registers its flags. The previous
// mock-reimplementation tests exercised a copy of the deleted single-token
// login logic and no longer apply.
func TestLoginCmd(t *testing.T) {
	cmd := LoginCmd()

	if cmd.Use != "login" {
		t.Errorf("expected Use %q, got %q", "login", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	for _, name := range []string{"scopes", "no-local-server", "no-save-tokens", "no-browser", "force"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}
}
