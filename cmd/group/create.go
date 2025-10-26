// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package group

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	createName        string
	createDescription string
	createPublic      bool
)

// CreateCmd represents the group create command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Globus group",
	Long: `Create a new Globus group with specified settings.

Groups provide collaborative access control and membership management
for Globus resources.

Examples:
  # Create a basic group
  globus group create --name "My Research Group" --description "Group for our research project"

  # Create a public group
  globus group create --name "Public Data Group" --description "Open access group" --public

Output:
  Displays the newly created group ID and details.`,
	RunE: runCreateGroup,
}

func init() {
	CreateCmd.Flags().StringVar(&createName, "name", "", "Name for the new group (required)")
	CreateCmd.Flags().StringVar(&createDescription, "description", "", "Description of the group")
	CreateCmd.Flags().BoolVar(&createPublic, "public", false, "Make the group publicly visible")
	CreateCmd.MarkFlagRequired("name")
}

func runCreateGroup(cmd *cobra.Command, args []string) error {
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

	// Create groups client
	groupsClient, err := groups.NewClient(
		groups.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create groups client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Prepare group creation request
	newGroup := &groups.GroupCreate{
		Name:        createName,
		Description: createDescription,
		PublicGroup: createPublic,
	}

	// Create the group
	createdGroup, err := groupsClient.CreateGroup(ctx, newGroup)
	if err != nil {
		return fmt.Errorf("error creating group: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Group created successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Group ID:    %s\n", createdGroup.ID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", createdGroup.Name)
	fmt.Fprintf(os.Stdout, "Description: %s\n", createdGroup.Description)
	fmt.Fprintf(os.Stdout, "Identity ID: %s\n", createdGroup.IdentityID)
	fmt.Fprintf(os.Stdout, "\nYou can now add members and configure the group.\n")

	return nil
}
