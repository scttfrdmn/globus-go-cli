// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateName        string
	updateDescription string
	updatePublic      *bool
)

// FunctionUpdateCmd represents the compute function update command
var FunctionUpdateCmd = &cobra.Command{
	Use:   "update FUNCTION_ID",
	Short: "Update a registered function's metadata",
	Long: `Update a function's name, description, or visibility.

NOTE: The Globus Compute API does not expose a function-update endpoint, so
this command is not supported. Register a new function instead with
"globus compute function register".`,
	Args: cobra.ExactArgs(1),
	RunE: runFunctionUpdate,
}

func init() {
	FunctionUpdateCmd.Flags().StringVar(&updateName, "name", "", "New function name")
	FunctionUpdateCmd.Flags().StringVar(&updateDescription, "description", "", "New function description")

	// Use a pointer so we can detect if flag was set
	var publicFlag bool
	FunctionUpdateCmd.Flags().BoolVar(&publicFlag, "public", false, "Make function publicly visible")
	updatePublic = &publicFlag
}

func runFunctionUpdate(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("updating functions is not supported by the Globus Compute API; register a new function instead")
}
