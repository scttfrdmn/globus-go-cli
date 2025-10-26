// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"github.com/spf13/cobra"
)

// GetTaskCmd returns the task command
func GetTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Monitor Globus Search indexing tasks",
		Long: `Commands for monitoring the status of indexing and deletion tasks.

Tasks are created when you ingest documents or delete subjects. You can
check the status of these background operations using these commands.

Examples:
  # Show task status
  globus search task show TASK_ID

  # List recent tasks (if supported)
  globus search task list INDEX_ID`,
	}

	// Add subcommands
	taskCmd.AddCommand(TaskShowCmd)
	taskCmd.AddCommand(TaskListCmd)

	return taskCmd
}
