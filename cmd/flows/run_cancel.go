// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// RunCancelCmd represents the flows run cancel command
var RunCancelCmd = &cobra.Command{
	Use:   "cancel RUN_ID",
	Short: "Cancel a flow run",
	Long: `Cancel an active flow run.

This attempts to gracefully cancel a running flow execution. The flow's
cancellation logic will be invoked if defined.

Examples:
  # Cancel a run
  globus flows run cancel RUN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunCancel,
}

func runFlowsRunCancel(cmd *cobra.Command, args []string) error {
	runID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Cancel run
	err = flowsClient.CancelRun(ctx, runID)
	if err != nil {
		return fmt.Errorf("error canceling run: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Run %s canceled successfully.\n", runID)
	fmt.Fprintf(os.Stdout, "\nCheck status with: globus flows run show %s\n", runID)

	return nil
}
