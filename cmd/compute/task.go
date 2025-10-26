// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"github.com/spf13/cobra"
)

// GetTaskCmd returns the task subcommand for compute
func GetTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage Globus Compute tasks",
		Long: `Commands for managing Globus Compute task executions.

Tasks represent individual executions of functions on specific endpoints.
You can run functions, monitor task status, and cancel running tasks.`,
	}

	// Add task subcommands
	taskCmd.AddCommand(TaskRunCmd)
	taskCmd.AddCommand(TaskShowCmd)
	taskCmd.AddCommand(TaskCancelCmd)
	taskCmd.AddCommand(TaskListCmd)

	return taskCmd
}
