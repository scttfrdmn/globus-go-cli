// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

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

// TaskShowCmd represents the compute task show command
var TaskShowCmd = &cobra.Command{
	Use:   "show TASK_ID",
	Short: "Show status and results of a task",
	Long: `Display detailed information about a task execution.

This includes the task's status, result (if completed), and any error
information if the task failed.

Examples:
  # Show task status
  globus compute task show TASK_ID

  # Show task with JSON output
  globus compute task show TASK_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskShow,
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Compute client authorized for the current profile.
	computeClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get task status (GET /v2/tasks/{id}); the response is an open-ended document.
	taskStatus, err := computeClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("error getting task status: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Task Details\n")
		fmt.Printf("============\n\n")

		id := mapStr(taskStatus, "task_id")
		if id == "" {
			id = taskID
		}
		fmt.Printf("Task ID:       %s\n", id)
		fmt.Printf("Status:        %s\n", mapStr(taskStatus, "status"))

		// Display result if available
		if result, ok := taskStatus["result"]; ok && result != nil {
			fmt.Printf("\nResult:\n")
			resultJSON, _ := json.MarshalIndent(result, "  ", "  ")
			fmt.Printf("%s\n", string(resultJSON))
		}

		// Display exception if task failed
		if ex := mapStr(taskStatus, "exception"); ex != "" {
			fmt.Printf("\nException:\n%s\n", ex)
		}
	} else {
		// JSON or CSV output — emit the raw passthrough document.
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"task_id", "status", "completion_t"}
		if err := formatter.FormatOutput(taskStatus, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
