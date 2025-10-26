// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package group

import (
	"github.com/spf13/cobra"
)

// GetMemberCmd returns the member command
func GetMemberCmd() *cobra.Command {
	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Manage group members",
		Long: `Commands for managing members of a Globus group.

Group members can have different roles including:
- member: Basic group membership
- manager: Can manage group membership
- admin: Full administrative access

Examples:
  # List group members
  globus group member list GROUP_ID

  # Add a member
  globus group member add GROUP_ID IDENTITY_ID

  # Remove a member
  globus group member remove GROUP_ID IDENTITY_ID

  # Invite a new member
  globus group member invite GROUP_ID EMAIL`,
	}

	// Add member subcommands
	memberCmd.AddCommand(MemberListCmd)
	memberCmd.AddCommand(MemberAddCmd)
	memberCmd.AddCommand(MemberRemoveCmd)
	memberCmd.AddCommand(MemberInviteCmd)

	return memberCmd
}
