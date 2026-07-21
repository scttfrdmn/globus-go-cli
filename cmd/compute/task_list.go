// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TaskListCmd represents the compute task list command
var TaskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List recent task executions.

NOTE: The Globus Compute API does not expose a task-listing endpoint, so this
command is not supported. Query individual tasks by ID with
"globus compute task show TASK_ID", or a batch with the task-group ID.`,
	RunE: runTaskList,
}

func runTaskList(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("listing tasks is not supported by the Globus Compute API; use \"globus compute task show TASK_ID\"")
}
