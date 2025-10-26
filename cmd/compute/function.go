// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"github.com/spf13/cobra"
)

// GetFunctionCmd returns the function subcommand for compute
func GetFunctionCmd() *cobra.Command {
	functionCmd := &cobra.Command{
		Use:   "function",
		Short: "Manage Globus Compute functions",
		Long: `Commands for managing Globus Compute functions.

Functions are Python code that can be executed remotely on compute endpoints.
Register functions once, then execute them many times on different endpoints.`,
	}

	// Add function subcommands
	functionCmd.AddCommand(FunctionListCmd)
	functionCmd.AddCommand(FunctionShowCmd)
	functionCmd.AddCommand(FunctionRegisterCmd)
	functionCmd.AddCommand(FunctionUpdateCmd)
	functionCmd.AddCommand(FunctionDeleteCmd)

	return functionCmd
}
