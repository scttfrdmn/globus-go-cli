// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SubjectShowCmd represents the search subject show command
var SubjectShowCmd = &cobra.Command{
	Use:   "show INDEX_ID SUBJECT",
	Short: "Show data for a subject in a Globus Search index",
	Long: `Display all entries for a specific subject in a Search index.

Only entries visible to you will be shown, based on the visible_to
access control list. If no entries are visible, an error is returned.

Examples:
  # Show subject details
  globus search subject show INDEX_ID my-document-id

  # Show with JSON output
  globus search subject show INDEX_ID doc123 --format json`,
	Args: cobra.ExactArgs(2),
	RunE: runSubjectShow,
}

func runSubjectShow(cmd *cobra.Command, args []string) error {
	indexID := args[0]
	subject := args[1]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Search client authorized for the current profile.
	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Fetch the subject directly. v4 exposes GET /index/{id}/subject?subject=...
	// (GetSubject), which returns the visible entries for the subject — a
	// cleaner match than the previous "subject:<id>" search query.
	result, err := searchClient.GetSubject(ctx, indexID, subject)
	if err != nil {
		return fmt.Errorf("error getting subject: %w", err)
	}

	// Check if subject was found
	if len(result.Entries) == 0 && len(result.Content) == 0 {
		return fmt.Errorf("subject '%s' not found or you don't have permission to view it", subject)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Subject: %s\n", result.Subject)
		fmt.Printf("========================================\n\n")

		// Display entries and their content
		if len(result.Entries) > 0 {
			for i, entry := range result.Entries {
				fmt.Printf("Entry %d", i+1)
				if entry.EntryID != "" {
					fmt.Printf(" (%s)", entry.EntryID)
				}
				fmt.Println(":")
				if entry.Content != nil {
					contentJSON, _ := json.MarshalIndent(entry.Content, "  ", "  ")
					fmt.Printf("  %s\n\n", string(contentJSON))
				}
			}
		} else if len(result.Content) > 0 {
			fmt.Printf("Content:\n")
			contentJSON, _ := json.MarshalIndent(result.Content, "  ", "  ")
			fmt.Printf("  %s\n\n", string(contentJSON))
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"Subject", "Entries", "Content"}
		if err := formatter.FormatOutput(result, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
