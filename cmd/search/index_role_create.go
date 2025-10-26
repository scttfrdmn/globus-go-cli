// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This command is a placeholder because the Go SDK v3.65.0-1 does not
// yet support role management for Search indices.

package search

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	roleCreateName string
	roleCreateRole string
)

// IndexRoleCreateCmd represents the search index role create command
var IndexRoleCreateCmd = &cobra.Command{
	Use:   "create INDEX_ID PRINCIPAL",
	Short: "Create a role on a Globus Search index (not yet supported)",
	Long: `Create a new role granting permissions to a principal on an index.

NOTE: This command is not yet fully implemented because the Go SDK v3.65.0-1
does not support role management operations for Search indices.

Examples (when supported):
  # Grant admin role
  globus search index role create INDEX_ID user@example.com --role admin

  # Grant reader role
  globus search index role create INDEX_ID GROUP_ID --role reader`,
	Args: cobra.MinimumNArgs(2),
	RunE: runIndexRoleCreate,
}

func init() {
	IndexRoleCreateCmd.Flags().StringVar(&roleCreateRole, "role", "reader", "Role name (reader, writer, admin, owner)")
}

func runIndexRoleCreate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("index role management is not yet available in SDK v3.65.0-1\n" +
		"Please use the Globus web interface or Python Globus CLI.")
}
