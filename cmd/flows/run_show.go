// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RunShowCmd represents the flows run show command
var RunShowCmd = &cobra.Command{
	Use:   "show RUN_ID",
	Short: "Show details of a flow run",
	Long: `Display detailed information about a specific flow run.

This includes the run's status, input, output, timestamps, and metadata.

Examples:
  # Show run details
  globus flows run show RUN_ID

  # Show run with JSON output
  globus flows run show RUN_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunShow,
}

func runFlowsRunShow(cmd *cobra.Command, args []string) error {
	runID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get run
	run, err := flowsClient.GetRun(ctx, runID, nil)
	if err != nil {
		return fmt.Errorf("error getting run: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Run Details\n")
		fmt.Printf("===========\n\n")

		fmt.Printf("Run ID:        %s\n", run.RunID)
		fmt.Printf("Flow ID:       %s\n", run.FlowID)
		if run.FlowTitle != "" {
			fmt.Printf("Flow Title:    %s\n", run.FlowTitle)
		}
		fmt.Printf("Status:        %s\n", run.Status)
		if run.Label != "" {
			fmt.Printf("Label:         %s\n", run.Label)
		}
		fmt.Printf("Owner:         %s\n", run.RunOwner)
		if !run.StartTime.IsZero() {
			fmt.Printf("Started:       %s\n", run.StartTime.Format(time.RFC3339))
		}
		if !run.EndTime.IsZero() {
			fmt.Printf("Completed:     %s\n", run.EndTime.Format(time.RFC3339))
		}

		// Display details
		if run.Details != nil {
			fmt.Printf("\nDetails:\n")
			detailsJSON, _ := json.MarshalIndent(run.Details, "  ", "  ")
			fmt.Printf("%s\n", string(detailsJSON))
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"RunID", "FlowID", "FlowTitle", "Status", "Label", "RunOwner", "StartTime", "EndTime"}
		if err := formatter.FormatOutput(run, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
