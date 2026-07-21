// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getVersionCommand returns the `version` command, matching the Python Globus
// CLI's `globus version`. It prints the CLI version (the same value surfaced by
// the root `--version` flag, set at build time).
func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the CLI version",
		Long:  "Print the version of the Globus CLI.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "globus-cli %s\n", Version)
			return nil
		},
	}
}
