// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"
	
	"github.com/scttfrdmn/globus-go-cli/cmd/auth"
)

// getAuthCommand returns the auth command
func getAuthCommand() *cobra.Command {
	// authCmd represents the auth command
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Commands for Globus Auth",
		Long: `Commands for working with Globus Auth including login, logout,
token management, and identity operations.`,
	}

	// Add auth subcommands
	authCmd.AddCommand(
		auth.LoginCmd(),
		auth.DeviceCmd(),
		auth.LogoutCmd(),
		auth.RefreshCmd(),
		auth.WhoamiCmd(),
		auth.TokensCmd(),
		auth.IdentitiesCmd(),
	)

	return authCmd
}