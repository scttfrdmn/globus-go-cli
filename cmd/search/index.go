// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"github.com/spf13/cobra"
)

// GetIndexCmd returns the index command
func GetIndexCmd() *cobra.Command {
	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "Manage Globus Search indices",
		Long: `Commands for managing Globus Search indices.

Indices are containers for searchable documents. You can create indices,
manage their settings, control access with roles, and delete them when
no longer needed.

Examples:
  # List all your indices
  globus search index list

  # Create a new index
  globus search index create --display-name "My Research Data"

  # Show index details
  globus search index show INDEX_ID

  # Delete an index
  globus search index delete INDEX_ID`,
	}

	// Add subcommands
	indexCmd.AddCommand(IndexListCmd)
	indexCmd.AddCommand(IndexCreateCmd)
	indexCmd.AddCommand(IndexShowCmd)
	indexCmd.AddCommand(IndexUpdateCmd)
	indexCmd.AddCommand(IndexDeleteCmd)
	indexCmd.AddCommand(GetIndexRoleCmd())

	return indexCmd
}
