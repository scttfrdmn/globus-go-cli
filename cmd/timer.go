// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/scttfrdmn/globus-go-cli/cmd/timer"
	"github.com/spf13/cobra"
)

// getTimerCommand returns the root timer command
func getTimerCommand() *cobra.Command {
	timerCmd := &cobra.Command{
		Use:   "timer",
		Short: "Commands for Globus Timers",
		Long: `Commands for managing Globus Timers for scheduled tasks.

Globus Timers allows you to schedule recurring transfers and flow executions:
- Schedule periodic transfers between endpoints
- Run flows on a schedule
- Manage timer lifecycle (pause, resume, delete)

Examples:
  # List your timers
  globus timer list

  # Show timer details
  globus timer show TIMER_ID

  # Create a recurring transfer timer
  globus timer create transfer --source-endpoint ID:/path --dest-endpoint ID:/path --interval P1D

  # Create a flow timer (v3.39.0 feature)
  globus timer create flow FLOW_ID --interval P1D

  # Pause a timer
  globus timer pause TIMER_ID

  # Resume a timer
  globus timer resume TIMER_ID`,
	}

	// Add subcommands
	timerCmd.AddCommand(timer.ListCmd)
	timerCmd.AddCommand(timer.ShowCmd)
	timerCmd.AddCommand(timer.GetCreateCmd())
	timerCmd.AddCommand(timer.PauseCmd)
	timerCmd.AddCommand(timer.ResumeCmd)
	timerCmd.AddCommand(timer.DeleteCmd)

	return timerCmd
}
