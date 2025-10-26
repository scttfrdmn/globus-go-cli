// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package group

import (
	"fmt"

	"github.com/spf13/cobra"
)

var joinRequest bool

// JoinCmd represents the group join command
var JoinCmd = &cobra.Command{
	Use:   "join GROUP_ID",
	Short: "Join a Globus group",
	Long: `Join a Globus group as a member.

For open groups, you will be added immediately. For closed groups,
use the --request flag to submit a membership request that requires approval.

Examples:
  # Join an open group
  globus group join GROUP_ID

  # Request to join a closed group
  globus group join GROUP_ID --request`,
	Args: cobra.ExactArgs(1),
	RunE: runJoinGroup,
}

func init() {
	JoinCmd.Flags().BoolVar(&joinRequest, "request", false, "Request membership (for groups requiring approval)")
}

func runJoinGroup(cmd *cobra.Command, args []string) error {
	// Note: Join functionality requires direct API integration
	// The SDK v3.65.0-1 doesn't yet expose a high-level JoinGroup method
	// This would typically be done via the Groups API /groups/{group_id}/join endpoint

	return fmt.Errorf("group join functionality is not yet available in SDK v3.65.0-1\n" +
		"Please use the Globus web interface to join groups: https://app.globus.org/groups")
}
