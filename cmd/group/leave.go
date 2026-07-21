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

var leaveIdentity string

// LeaveCmd represents the group leave command
var LeaveCmd = &cobra.Command{
	Use:   "leave GROUP_ID",
	Short: "Leave a Globus group",
	Long: `Leave a Globus group where you are currently a member.

The --identity flag specifies the identity ID that leaves the group.

Note: If you are the last administrator, you may not be able to leave
the group without first assigning another administrator.

Examples:
  # Leave a group as a specific identity
  globus group leave GROUP_ID --identity IDENTITY_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runLeaveGroup,
}

func init() {
	LeaveCmd.Flags().StringVar(&leaveIdentity, "identity", "", "Identity ID to leave the group as (required)")
	_ = LeaveCmd.MarkFlagRequired("identity")
}

func runLeaveGroup(cmd *cobra.Command, args []string) error {
	groupID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Leave the group via a single batch membership action
	// (POST /groups/{id}); the API has no dedicated leave route.
	_, err = groupsClient.BatchMembershipAction(ctx, groupID, &groups.BatchMembershipActions{
		Leave: []groups.MemberID{{IdentityID: leaveIdentity}},
	})
	if err != nil {
		return fmt.Errorf("error leaving group: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Successfully left group %s as identity %s.\n", groupID, leaveIdentity)

	return nil
}
