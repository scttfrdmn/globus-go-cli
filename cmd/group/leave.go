// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package group

import (
	"fmt"

	"github.com/spf13/cobra"
)

// LeaveCmd represents the group leave command
var LeaveCmd = &cobra.Command{
	Use:   "leave GROUP_ID",
	Short: "Leave a Globus group",
	Long: `Leave a Globus group where you are currently a member.

Note: If you are the last administrator, you may not be able to leave
the group without first assigning another administrator.

Examples:
  # Leave a group
  globus group leave GROUP_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runLeaveGroup,
}

func runLeaveGroup(cmd *cobra.Command, args []string) error {
	// Note: Leave functionality requires direct API integration
	// The SDK v3.65.0-1 doesn't yet expose a high-level LeaveGroup method
	// This can be done by removing yourself as a member using RemoveMember

	return fmt.Errorf("group leave functionality is not yet fully implemented in SDK v3.65.0-1\n" +
		"To leave a group, have an admin remove you, or use the Globus web interface: https://app.globus.org/groups")
}
