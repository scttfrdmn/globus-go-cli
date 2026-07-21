// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// RunShowDefinitionCmd represents the flows run show-definition command
var RunShowDefinitionCmd = &cobra.Command{
	Use:   "show-definition RUN_ID",
	Short: "Show flow definition and input schema for a run",
	Long: `Display the flow definition and input schema that were used for a specific run.

This shows the exact flow definition and input schema that were in effect
when the run was started, which is useful for understanding the run's behavior.

Examples:
  # Show flow definition for a run
  globus flows run show-definition RUN_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunShowDefinition,
}

func runFlowsRunShowDefinition(cmd *cobra.Command, args []string) error {
	runID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get run to find the flow ID
	run, err := flowsClient.GetRun(ctx, runID, nil)
	if err != nil {
		return fmt.Errorf("error getting run: %w", err)
	}

	// Get the flow definition
	flow, err := flowsClient.GetFlow(ctx, run.FlowID)
	if err != nil {
		return fmt.Errorf("error getting flow: %w", err)
	}

	// Display flow definition and schema
	fmt.Printf("Flow Definition and Input Schema\n")
	fmt.Printf("================================\n\n")

	fmt.Printf("Run ID:        %s\n", run.RunID)
	fmt.Printf("Flow ID:       %s\n", flow.ID)
	fmt.Printf("Flow Title:    %s\n", flow.Title)

	// Display definition
	if flow.Definition != nil {
		fmt.Printf("\nFlow Definition:\n")
		definitionJSON, _ := json.MarshalIndent(flow.Definition, "  ", "  ")
		fmt.Printf("%s\n", string(definitionJSON))
	}

	// Display input schema
	if flow.InputSchema != nil {
		fmt.Printf("\nInput Schema:\n")
		schemaJSON, _ := json.MarshalIndent(flow.InputSchema, "  ", "  ")
		fmt.Printf("%s\n", string(schemaJSON))
	}

	return nil
}
