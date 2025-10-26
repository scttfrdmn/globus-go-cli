// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// Create compute client
	computeClient, err := compute.NewClient(
		compute.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Delete function
	err = computeClient.DeleteFunction(ctx, functionID)
	if err != nil {
		return fmt.Errorf("error deleting function: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Function %s deleted successfully.\n", functionID)

	return nil
}
