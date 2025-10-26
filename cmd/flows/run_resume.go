// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This command is a placeholder because the Go SDK v3.65.0-1 does not
// yet support resuming flow runs.

package flows

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RunResumeCmd represents the flows run resume command
var RunResumeCmd = &cobra.Command{
	Use:   "resume RUN_ID",
	Short: "Resume a flow run (not yet supported)",
	Long: `Resume a paused or failed flow run.

NOTE: This command is not yet fully implemented because the Go SDK v3.65.0-1
does not support resuming flow runs.

You can use the Globus web interface or Python CLI to resume runs:
  - Web interface: https://app.globus.org
  - Python CLI: globus flows run resume RUN_ID

Examples (when supported):
  # Resume a paused run
  globus flows run resume RUN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunResume,
}

func runFlowsRunResume(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("run resume is not yet available in SDK v3.65.0-1\n" +
		"You can resume runs using:\n" +
		"  1. The Globus web interface (https://app.globus.org)\n" +
		"  2. The Python Globus CLI: globus flows run resume\n\n" +
		"The Go SDK will add run resume support in a future release.")
}
