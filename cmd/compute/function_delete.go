// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var functionDeleteYes bool

// FunctionDeleteCmd represents the compute function delete command
var FunctionDeleteCmd = &cobra.Command{
	Use:   "delete FUNCTION_ID",
	Short: "Delete a registered function",
	Long: `Delete a function from Globus Compute.

WARNING: This action cannot be undone. The function will be permanently deleted.

Examples:
  # Delete a function (with confirmation)
  globus compute function delete FUNCTION_ID

  # Delete without confirmation prompt
  globus compute function delete FUNCTION_ID --yes`,
	Args: cobra.ExactArgs(1),
	RunE: runFunctionDelete,
}

func init() {
	FunctionDeleteCmd.Flags().BoolVarP(&functionDeleteYes, "yes", "y", false, "Skip confirmation prompt")
}

func runFunctionDelete(cmd *cobra.Command, args []string) error {
	functionID := args[0]

	// Confirm deletion unless --yes flag is set
	if !functionDeleteYes {
		fmt.Fprintf(os.Stderr, "Are you sure you want to delete function %s? This cannot be undone.\n", functionID)
		fmt.Fprintf(os.Stderr, "Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Compute client authorized for the current profile.
	computeClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Delete function (returns the passthrough result document, ignored here)
	if _, err = computeClient.DeleteFunction(ctx, functionID); err != nil {
		return fmt.Errorf("error deleting function: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Function %s deleted successfully.\n", functionID)

	return nil
}
