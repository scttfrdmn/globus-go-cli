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

// ShowCmd represents the flows show command
var ShowCmd = &cobra.Command{
	Use:   "show FLOW_ID",
	Short: "Show details of a Globus Flow",
	Long: `Display detailed information about a specific flow.

This includes the flow's metadata, definition, input schema, and run statistics.

Examples:
  # Show flow details
  globus flows show FLOW_ID

  # Show flow with JSON output
  globus flows show FLOW_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsShow,
}

func runFlowsShow(cmd *cobra.Command, args []string) error {
	flowID := args[0]

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

	// Get flow
	flow, err := flowsClient.GetFlow(ctx, flowID)
	if err != nil {
		return fmt.Errorf("error getting flow: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Flow Details\n")
		fmt.Printf("============\n\n")

		fmt.Printf("Flow ID:       %s\n", flow.ID)
		fmt.Printf("Title:         %s\n", flow.Title)
		if flow.Description != "" {
			fmt.Printf("Description:   %s\n", flow.Description)
		}
		fmt.Printf("Owner:         %s\n", flow.FlowOwner)
		fmt.Printf("Public:        %t\n", flow.Public)
		fmt.Printf("Managed:       %t\n", flow.Managed)
		fmt.Printf("Run Count:     %d\n", flow.RunCount)
		fmt.Printf("Created:       %s\n", flow.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated:       %s\n", flow.UpdatedAt.Format(time.RFC3339))

		if len(flow.Keywords) > 0 {
			fmt.Printf("Keywords:      %v\n", flow.Keywords)
		}

		// Display definition
		if flow.Definition != nil {
			fmt.Printf("\nDefinition:\n")
			definitionJSON, _ := json.MarshalIndent(flow.Definition, "  ", "  ")
			fmt.Printf("  %s\n", string(definitionJSON))
		}

		// Display input schema
		if flow.InputSchema != nil {
			fmt.Printf("\nInput Schema:\n")
			schemaJSON, _ := json.MarshalIndent(flow.InputSchema, "  ", "  ")
			fmt.Printf("  %s\n", string(schemaJSON))
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "Title", "Description", "FlowOwner", "Public", "Managed", "RunCount", "CreatedAt", "UpdatedAt"}
		if err := formatter.FormatOutput(flow, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
