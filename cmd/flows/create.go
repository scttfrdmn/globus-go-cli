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
	createTitle          string
	createDescription    string
	createSubtitle       string
	createDefinitionFile string
	createSchemaFile     string
	createKeywords       []string
	createPublic         bool

	createAdministrators []string
	createStarters       []string
	createViewers        []string
	createRunManagers    []string
	createRunMonitors    []string
	createSubscriptionID string
	createAuthPolicyID   string

	// Authentication policy flags (Python SDK v4.1.0)
	createHighAssurance   bool
	createRequiredMFA     bool
	createSessionPolicies []string
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
	CreateCmd.Flags().StringVar(&createSubtitle, "subtitle", "", "A concise summary of the flow's purpose")
	CreateCmd.Flags().StringVar(&createDefinitionFile, "definition-file", "", "Path to flow definition JSON file (required)")
	CreateCmd.Flags().StringVar(&createSchemaFile, "schema-file", "", "Path to input schema JSON file")
	CreateCmd.Flags().StringSliceVar(&createKeywords, "keywords", []string{}, "Comma-separated keywords")
	CreateCmd.Flags().BoolVar(&createPublic, "public", false, "Make flow publicly visible")

	// Principal role lists (repeatable). Backed by FlowCreate.
	CreateCmd.Flags().StringArrayVar(&createAdministrators, "administrator", nil, "A principal that may administer the flow (repeatable)")
	CreateCmd.Flags().StringArrayVar(&createStarters, "starter", nil, "A principal that may start a run of the flow (repeatable); use \"all_authenticated_users\" for any user")
	CreateCmd.Flags().StringArrayVar(&createViewers, "viewer", nil, "A principal that may view the flow (repeatable); use \"public\" to make it visible to everyone")
	CreateCmd.Flags().StringArrayVar(&createRunManagers, "run-manager", nil, "A principal that may manage the flow's runs (repeatable)")
	CreateCmd.Flags().StringArrayVar(&createRunMonitors, "run-monitor", nil, "A principal that may monitor the flow's runs (repeatable)")
	CreateCmd.Flags().StringVar(&createSubscriptionID, "subscription-id", "", "Set a subscription_id for the flow, marking it as subscription tier")
	CreateCmd.Flags().StringVar(&createAuthPolicyID, "authentication-policy-id", "", "A Globus Auth authentication policy ID to enforce on the flow (must require high-assurance)")

	// Authentication policy flags (Python SDK v4.1.0)
	CreateCmd.Flags().BoolVar(&createHighAssurance, "high-assurance", false, "Require high-assurance authentication for flow runs")
	CreateCmd.Flags().BoolVar(&createRequiredMFA, "required-mfa", false, "Require multi-factor authentication for flow runs")
	CreateCmd.Flags().StringSliceVar(&createSessionPolicies, "session-policies", []string{}, "Named authentication session policies required for flow runs")

	_ = CreateCmd.MarkFlagRequired("title")
	_ = CreateCmd.MarkFlagRequired("definition-file")
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Flows client authorized for the current profile.
	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Build create request
	request := &flows.FlowCreate{
		Title:                  createTitle,
		Description:            createDescription,
		Subtitle:               createSubtitle,
		Definition:             definition,
		InputSchema:            inputSchema,
		Keywords:               createKeywords,
		FlowAdministrators:     createAdministrators,
		FlowStarters:           createStarters,
		FlowViewers:            createViewers,
		RunManagers:            createRunManagers,
		RunMonitors:            createRunMonitors,
		SubscriptionID:         createSubscriptionID,
		AuthenticationPolicyID: createAuthPolicyID,
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
	fmt.Fprintf(os.Stdout, "Owner:     %s\n", flow.OwnerID)
	fmt.Fprintf(os.Stdout, "Created:   %s\n", flow.Created.Format(time.RFC3339))

	return nil
}
