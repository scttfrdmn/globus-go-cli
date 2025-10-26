// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/scttfrdmn/globus-go-cli/cmd/flows"
	"github.com/spf13/cobra"
)

func getFlowsCommand() *cobra.Command {
	flowsCmd := &cobra.Command{
		Use:   "flows",
		Short: "Commands for Globus Flows",
		Long: `Commands for interacting with the Globus Flows service.

Globus Flows allows you to:
- Create and manage automated workflows
- Execute flows and monitor their progress
- Manage flow runs and view execution logs
- Configure flow permissions and visibility

Examples:
  # List your flows
  globus flows list

  # Show flow details
  globus flows show FLOW_ID

  # Start a flow with input
  globus flows start FLOW_ID --input input.json

  # Monitor a flow run
  globus flows run show RUN_ID

  # View run logs
  globus flows run show-logs RUN_ID`,
	}

	// Add subcommands
	flowsCmd.AddCommand(flows.ListCmd)
	flowsCmd.AddCommand(flows.ShowCmd)
	flowsCmd.AddCommand(flows.CreateCmd)
	flowsCmd.AddCommand(flows.UpdateCmd)
	flowsCmd.AddCommand(flows.DeleteCmd)
	flowsCmd.AddCommand(flows.StartCmd)
	flowsCmd.AddCommand(flows.ValidateCmd)
	flowsCmd.AddCommand(flows.GetRunCmd())

	return flowsCmd
}
