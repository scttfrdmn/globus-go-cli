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
token management, and identity operations.

Available Commands:
  login       Log in to Globus (web browser flow)
  device      Log in using device code flow (no browser)
  logout      Log out from Globus
  refresh     Refresh access tokens
  whoami      Show information about the current user
  tokens      Manage Globus Auth tokens
  identities  Look up and manage Globus identities

Examples:
  globus auth login                   Log in using a web browser
  globus auth device                  Log in using device code (no browser)
  globus auth whoami                  Show current user information
  globus auth tokens show             Show current token information
  globus auth identities lookup user  Look up a Globus identity by username`,
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