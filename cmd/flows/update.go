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
	updateTitle         string
	updateDescription   string
	updateDefinitionFile string
	updateSchemaFile    string
	updateKeywords      []string
	updatePublic        *bool
)

// UpdateCmd represents the flows update command
var UpdateCmd = &cobra.Command{
	Use:   "update FLOW_ID",
	Short: "Update a Globus Flow",
	Long: `Update an existing flow's metadata, definition, or input schema.

You can update any combination of title, description, definition, schema,
keywords, and visibility. Only the fields you specify will be updated.

Examples:
  # Update flow title and description
  globus flows update FLOW_ID --title "New Title" --description "New description"

  # Update flow definition
  globus flows update FLOW_ID --definition-file new_flow.json

  # Make flow public
  globus flows update FLOW_ID --public=true

  # Update multiple fields
  globus flows update FLOW_ID \\
    --title "Updated Flow" \\
    --definition-file flow_v2.json \\
    --keywords "transfer,v2"`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsUpdate,
}

func init() {
	UpdateCmd.Flags().StringVar(&updateTitle, "title", "", "Flow title")
	UpdateCmd.Flags().StringVar(&updateDescription, "description", "", "Flow description")
	UpdateCmd.Flags().StringVar(&updateDefinitionFile, "definition-file", "", "Path to flow definition JSON file")
	UpdateCmd.Flags().StringVar(&updateSchemaFile, "schema-file", "", "Path to input schema JSON file")
	UpdateCmd.Flags().StringSliceVar(&updateKeywords, "keywords", []string{}, "Comma-separated keywords")

	// Use a pointer so we can detect if flag was set
	var publicFlag bool
	UpdateCmd.Flags().BoolVar(&publicFlag, "public", false, "Make flow publicly visible")
	updatePublic = &publicFlag
}

func runFlowsUpdate(cmd *cobra.Command, args []string) error {
	flowID := args[0]

	// Build update request with only specified fields
	request := &flows.FlowUpdateRequest{}

	if updateTitle != "" {
		request.Title = updateTitle
	}

	if updateDescription != "" {
		request.Description = updateDescription
	}

	if updateDefinitionFile != "" {
		definitionData, err := os.ReadFile(updateDefinitionFile)
		if err != nil {
			return fmt.Errorf("failed to read definition file: %w", err)
		}

		var definition map[string]interface{}
		if err := json.Unmarshal(definitionData, &definition); err != nil {
			return fmt.Errorf("failed to parse definition JSON: %w", err)
		}
		request.Definition = definition
	}

	if updateSchemaFile != "" {
		schemaData, err := os.ReadFile(updateSchemaFile)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}

		var inputSchema map[string]interface{}
		if err := json.Unmarshal(schemaData, &inputSchema); err != nil {
			return fmt.Errorf("failed to parse schema JSON: %w", err)
		}
		request.InputSchema = inputSchema
	}

	if len(updateKeywords) > 0 {
		request.Keywords = updateKeywords
	}

	if cmd.Flags().Changed("public") {
		request.Public = updatePublic
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

	// Update flow
	flow, err := flowsClient.UpdateFlow(ctx, flowID, request)
	if err != nil {
		return fmt.Errorf("error updating flow: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Flow updated successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Flow ID:   %s\n", flow.ID)
	fmt.Fprintf(os.Stdout, "Title:     %s\n", flow.Title)
	fmt.Fprintf(os.Stdout, "Updated:   %s\n", flow.UpdatedAt.Format(time.RFC3339))

	return nil
}
