// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package group

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	inviteRole              string
	inviteProvisionIdentity bool
)

// MemberInviteCmd represents the member invite command
var MemberInviteCmd = &cobra.Command{
	Use:   "invite GROUP_ID EMAIL",
	Short: "Invite a member to join a group",
	Long: `Invite a user to join a Globus group by email address.

The invitee will receive an email invitation to join the group.
Use --provision-identity to create a Globus identity for users who
don't have one yet.

Available roles:
  - member: Basic group membership (default)
  - manager: Can manage group membership
  - admin: Full administrative access

Examples:
  # Invite a basic member
  globus group member invite GROUP_ID user@example.com

  # Invite with a specific role
  globus group member invite GROUP_ID user@example.com --role manager

  # Invite and provision identity if needed
  globus group member invite GROUP_ID user@example.com --provision-identity`,
	Args: cobra.ExactArgs(2),
	RunE: runMemberInvite,
}

func init() {
	MemberInviteCmd.Flags().StringVar(&inviteRole, "role", "member", "Role for the invited member (member, manager, admin)")
	MemberInviteCmd.Flags().BoolVar(&inviteProvisionIdentity, "provision-identity", false, "Provision a Globus identity if the user doesn't have one")
}

func runMemberInvite(cmd *cobra.Command, args []string) error {
	groupID := args[0]

	// Note: Invite functionality requires direct API integration
	// The SDK v3.65.0-1 doesn't yet expose a high-level InviteMember method
	// This would typically be done via the Groups API /groups/{group_id}/invite endpoint

	return fmt.Errorf("member invite functionality is not yet available in SDK v3.65.0-1\n" +
		"Please use the Globus web interface to invite members: https://app.globus.org/groups/%s", groupID)
}
