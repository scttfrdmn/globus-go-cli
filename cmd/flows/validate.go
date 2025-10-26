// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

// NOTE: This command is a placeholder because the Go SDK v3.65.0-1 does not
// yet support flow definition validation.

package flows

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ValidateCmd represents the flows validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate DEFINITION_FILE",
	Short: "Validate a flow definition (not yet supported)",
	Long: `Validate a flow definition against the Flows schema.

NOTE: This command is not yet fully implemented because the Go SDK v3.65.0-1
does not support flow definition validation.

You can validate your flow definition by:
1. Using the Globus web interface (https://app.globus.org)
2. Using the Python Globus CLI: globus flows validate
3. Attempting to create the flow, which will validate on the server

Examples (when supported):
  # Validate a flow definition
  globus flows validate flow_definition.json`,
	Args: cobra.ExactArgs(1),
	RunE: runFlowsValidate,
}

func runFlowsValidate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("flow definition validation is not yet available in SDK v3.65.0-1\n" +
		"You can validate your flow definition by:\n" +
		"  1. Using the Globus web interface (https://app.globus.org)\n" +
		"  2. Using the Python Globus CLI: globus flows validate\n" +
		"  3. Attempting to create the flow (server-side validation)\n\n" +
		"The Go SDK will add validation support in a future release.")
}
