// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"fmt"

	"github.com/spf13/cobra"
)

// FunctionListCmd represents the compute function list command
var FunctionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered Globus Compute functions",
	Long: `List functions you have registered with Globus Compute.

NOTE: The Globus Compute API does not expose a function-listing endpoint, so
this command is not supported. Retrieve a specific function by ID with
"globus compute function show FUNCTION_ID".`,
	RunE: runFunctionList,
}

func runFunctionList(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("listing functions is not supported by the Globus Compute API; use \"globus compute function show FUNCTION_ID\"")
}
