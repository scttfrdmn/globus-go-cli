// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	indexListLimit  int
	indexListOffset int
)

// IndexListCmd represents the search index list command
var IndexListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Globus Search indices",
	Long: `List all Globus Search indices where you have some permissions.

This shows indices you own, administer, or have access to.

Examples:
  # List all your indices
  globus search index list

  # Limit results
  globus search index list --limit 20

  # JSON output for scripting
  globus search index list --format json`,
	RunE: runIndexList,
}

func init() {
	// index_list is not paginated by the Search API, so these are no-ops kept
	// for backward compatibility.
	IndexListCmd.Flags().IntVar(&indexListLimit, "limit", 0, "Deprecated: index_list is not paginated")
	IndexListCmd.Flags().IntVar(&indexListOffset, "offset", 0, "Deprecated: index_list is not paginated")
	_ = IndexListCmd.Flags().MarkDeprecated("limit", "index_list is not paginated")
	_ = IndexListCmd.Flags().MarkDeprecated("offset", "index_list is not paginated")
}

func runIndexList(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Search client authorized for the current profile.
	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// List indices. The Search index_list endpoint is NOT paginated — sending
	// limit/offset returns HTTP 400 — so no options are passed. The --limit and
	// --offset flags are retained as deprecated no-ops for compatibility.
	indexList, err := searchClient.IndexList(ctx, nil)
	if err != nil {
		return fmt.Errorf("error listing indices: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(indexList.Indexes) == 0 {
			fmt.Println("No indices found.")
			return nil
		}

		fmt.Printf("%-36s  %-40s  %-12s  %-10s\n", "Index ID", "Display Name", "Status", "Entries")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"----------------------------------------",
			"------------",
			"----------")

		for _, index := range indexList.Indexes {
			displayName := index.DisplayName
			if len(displayName) > 40 {
				displayName = displayName[:37] + "..."
			}

			fmt.Printf("%-36s  %-40s  %-12s  %-10d\n",
				index.ID,
				displayName,
				index.Status,
				index.NumEntries)
		}

		fmt.Printf("\nTotal: %d index(es)\n", len(indexList.Indexes))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "DisplayName", "Description", "Status", "NumEntries", "NumSubjects"}
		if err := formatter.FormatOutput(indexList.Indexes, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
