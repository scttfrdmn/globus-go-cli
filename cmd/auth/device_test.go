// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import "testing"

// TestDeviceCmd verifies the device command is wired correctly. device.go now
// runs a real v4 OAuth2 device-code flow (no fake TEMPORARY_ACCESS_TOKEN stub),
// so the prior tests that executed the command and asserted the stub's
// "Login successful" output would require a live network and no longer apply.
func TestDeviceCmd(t *testing.T) {
	cmd := DeviceCmd()

	if cmd.Use != "device" {
		t.Errorf("expected Use %q, got %q", "device", cmd.Use)
	}
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	for _, name := range []string{"scopes", "no-save-tokens"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}
}
