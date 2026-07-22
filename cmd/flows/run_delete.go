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

// RunDeleteCmd represents the flows run delete command
var RunDeleteCmd = &cobra.Command{
	Use:   "delete RUN_ID",
	Short: "Delete a flow run",
	Long: `Delete a flow run and its associated data.

Examples:
  # Delete a run
  globus flows run delete RUN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunDelete,
}

func runFlowsRunDelete(cmd *cobra.Command, args []string) error {
	runID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	if _, err := flowsClient.DeleteRun(ctx, runID); err != nil {
		return fmt.Errorf("error deleting run: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Run %s deleted.\n", runID)
	return nil
}
