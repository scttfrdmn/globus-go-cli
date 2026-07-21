// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TaskCancelCmd represents the compute task cancel command
var TaskCancelCmd = &cobra.Command{
	Use:   "cancel TASK_ID",
	Short: "Cancel a running task",
	Long: `Cancel a task that is currently executing or pending.

NOTE: The Globus Compute API does not expose a task-cancellation endpoint,
so this command is not supported. Tasks are managed through the executing
Compute endpoint, not the web service.`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskCancel,
}

func runTaskCancel(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("task cancellation is not supported by the Globus Compute API")
}
