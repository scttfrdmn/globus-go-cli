// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package group

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/groups"
	"github.com/spf13/cobra"
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Add member to group via a single batch membership action
	// (POST /groups/{id}); the API has no dedicated add-member route.
	_, err = groupsClient.BatchMembershipAction(ctx, groupID, &groups.BatchMembershipActions{
		Add: []groups.MemberWithRole{{IdentityID: identityID, Role: memberAddRole}},
	})
	if err != nil {
		return fmt.Errorf("error adding member: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Successfully added member %s to group %s with role '%s'.\n", identityID, groupID, memberAddRole)

	return nil
}
