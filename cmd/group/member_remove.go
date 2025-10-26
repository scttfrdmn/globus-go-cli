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

// MemberRemoveCmd represents the member remove command
var MemberRemoveCmd = &cobra.Command{
	Use:   "remove GROUP_ID IDENTITY_ID",
	Short: "Remove a member from a group",
	Long: `Remove a member from a Globus group.

You must be an administrator or manager of the group to remove members.

Examples:
  # Remove a member
  globus group member remove GROUP_ID IDENTITY_ID`,
	Args: cobra.ExactArgs(2),
	RunE: runMemberRemove,
}

func runMemberRemove(cmd *cobra.Command, args []string) error {
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

	// Remove member from group
	err = groupsClient.RemoveMember(ctx, groupID, identityID)
	if err != nil {
		return fmt.Errorf("error removing member: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Successfully removed member %s from group %s.\n", identityID, groupID)

	return nil
}
