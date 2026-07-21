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

var joinIdentity string

// JoinCmd represents the group join command
var JoinCmd = &cobra.Command{
	Use:   "join GROUP_ID",
	Short: "Join a Globus group",
	Long: `Join a Globus group as a member.

The --identity flag specifies the identity ID that joins the group.

Examples:
  # Join a group as a specific identity
  globus group join GROUP_ID --identity IDENTITY_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runJoinGroup,
}

func init() {
	JoinCmd.Flags().StringVar(&joinIdentity, "identity", "", "Identity ID to join the group as (required)")
	_ = JoinCmd.MarkFlagRequired("identity")
}

func runJoinGroup(cmd *cobra.Command, args []string) error {
	groupID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Join the group via a single batch membership action
	// (POST /groups/{id}); the API has no dedicated join route.
	_, err = groupsClient.BatchMembershipAction(ctx, groupID, &groups.BatchMembershipActions{
		Join: []groups.MemberID{{IdentityID: joinIdentity}},
	})
	if err != nil {
		return fmt.Errorf("error joining group: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Successfully joined group %s as identity %s.\n", groupID, joinIdentity)

	return nil
}
