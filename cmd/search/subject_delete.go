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

// SubjectDeleteCmd represents the search subject delete command
var SubjectDeleteCmd = &cobra.Command{
	Use:   "delete INDEX_ID SUBJECT",
	Short: "Delete a subject from a Globus Search index",
	Long: `Submit a delete task to remove a subject from a Search index.

This requires writer or stronger privileges on the index. Deletions are
queued as tasks and are not guaranteed to be immediate.

Examples:
  # Delete a subject
  globus search subject delete INDEX_ID my-document-id

  # The command returns a task ID for monitoring progress
  globus search subject delete INDEX_ID doc123`,
	Args: cobra.ExactArgs(2),
	RunE: runSubjectDelete,
}

func runSubjectDelete(cmd *cobra.Command, args []string) error {
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

	// Execute delete. v4 deletes a single subject via DELETE
	// /index/{id}/subject?subject=...
	response, err := searchClient.DeleteSubject(ctx, indexID, subject)
	if err != nil {
		return fmt.Errorf("error deleting subject: %w", err)
	}

	// Display success message. The delete response carries the top-level task_id.
	fmt.Fprintf(os.Stdout, "Subject deletion task submitted!\n\n")
	fmt.Fprintf(os.Stdout, "Task ID:    %s\n", response.TaskID)
	fmt.Fprintf(os.Stdout, "Subject:    %s\n", subject)

	if response.TaskID != "" {
		fmt.Fprintf(os.Stdout, "\nCheck task status with: globus search task show %s\n", response.TaskID)
	}

	return nil
}
