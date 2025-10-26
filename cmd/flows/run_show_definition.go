// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// Get run to find the flow ID
	run, err := flowsClient.GetRun(ctx, runID)
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
