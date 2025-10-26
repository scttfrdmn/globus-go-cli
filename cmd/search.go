// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/scttfrdmn/globus-go-cli/cmd/search"
	"github.com/spf13/cobra"
)

// getSearchCommand returns the root search command
func getSearchCommand() *cobra.Command {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Commands for Globus Search",
		Long: `Commands for interacting with the Globus Search service.

Globus Search allows you to:
- Create and manage search indices
- Ingest and query documents
- Manage index roles and permissions
- Monitor indexing tasks

Examples:
  # Query an index
  globus search query INDEX_ID --query "my search terms"

  # Ingest documents
  globus search ingest INDEX_ID --file documents.json

  # List your indices
  globus search index list

  # Create a new index
  globus search index create --display-name "My Index"`,
	}

	// Add subcommands
	searchCmd.AddCommand(search.QueryCmd)
	searchCmd.AddCommand(search.IngestCmd)
	searchCmd.AddCommand(search.GetIndexCmd())
	searchCmd.AddCommand(search.GetSubjectCmd())
	searchCmd.AddCommand(search.GetTaskCmd())

	return searchCmd
}
