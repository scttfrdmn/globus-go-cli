// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This command is a placeholder because the Go SDK v3.65.0-1 does not
// yet support role management for Search indices. The Python SDK has role
// management capabilities that need to be ported to the Go SDK.
//
// Related functionality exists in the Globus Search API but is not exposed
// in the current SDK version.

package search

import (
	"fmt"

	"github.com/spf13/cobra"
)

// IndexRoleListCmd represents the search index role list command
var IndexRoleListCmd = &cobra.Command{
	Use:   "list INDEX_ID",
	Short: "List roles on a Globus Search index (not yet supported)",
	Long: `List all roles and permissions on a Globus Search index.

NOTE: This command is not yet fully implemented because the Go SDK v3.65.0-1
does not support role management operations for Search indices.

The Python Globus CLI and SDK support role management, but this functionality
has not been ported to the Go SDK yet.

To manage index roles, please use:
- The Globus web interface at https://app.globus.org
- The Python Globus CLI: pip install globus-cli

Examples (when supported):
  # List roles on an index
  globus search index role list INDEX_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexRoleList,
}

func runIndexRoleList(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("index role management is not yet available in SDK v3.65.0-1\n" +
		"Please use the Globus web interface (https://app.globus.org) or\n" +
		"the Python Globus CLI to manage Search index roles.\n\n" +
		"The Go SDK will add role management support in a future release.")
}
