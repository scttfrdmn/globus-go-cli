// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"github.com/spf13/cobra"
)

// GetRunCmd returns the run subcommand for flows
func GetRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Manage flow runs",
		Long: `Commands for managing flow run executions.

Flow runs represent individual executions of flows with specific inputs.
You can monitor runs, view their logs, and manage their lifecycle.`,
	}

	// Add run subcommands
	runCmd.AddCommand(RunListCmd)
	runCmd.AddCommand(RunShowCmd)
	runCmd.AddCommand(RunCancelCmd)
	runCmd.AddCommand(RunUpdateCmd)
	runCmd.AddCommand(RunShowLogsCmd)
	runCmd.AddCommand(RunShowDefinitionCmd)
	runCmd.AddCommand(RunDeleteCmd)
	runCmd.AddCommand(RunResumeCmd)

	return runCmd
}
