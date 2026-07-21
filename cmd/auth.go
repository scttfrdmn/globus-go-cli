// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-cli/cmd/auth"
)

// addAuthCommands wires the auth commands directly onto the root command as
// flat top-level commands, matching the Python Globus CLI (globus login,
// globus logout, globus whoami, globus get-identities, ...).
func addAuthCommands(root *cobra.Command) {
	root.AddCommand(
		auth.LoginCmd(),
		auth.LogoutCmd(),
		auth.WhoamiCmd(),
		auth.DeviceCmd(),
		auth.RefreshCmd(),
		auth.TokensCmd(),
		auth.GetIdentitiesCmd(),
	)
}
