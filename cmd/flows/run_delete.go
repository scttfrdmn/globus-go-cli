// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This command is a placeholder because the Go SDK v3.65.0-1 does not
// yet support deleting flow runs.

package flows

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RunDeleteCmd represents the flows run delete command
var RunDeleteCmd = &cobra.Command{
	Use:   "delete RUN_ID",
	Short: "Delete a flow run (not yet supported)",
	Long: `Delete a flow run and its associated data.

NOTE: This command is not yet fully implemented because the Go SDK v3.65.0-1
does not support deleting flow runs.

You can use the Globus web interface or Python CLI to delete runs:
  - Web interface: https://app.globus.org
  - Python CLI: globus flows run delete RUN_ID

Examples (when supported):
  # Delete a run
  globus flows run delete RUN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunDelete,
}

func runFlowsRunDelete(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("run deletion is not yet available in SDK v3.65.0-1\n" +
		"You can delete runs using:\n" +
		"  1. The Globus web interface (https://app.globus.org)\n" +
		"  2. The Python Globus CLI: globus flows run delete\n\n" +
		"The Go SDK will add run deletion support in a future release.")
}
