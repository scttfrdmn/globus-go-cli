// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"github.com/spf13/cobra"
)

// GetIndexRoleCmd returns the index role command
func GetIndexRoleCmd() *cobra.Command {
	roleCmd := &cobra.Command{
		Use:   "role",
		Short: "Manage Globus Search index roles",
		Long: `Commands for managing roles and permissions on Globus Search indices.

Roles control who can read, write, or administer an index.

Examples:
  # List roles on an index
  globus search index role list INDEX_ID

  # Create a new role
  globus search index role create INDEX_ID PRINCIPAL --role-name admin

  # Delete a role
  globus search index role delete INDEX_ID ROLE_ID`,
	}

	// Add subcommands
	roleCmd.AddCommand(IndexRoleListCmd)
	roleCmd.AddCommand(IndexRoleCreateCmd)
	roleCmd.AddCommand(IndexRoleDeleteCmd)

	return roleCmd
}
