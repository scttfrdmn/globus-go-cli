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

// IndexShowCmd represents the search index show command
var IndexShowCmd = &cobra.Command{
	Use:   "show INDEX_ID",
	Short: "Show details for a Globus Search index",
	Long: `Display detailed information about a Globus Search index.

This shows the index configuration, status, and metadata.

Examples:
  # Show index details
  globus search index show INDEX_ID

  # Show with JSON output
  globus search index show INDEX_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexShow,
}

func runIndexShow(cmd *cobra.Command, args []string) error {
	indexID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Search client authorized for the current profile.
	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get index details
	index, err := searchClient.GetIndex(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error getting index: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Index Information\n")
		fmt.Printf("=================\n\n")
		fmt.Printf("Index ID:     %s\n", index.ID)
		fmt.Printf("Display Name: %s\n", index.DisplayName)
		if index.Description != "" {
			fmt.Printf("Description:  %s\n", index.Description)
		}
		fmt.Printf("Status:       %s\n", index.Status)
		fmt.Printf("Entries:      %d\n", index.NumEntries)
		fmt.Printf("Subjects:     %d\n", index.NumSubjects)
		fmt.Printf("Size:         %.2f MB\n", index.SizeInMB)
		if index.MaxSizeInMB > 0 {
			fmt.Printf("Max Size:     %d MB\n", index.MaxSizeInMB)
		}
		if index.SubscriptionID != "" {
			fmt.Printf("Subscription: %s\n", index.SubscriptionID)
		}

		fmt.Printf("\nMetadata\n")
		fmt.Printf("--------\n")
		if !index.Created.IsZero() {
			fmt.Printf("Created At:   %s\n", index.Created.Format(time.RFC3339))
		}
		if !index.LastModified.IsZero() {
			fmt.Printf("Updated At:   %s\n", index.LastModified.Format(time.RFC3339))
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "DisplayName", "Description", "Status", "NumEntries", "NumSubjects"}
		if err := formatter.FormatOutput(index, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
