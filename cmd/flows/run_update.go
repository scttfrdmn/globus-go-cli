// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	runUpdateLabel string
	runUpdateTags  []string
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
	RunUpdateCmd.Flags().StringSliceVar(&runUpdateTags, "tags", []string{}, "New comma-separated tags")
}

func runFlowsRunUpdate(cmd *cobra.Command, args []string) error {
	runID := args[0]

	// Build update request with only specified fields
	request := &flows.RunUpdateRequest{}

	if cmd.Flags().Changed("label") {
		request.Label = runUpdateLabel
	}

	if cmd.Flags().Changed("tags") {
		request.Tags = runUpdateTags
	}

	// Validate that at least one field is being updated
	if !cmd.Flags().Changed("label") && !cmd.Flags().Changed("tags") {
		return fmt.Errorf("at least one of --label or --tags must be specified")
	}

	// Get current profile
	profile := viper.GetString("profile")

	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create authorizer
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create flows client
	flowsClient, err := flows.NewClient(
		flows.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create flows client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
	if len(run.Tags) > 0 {
		fmt.Fprintf(os.Stdout, "Tags:      %v\n", run.Tags)
	}

	return nil
}
