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
	createTitle         string
	createDescription   string
	createDefinitionFile string
	createSchemaFile    string
	createKeywords      []string
	createPublic        bool
)

// CreateCmd represents the flows create command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Globus Flow",
	Long: `Create a new flow with the specified definition and metadata.

A flow definition is a JSON document that describes the workflow steps,
actions, and logic. The input schema defines the required and optional
parameters for running the flow.

Examples:
  # Create a flow from a definition file
  globus flows create --title "My Flow" --definition-file flow.json

  # Create with description and keywords
  globus flows create --title "Transfer Flow" \\
    --description "Automated data transfer" \\
    --definition-file flow.json \\
    --keywords "transfer,automation"

  # Create a public flow
  globus flows create --title "Public Flow" \\
    --definition-file flow.json \\
    --public`,
	RunE: runFlowsCreate,
}

func init() {
	CreateCmd.Flags().StringVar(&createTitle, "title", "", "Flow title (required)")
	CreateCmd.Flags().StringVar(&createDescription, "description", "", "Flow description")
	CreateCmd.Flags().StringVar(&createDefinitionFile, "definition-file", "", "Path to flow definition JSON file (required)")
	CreateCmd.Flags().StringVar(&createSchemaFile, "schema-file", "", "Path to input schema JSON file")
	CreateCmd.Flags().StringSliceVar(&createKeywords, "keywords", []string{}, "Comma-separated keywords")
	CreateCmd.Flags().BoolVar(&createPublic, "public", false, "Make flow publicly visible")

	CreateCmd.MarkFlagRequired("title")
	CreateCmd.MarkFlagRequired("definition-file")
}

func runFlowsCreate(cmd *cobra.Command, args []string) error {
	// Read definition file
	definitionData, err := os.ReadFile(createDefinitionFile)
	if err != nil {
		return fmt.Errorf("failed to read definition file: %w", err)
	}

	var definition map[string]interface{}
	if err := json.Unmarshal(definitionData, &definition); err != nil {
		return fmt.Errorf("failed to parse definition JSON: %w", err)
	}

	// Read schema file if provided
	var inputSchema map[string]interface{}
	if createSchemaFile != "" {
		schemaData, err := os.ReadFile(createSchemaFile)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}

		if err := json.Unmarshal(schemaData, &inputSchema); err != nil {
			return fmt.Errorf("failed to parse schema JSON: %w", err)
		}
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

	// Build create request
	request := &flows.FlowCreateRequest{
		Title:       createTitle,
		Description: createDescription,
		Definition:  definition,
		InputSchema: inputSchema,
		Keywords:    createKeywords,
		Public:      createPublic,
	}

	// Create flow
	flow, err := flowsClient.CreateFlow(ctx, request)
	if err != nil {
		return fmt.Errorf("error creating flow: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Flow created successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Flow ID:   %s\n", flow.ID)
	fmt.Fprintf(os.Stdout, "Title:     %s\n", flow.Title)
	fmt.Fprintf(os.Stdout, "Owner:     %s\n", flow.FlowOwner)
	fmt.Fprintf(os.Stdout, "Public:    %t\n", flow.Public)
	fmt.Fprintf(os.Stdout, "Created:   %s\n", flow.CreatedAt.Format(time.RFC3339))

	return nil
}
