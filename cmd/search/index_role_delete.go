// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This command is a placeholder because the Go SDK v3.65.0-1 does not
// yet support role management for Search indices.

package search

import (
	"fmt"

	"github.com/spf13/cobra"
)

// IndexRoleDeleteCmd represents the search index role delete command
var IndexRoleDeleteCmd = &cobra.Command{
	Use:   "delete INDEX_ID ROLE_ID",
	Short: "Delete a role from a Globus Search index (not yet supported)",
	Long: `Delete a role from a Globus Search index.

NOTE: This command is not yet fully implemented because the Go SDK v3.65.0-1
does not support role management operations for Search indices.

Examples (when supported):
  # Delete a role
  globus search index role delete INDEX_ID ROLE_ID`,
	Args: cobra.ExactArgs(2),
	RunE: runIndexRoleDelete,
}

func runIndexRoleDelete(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("index role management is not yet available in SDK v3.65.0-1\n" +
		"Please use the Globus web interface or Python Globus CLI.")
}
