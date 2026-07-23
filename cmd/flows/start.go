// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
	"github.com/spf13/cobra"
)

var (
	startInputFile string
	startInputJSON string
	startLabel     string
	startTags      []string
	startManagers  []string
	startMonitors  []string
	startNotify    []string
	startWait      bool
)

// StartCmd represents the flows start command
var StartCmd = &cobra.Command{
	Use:   "start FLOW_ID",
	Short: "Start a flow execution",
	Long: `Start a new execution (run) of a flow with the specified input.

The input must conform to the flow's input schema. You can provide input
from a JSON file or as a JSON string on the command line.

Examples:
  # Start a flow from an input file
  globus flows start FLOW_ID --input-file input.json

  # Start with inline JSON input
  globus flows start FLOW_ID --input '{"param1": "value1"}'

  # Start with label and tags
  globus flows start FLOW_ID --input-file input.json \\
    --label "Production run" \\
    --tags "production,automated"

  # Start and wait for completion
  globus flows start FLOW_ID --input-file input.json --wait`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsStart,
}

func init() {
	StartCmd.Flags().StringVar(&startInputFile, "input-file", "", "Path to input JSON file")
	StartCmd.Flags().StringVar(&startInputJSON, "input", "", "Input as JSON string")
	StartCmd.Flags().StringVar(&startLabel, "label", "", "Label for this run")
	StartCmd.Flags().StringSliceVar(&startTags, "tags", []string{}, "Comma-separated tags")
	StartCmd.Flags().StringArrayVar(&startManagers, "manager", nil, "A principal that may manage the execution of the run (repeatable)")
	StartCmd.Flags().StringArrayVar(&startMonitors, "monitor", nil, "A principal that may monitor the execution of the run (repeatable)")
	StartCmd.Flags().StringSliceVar(&startNotify, "activity-notification-policy", nil, "Comma-separated run statuses that trigger notifications (INACTIVE, SUCCEEDED, FAILED)")
	StartCmd.Flags().BoolVar(&startWait, "wait", false, "Wait for flow to complete")
}

func runFlowsStart(cmd *cobra.Command, args []string) error {
	flowID := args[0]

	// Validate input
	if startInputFile == "" && startInputJSON == "" {
		return fmt.Errorf("either --input-file or --input must be provided")
	}
	if startInputFile != "" && startInputJSON != "" {
		return fmt.Errorf("cannot specify both --input-file and --input")
	}

	// Read input
	var inputJSON []byte
	var err error

	if startInputFile != "" {
		inputJSON, err = os.ReadFile(startInputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	} else {
		inputJSON = []byte(startInputJSON)
	}

	// Parse input
	var input map[string]interface{}
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		return fmt.Errorf("failed to parse input JSON: %w", err)
	}

	// Create context with timeout
	var ctx context.Context
	var cancel context.CancelFunc
	if startWait {
		// Longer timeout for waiting
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	}
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Build run input. In v4 the flow ID is passed to RunFlow directly and the
	// first-state input goes under Body.
	runInput := &flows.FlowInput{
		Body:        input,
		Label:       startLabel,
		Tags:        startTags,
		RunManagers: startManagers,
		RunMonitors: startMonitors,
	}
	if len(startNotify) > 0 {
		runInput.ActivityNotificationPolicy = &flows.RunActivityNotificationPolicy{Status: startNotify}
	}

	// Start the flow (v4 RunFlow replaces the v3 RunFlow(request) form).
	run, err := flowsClient.RunFlow(ctx, flowID, runInput)
	if err != nil {
		return fmt.Errorf("error starting flow: %w", err)
	}

	// Display initial run information
	fmt.Fprintf(os.Stdout, "Flow started successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Run ID:    %s\n", run.RunID)
	fmt.Fprintf(os.Stdout, "Flow ID:   %s\n", run.FlowID)
	fmt.Fprintf(os.Stdout, "Status:    %s\n", run.Status)
	fmt.Fprintf(os.Stdout, "Started:   %s\n", run.StartTime.Format(time.RFC3339))

	// Wait for completion if requested
	if startWait {
		fmt.Fprintf(os.Stdout, "\nWaiting for flow to complete...\n")

		finalRun, err := flowsClient.WaitForRun(ctx, run.RunID, 5*time.Second)
		if err != nil {
			return fmt.Errorf("error waiting for flow completion: %w", err)
		}

		fmt.Fprintf(os.Stdout, "\nFlow completed!\n")
		fmt.Fprintf(os.Stdout, "Final Status:  %s\n", finalRun.Status)
		if !finalRun.EndTime.IsZero() {
			fmt.Fprintf(os.Stdout, "Completed At:  %s\n", finalRun.EndTime.Format(time.RFC3339))
		}

		// Display details if available
		if finalRun.Details != nil {
			fmt.Fprintf(os.Stdout, "\nDetails:\n")
			detailsJSON, _ := json.MarshalIndent(finalRun.Details, "  ", "  ")
			fmt.Fprintf(os.Stdout, "%s\n", string(detailsJSON))
		}
	} else {
		fmt.Fprintf(os.Stdout, "\nMonitor run status with: globus flows run show %s\n", run.RunID)
	}

	return nil
}
