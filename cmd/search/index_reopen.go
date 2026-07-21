// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// IndexReopenCmd represents the search index reopen command.
// Added in Python SDK v4.0.0.
var IndexReopenCmd = &cobra.Command{
	Use:   "reopen INDEX_ID",
	Short: "Reopen a previously deleted Globus Search index",
	Long: `Reopen a previously deleted Globus Search index.

A deleted index can be reopened to restore access to its documents.
Added in Python SDK v4.0.0.

Examples:
  globus search index reopen INDEX_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexReopen,
}

func runIndexReopen(cmd *cobra.Command, args []string) error {
	indexID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Search client authorized for the current profile.
	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Reopen the index
	index, err := searchClient.ReopenIndex(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error reopening index: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Index reopened successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Index ID:     %s\n", index.ID)
	fmt.Fprintf(os.Stdout, "Display Name: %s\n", index.DisplayName)
	fmt.Fprintf(os.Stdout, "Status:       %s\n", index.Status)

	return nil
}
