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

// TaskShowCmd represents the search task show command
var TaskShowCmd = &cobra.Command{
	Use:   "show TASK_ID",
	Short: "Show status of a Globus Search task",
	Long: `Display the status and details of a Globus Search indexing task.

Tasks are created when you ingest or delete documents. Use this command
to monitor the progress and check for errors.

Examples:
  # Show task status
  globus search task show TASK_ID

  # Show with JSON output
  globus search task show TASK_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskShow,
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Search client authorized for the current profile.
	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get task status. v4 GetTask returns the task envelope directly.
	taskStatus, err := searchClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("error getting task status: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Task Information\n")
		fmt.Printf("================\n\n")
		fmt.Printf("Task ID:    %s\n", taskStatus.TaskID)
		fmt.Printf("Index ID:   %s\n", taskStatus.IndexID)
		fmt.Printf("State:      %s\n", taskStatus.State)
		if !taskStatus.Created.IsZero() {
			fmt.Printf("Created At: %s\n", taskStatus.Created.Format(time.RFC3339))
		}
		if !taskStatus.Completed.IsZero() {
			fmt.Printf("Completed:  %s\n", taskStatus.Completed.Format(time.RFC3339))
		}
		if taskStatus.Message != "" {
			fmt.Printf("Message:    %s\n", taskStatus.Message)
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"TaskID", "IndexID", "State", "Created", "Completed", "Message"}
		if err := formatter.FormatOutput(taskStatus, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
