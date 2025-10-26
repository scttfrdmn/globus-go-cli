// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/scttfrdmn/globus-go-cli/cmd/group"
	"github.com/spf13/cobra"
)

// getGroupCommand returns the root group command
func getGroupCommand() *cobra.Command {
	groupCmd := &cobra.Command{
		Use:   "group",
		Short: "Commands for Globus Groups",
		Long: `Commands for managing Globus Groups.

Globus Groups provides collaborative group management including:
- Creating and managing groups
- Managing group memberships
- Handling join requests and invitations
- Setting group policies and permissions

Examples:
  # List your groups
  globus group list

  # Create a new group
  globus group create --name "My Research Group"

  # Show group details
  globus group show GROUP_ID

  # Add a member to a group
  globus group member add GROUP_ID IDENTITY_ID`,
	}

	// Add subcommands
	groupCmd.AddCommand(group.ListCmd)
	groupCmd.AddCommand(group.CreateCmd)
	groupCmd.AddCommand(group.ShowCmd)
	groupCmd.AddCommand(group.UpdateCmd)
	groupCmd.AddCommand(group.DeleteCmd)
	groupCmd.AddCommand(group.JoinCmd)
	groupCmd.AddCommand(group.LeaveCmd)
	groupCmd.AddCommand(group.GetMemberCmd())

	return groupCmd
}
