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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Remove member from group via a single batch membership action
	// (POST /groups/{id}); the API has no dedicated remove-member route.
	_, err = groupsClient.BatchMembershipAction(ctx, groupID, &groups.BatchMembershipActions{
		Remove: []groups.MemberID{{IdentityID: identityID}},
	})
	if err != nil {
		return fmt.Errorf("error removing member: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Successfully removed member %s from group %s.\n", identityID, groupID)

	return nil
}
