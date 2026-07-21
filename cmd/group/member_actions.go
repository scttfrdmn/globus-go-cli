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

// runMemberAction applies a single membership action (built from the given
// identity ID) to a group via BatchMembershipAction and prints a success
// message. It centralizes the shared client/context/error handling for the
// approve/reject/accept/decline commands.
func runMemberAction(verb string, buildActions func(identityID string) *groups.BatchMembershipActions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
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

		// Apply the membership action via a single batch action
		// (POST /groups/{id}); the API has no dedicated per-action route.
		_, err = groupsClient.BatchMembershipAction(ctx, groupID, buildActions(identityID))
		if err != nil {
			return fmt.Errorf("error applying %s to member: %w", verb, err)
		}

		// Display success message
		fmt.Fprintf(os.Stdout, "Successfully applied %s to identity %s in group %s.\n", verb, identityID, groupID)

		return nil
	}
}

// memberApproveCmd approves a pending membership request.
func memberApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve GROUP_ID IDENTITY_ID",
		Short: "Approve a pending membership request",
		Long: `Approve a pending request to join a Globus group.

Examples:
  # Approve a membership request
  globus group member approve GROUP_ID IDENTITY_ID`,
		Args: cobra.ExactArgs(2),
		RunE: runMemberAction("approve", func(identityID string) *groups.BatchMembershipActions {
			return &groups.BatchMembershipActions{Approve: []groups.MemberID{{IdentityID: identityID}}}
		}),
	}
}

// memberRejectCmd rejects a pending membership request.
func memberRejectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reject GROUP_ID IDENTITY_ID",
		Short: "Reject a pending membership request",
		Long: `Reject a pending request to join a Globus group.

Examples:
  # Reject a membership request
  globus group member reject GROUP_ID IDENTITY_ID`,
		Args: cobra.ExactArgs(2),
		RunE: runMemberAction("reject", func(identityID string) *groups.BatchMembershipActions {
			return &groups.BatchMembershipActions{Reject: []groups.MemberID{{IdentityID: identityID}}}
		}),
	}
}

// memberAcceptCmd accepts a pending invitation to join a group.
func memberAcceptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "accept GROUP_ID IDENTITY_ID",
		Short: "Accept a pending group invitation",
		Long: `Accept a pending invitation to join a Globus group.

Examples:
  # Accept an invitation
  globus group member accept GROUP_ID IDENTITY_ID`,
		Args: cobra.ExactArgs(2),
		RunE: runMemberAction("accept", func(identityID string) *groups.BatchMembershipActions {
			return &groups.BatchMembershipActions{Accept: []groups.MemberID{{IdentityID: identityID}}}
		}),
	}
}

// memberDeclineCmd declines a pending invitation to join a group.
func memberDeclineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decline GROUP_ID IDENTITY_ID",
		Short: "Decline a pending group invitation",
		Long: `Decline a pending invitation to join a Globus group.

Examples:
  # Decline an invitation
  globus group member decline GROUP_ID IDENTITY_ID`,
		Args: cobra.ExactArgs(2),
		RunE: runMemberAction("decline", func(identityID string) *groups.BatchMembershipActions {
			return &groups.BatchMembershipActions{Decline: []groups.MemberID{{IdentityID: identityID}}}
		}),
	}
}
