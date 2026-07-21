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

var inviteRole string

// MemberInviteCmd represents the member invite command
var MemberInviteCmd = &cobra.Command{
	Use:   "invite GROUP_ID IDENTITY_ID",
	Short: "Invite a member to join a group",
	Long: `Invite a user to join a Globus group.

The invitee is identified by their Globus identity ID (not an email
address); the Groups API requires an identity ID for invitations.

Available roles:
  - member: Basic group membership (default)
  - manager: Can manage group membership
  - admin: Full administrative access

Examples:
  # Invite a basic member
  globus group member invite GROUP_ID IDENTITY_ID

  # Invite with a specific role
  globus group member invite GROUP_ID IDENTITY_ID --role manager`,
	Args: cobra.ExactArgs(2),
	RunE: runMemberInvite,
}

func init() {
	MemberInviteCmd.Flags().StringVar(&inviteRole, "role", "member", "Role for the invited member (member, manager, admin)")
}

func runMemberInvite(cmd *cobra.Command, args []string) error {
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

	// Invite member via a single batch membership action
	// (POST /groups/{id}); the API has no dedicated invite route.
	_, err = groupsClient.BatchMembershipAction(ctx, groupID, &groups.BatchMembershipActions{
		Invite: []groups.MemberWithRole{{IdentityID: identityID, Role: inviteRole}},
	})
	if err != nil {
		return fmt.Errorf("error inviting member: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Successfully invited identity %s to group %s with role '%s'.\n", identityID, groupID, inviteRole)

	return nil
}
