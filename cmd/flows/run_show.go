// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
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

	// Get run
	run, err := flowsClient.GetRun(ctx, runID)
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
		if len(run.Tags) > 0 {
			fmt.Printf("Tags:          %v\n", run.Tags)
		}
		fmt.Printf("Owner:         %s\n", run.RunOwner)
		fmt.Printf("Created:       %s\n", run.CreatedAt.Format(time.RFC3339))
		if !run.StartedAt.IsZero() {
			fmt.Printf("Started:       %s\n", run.StartedAt.Format(time.RFC3339))
		}
		if !run.CompletedAt.IsZero() {
			fmt.Printf("Completed:     %s\n", run.CompletedAt.Format(time.RFC3339))
		}

		// Display input
		if run.Input != nil {
			fmt.Printf("\nInput:\n")
			inputJSON, _ := json.MarshalIndent(run.Input, "  ", "  ")
			fmt.Printf("%s\n", string(inputJSON))
		}

		// Display output
		if run.Output != nil {
			fmt.Printf("\nOutput:\n")
			outputJSON, _ := json.MarshalIndent(run.Output, "  ", "  ")
			fmt.Printf("%s\n", string(outputJSON))
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"RunID", "FlowID", "Status", "Label", "RunOwner", "CreatedAt", "StartedAt", "CompletedAt"}
		if err := formatter.FormatOutput(run, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
