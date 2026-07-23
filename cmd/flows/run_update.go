// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
	"github.com/spf13/cobra"
)

var (
	runUpdateLabel    string
	runUpdateTags     []string
	runUpdateManagers []string
	runUpdateMonitors []string
)

// RunUpdateCmd represents the flows run update command
var RunUpdateCmd = &cobra.Command{
	Use:   "update RUN_ID",
	Short: "Update a flow run's metadata",
	Long: `Update a run's label and tags.

You can modify the label and tags of a run for better organization
and searchability.

Examples:
  # Update run label
  globus flows run update RUN_ID --label "Production run v2"

  # Update run tags
  globus flows run update RUN_ID --tags "prod,critical,automated"

  # Update both label and tags
  globus flows run update RUN_ID \\
    --label "Updated label" \\
    --tags "new,tags"`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsRunUpdate,
}

func init() {
	RunUpdateCmd.Flags().StringVar(&runUpdateLabel, "label", "", "New label for the run")
	RunUpdateCmd.Flags().StringSliceVar(&runUpdateTags, "tags", []string{}, "New comma-separated tags (empty string clears)")
	RunUpdateCmd.Flags().StringSliceVar(&runUpdateManagers, "managers", nil, "Comma-separated principals that may manage the run (empty string clears)")
	RunUpdateCmd.Flags().StringSliceVar(&runUpdateMonitors, "monitors", nil, "Comma-separated principals that may monitor the run (empty string clears)")
}

func runFlowsRunUpdate(cmd *cobra.Command, args []string) error {
	runID := args[0]

	// Build update request with only specified fields
	request := &flows.RunUpdate{}

	if cmd.Flags().Changed("label") {
		request.Label = runUpdateLabel
	}

	if cmd.Flags().Changed("tags") {
		request.Tags = runUpdateTags
	}

	if cmd.Flags().Changed("managers") {
		request.RunManagers = runUpdateManagers
	}

	if cmd.Flags().Changed("monitors") {
		request.RunMonitors = runUpdateMonitors
	}

	// Validate that at least one field is being updated
	if !cmd.Flags().Changed("label") && !cmd.Flags().Changed("tags") &&
		!cmd.Flags().Changed("managers") && !cmd.Flags().Changed("monitors") {
		return fmt.Errorf("at least one of --label, --tags, --managers, or --monitors must be specified")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Update run
	run, err := flowsClient.UpdateRun(ctx, runID, request)
	if err != nil {
		return fmt.Errorf("error updating run: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Run updated successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Run ID:    %s\n", run.RunID)
	if run.Label != "" {
		fmt.Fprintf(os.Stdout, "Label:     %s\n", run.Label)
	}

	return nil
}
