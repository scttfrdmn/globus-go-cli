// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

Note: You cannot update the function code itself. To change code,
register a new function version.

Examples:
  # Update function name
  globus compute function update FUNCTION_ID --name "new_name"

  # Update description
  globus compute function update FUNCTION_ID --description "Updated description"

  # Make function public
  globus compute function update FUNCTION_ID --public=true`,
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
	functionID := args[0]

	// Build update request with only specified fields
	request := &compute.FunctionUpdateRequest{}

	if updateName != "" {
		request.Name = updateName
	}

	if updateDescription != "" {
		request.Description = updateDescription
	}

	if cmd.Flags().Changed("public") {
		request.Public = updatePublic
	}

	// Validate that at least one field is being updated
	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("description") && !cmd.Flags().Changed("public") {
		return fmt.Errorf("at least one of --name, --description, or --public must be specified")
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

	// Update function
	function, err := computeClient.UpdateFunction(ctx, functionID, request)
	if err != nil {
		return fmt.Errorf("error updating function: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Function updated successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Function ID:   %s\n", function.ID)
	if function.Name != "" {
		fmt.Fprintf(os.Stdout, "Name:          %s\n", function.Name)
	}
	fmt.Fprintf(os.Stdout, "Modified:      %s\n", function.ModifiedAt.Format(time.RFC3339))

	return nil
}
