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

var memberAddRole string

// MemberAddCmd represents the member add command
var MemberAddCmd = &cobra.Command{
	Use:   "add GROUP_ID IDENTITY_ID",
	Short: "Add a member to a group",
	Long: `Add a member to a Globus group with a specified role.

Available roles:
  - member: Basic group membership (default)
  - manager: Can manage group membership
  - admin: Full administrative access

Examples:
  # Add a basic member
  globus group member add GROUP_ID IDENTITY_ID

  # Add a manager
  globus group member add GROUP_ID IDENTITY_ID --role manager

  # Add an admin
  globus group member add GROUP_ID IDENTITY_ID --role admin`,
	Args: cobra.ExactArgs(2),
	RunE: runMemberAdd,
}

func init() {
	MemberAddCmd.Flags().StringVar(&memberAddRole, "role", "member", "Role for the new member (member, manager, admin)")
}

func runMemberAdd(cmd *cobra.Command, args []string) error {
	groupID := args[0]
	identityID := args[1]

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

	// Add member to group
	err = groupsClient.AddMember(ctx, groupID, identityID, memberAddRole)
	if err != nil {
		return fmt.Errorf("error adding member: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Successfully added member %s to group %s with role '%s'.\n", identityID, groupID, memberAddRole)

	return nil
}
