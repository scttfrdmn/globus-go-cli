// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This command is a placeholder because the Go SDK v3.65.0-1 does not
// yet support listing tasks for a Search index.

package search

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TaskListCmd represents the search task list command
var TaskListCmd = &cobra.Command{
	Use:   "list INDEX_ID",
	Short: "List recent tasks for an index (not yet supported)",
	Long: `List recent indexing and deletion tasks for a Globus Search index.

NOTE: This command is not yet fully implemented because the Go SDK v3.65.0-1
does not support listing tasks for an index.

You can still monitor individual tasks using: globus search task show TASK_ID

Examples (when supported):
  # List recent tasks
  globus search task list INDEX_ID

  # Limit results
  globus search task list INDEX_ID --limit 20`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskList,
}

func runTaskList(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("task listing is not yet available in SDK v3.65.0-1\n" +
		"You can still check individual task status using:\n" +
		"  globus search task show TASK_ID\n\n" +
		"The Go SDK will add task listing support in a future release.")
}
