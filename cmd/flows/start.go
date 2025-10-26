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
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	startInputFile string
	startInputJSON string
	startLabel     string
	startTags      []string
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
	var ctx context.Context
	var cancel context.CancelFunc
	if startWait {
		// Longer timeout for waiting
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	}
	defer cancel()

	// Build run request
	request := &flows.RunRequest{
		FlowID: flowID,
		Input:  input,
		Label:  startLabel,
		Tags:   startTags,
	}

	// Start the flow
	run, err := flowsClient.RunFlow(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting flow: %w", err)
	}

	// Display initial run information
	fmt.Fprintf(os.Stdout, "Flow started successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Run ID:    %s\n", run.RunID)
	fmt.Fprintf(os.Stdout, "Flow ID:   %s\n", run.FlowID)
	fmt.Fprintf(os.Stdout, "Status:    %s\n", run.Status)
	fmt.Fprintf(os.Stdout, "Created:   %s\n", run.CreatedAt.Format(time.RFC3339))

	// Wait for completion if requested
	if startWait {
		fmt.Fprintf(os.Stdout, "\nWaiting for flow to complete...\n")

		finalRun, err := flowsClient.WaitForRun(ctx, run.RunID, 5*time.Second)
		if err != nil {
			return fmt.Errorf("error waiting for flow completion: %w", err)
		}

		fmt.Fprintf(os.Stdout, "\nFlow completed!\n")
		fmt.Fprintf(os.Stdout, "Final Status:  %s\n", finalRun.Status)
		if !finalRun.CompletedAt.IsZero() {
			fmt.Fprintf(os.Stdout, "Completed At:  %s\n", finalRun.CompletedAt.Format(time.RFC3339))
		}

		// Display output if available
		if finalRun.Output != nil {
			fmt.Fprintf(os.Stdout, "\nOutput:\n")
			outputJSON, _ := json.MarshalIndent(finalRun.Output, "  ", "  ")
			fmt.Fprintf(os.Stdout, "%s\n", string(outputJSON))
		}
	} else {
		fmt.Fprintf(os.Stdout, "\nMonitor run status with: globus flows run show %s\n", run.RunID)
	}

	return nil
}
