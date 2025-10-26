// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package timer

import (
	"github.com/spf13/cobra"
)

// GetCreateCmd returns the create command
func GetCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new timer",
		Long: `Create a new Globus timer for recurring tasks.

Available timer types:
  transfer - Schedule recurring transfers between endpoints
  flow     - Schedule recurring flow executions (v3.39.0 feature)

Examples:
  # Create a recurring transfer timer
  globus timer create transfer --source SOURCE_EP:/path --dest DEST_EP:/path --interval P1D

  # Create a flow timer (v3.39.0)
  globus timer create flow FLOW_ID --interval P1D`,
	}

	// Add subcommands
	createCmd.AddCommand(CreateTransferCmd)
	createCmd.AddCommand(CreateFlowCmd)

	return createCmd
}
