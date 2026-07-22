// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var validateSchemaFile string

// ValidateCmd represents the flows validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate DEFINITION_FILE",
	Short: "Validate a flow definition",
	Long: `Validate a flow definition against the Globus Flows schema, without
creating a flow.

DEFINITION_FILE is a path to a JSON file containing the flow definition.

Examples:
  # Validate a flow definition
  globus flows validate flow_definition.json

  # Validate with an input schema
  globus flows validate flow_definition.json --input-schema schema.json`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsValidate,
}

func init() {
	ValidateCmd.Flags().StringVar(&validateSchemaFile, "input-schema", "", "Path to an input schema JSON file")
}

func runFlowsValidate(cmd *cobra.Command, args []string) error {
	definitionData, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("error reading definition file: %w", err)
	}
	var definition map[string]interface{}
	if err := json.Unmarshal(definitionData, &definition); err != nil {
		return fmt.Errorf("error parsing definition JSON: %w", err)
	}

	var inputSchema map[string]interface{}
	if validateSchemaFile != "" {
		schemaData, err := os.ReadFile(validateSchemaFile)
		if err != nil {
			return fmt.Errorf("error reading input schema file: %w", err)
		}
		if err := json.Unmarshal(schemaData, &inputSchema); err != nil {
			return fmt.Errorf("error parsing input schema JSON: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	flowsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	result, err := flowsClient.ValidateFlow(ctx, definition, inputSchema)
	if err != nil {
		return fmt.Errorf("flow definition is invalid: %w", err)
	}

	format := viper.GetString("format")
	if format != "text" {
		return output.NewFormatter(format, os.Stdout).FormatOutput(result, nil)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Flow definition is valid.")
	return nil
}
