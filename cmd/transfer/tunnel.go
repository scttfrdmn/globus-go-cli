// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TunnelCmd returns the tunnel subcommand group.
//
// Globus Streams tunnels are a Python SDK v4.3.0+ feature and are not part of
// the frozen v3 SDK this CLI builds against, so these commands are not
// available. The subcommands are retained so the help tree is stable, but each
// returns a clear "unsupported" error.
func TunnelCmd() *cobra.Command {
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Commands for managing Globus Streams tunnels (unavailable)",
		Long: `Commands for managing Globus Streams tunnels.

NOTE: Globus Streams is a newer API not included in the SDK version this CLI
builds against, so these commands are not available.`,
	}

	tunnelCmd.AddCommand(
		tunnelUnsupportedCmd("list", "List Globus Streams tunnels"),
		tunnelUnsupportedCmd("create", "Create a Globus Streams tunnel"),
		tunnelUnsupportedCmd("show", "Show a Globus Streams tunnel"),
		tunnelUnsupportedCmd("update", "Update a Globus Streams tunnel"),
		tunnelUnsupportedCmd("delete", "Delete a Globus Streams tunnel"),
		tunnelUnsupportedCmd("events", "Show events for a Globus Streams tunnel"),
	)

	return tunnelCmd
}

// tunnelUnsupportedCmd builds a placeholder subcommand that reports the feature
// is unavailable in this SDK version.
func tunnelUnsupportedCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short + " (unavailable)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("this SDK version does not support Globus Streams tunnels")
		},
	}
}
