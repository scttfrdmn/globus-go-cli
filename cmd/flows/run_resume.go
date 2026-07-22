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

// RunResumeCmd represents the flows run resume command
var RunResumeCmd = &cobra.Command{
	Use:   "resume RUN_ID",
	Short: "Resume a flow run",
	Long: `Resume a paused or inactive flow run.

Examples:
  # Resume a run
  globus flows run resume RUN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunResume,
}

func runFlowsRunResume(cmd *cobra.Command, args []string) error {
	runID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	if _, err := flowsClient.ResumeRun(ctx, runID); err != nil {
		return fmt.Errorf("error resuming run: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Run %s resumed.\n", runID)
	fmt.Fprintf(os.Stdout, "\nCheck status with: globus flows run show %s\n", runID)
	return nil
}
