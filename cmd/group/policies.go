// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package group

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// PolicyCmd returns the policies command group for managing a group's policy
// settings (GET/PUT /groups/{id}/policies).
func PolicyCmd() *cobra.Command {
	policiesCmd := &cobra.Command{
		Use:   "policies",
		Short: "Manage group policies",
		Long: `Commands for viewing and updating the policy settings of a Globus group.

Group policies control high-assurance enforcement, visibility, join
requests, and signup fields.

Examples:
  # Show a group's policies
  globus group policies show GROUP_ID

  # Update a group's join-request policy
  globus group policies set GROUP_ID --join-requests`,
	}

	policiesCmd.AddCommand(policiesShowCmd())
	policiesCmd.AddCommand(policiesSetCmd())

	return policiesCmd
}

// policiesShowCmd shows a group's policy settings.
func policiesShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show GROUP_ID",
		Short: "Show a group's policy settings",
		Long: `Show the policy settings for a Globus group.

Examples:
  # Show group policies
  globus group policies show GROUP_ID

  # Show with JSON output
  globus group policies show GROUP_ID --format=json`,
		Args: cobra.ExactArgs(1),
		RunE: runPoliciesShow,
	}
}

func runPoliciesShow(cmd *cobra.Command, args []string) error {
	groupID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	policies, err := groupsClient.GetGroupPolicies(ctx, groupID)
	if err != nil {
		return fmt.Errorf("error getting group policies: %w", err)
	}

	format := viper.GetString("format")

	if format == "text" {
		fmt.Fprintf(cmd.OutOrStdout(), "Group Policies\n")
		fmt.Fprintf(cmd.OutOrStdout(), "==============\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "Is High Assurance:        %v\n", policies.IsHighAssurance)
		fmt.Fprintf(cmd.OutOrStdout(), "Group Visibility:         %s\n", policies.GroupVisibility)
		fmt.Fprintf(cmd.OutOrStdout(), "Group Members Visibility: %s\n", policies.GroupMembersVisibility)
		fmt.Fprintf(cmd.OutOrStdout(), "Join Requests:            %v\n", policies.JoinRequests)
		fmt.Fprintf(cmd.OutOrStdout(), "Signup Fields:            %s\n", strings.Join(policies.SignupFields, ", "))
		if policies.AuthenticationAssuranceTimeout != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Auth Assurance Timeout:   %d\n", *policies.AuthenticationAssuranceTimeout)
		}
		return nil
	}

	// JSON/CSV output emits the policies struct.
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	headers := []string{"IsHighAssurance", "GroupVisibility", "GroupMembersVisibility", "JoinRequests", "SignupFields", "AuthenticationAssuranceTimeout"}
	if err := formatter.FormatOutput(policies, headers); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}

var (
	policiesSetHighAssurance     bool
	policiesSetJoinRequests      bool
	policiesSetVisibility        string
	policiesSetMembersVisibility string
	policiesSetSignupFields      []string
	policiesSetAuthTimeout       int
)

// policiesSetCmd updates a group's policy settings. It performs a
// get-modify-put so that flags left unset preserve their existing values (the
// underlying PUT replaces the entire policies document).
func policiesSetCmd() *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set GROUP_ID",
		Short: "Update a group's policy settings",
		Long: `Update the policy settings for a Globus group.

Only the flags you provide are changed; unset fields keep their current
values (the update replaces the whole policies document, so it is applied
on top of the group's existing policies).

Examples:
  # Enable join requests
  globus group policies set GROUP_ID --join-requests

  # Set group visibility to private
  globus group policies set GROUP_ID --visibility private`,
		Args: cobra.ExactArgs(1),
		RunE: runPoliciesSet,
	}

	setCmd.Flags().BoolVar(&policiesSetHighAssurance, "high-assurance", false, "Enable high-assurance enforcement")
	setCmd.Flags().BoolVar(&policiesSetJoinRequests, "join-requests", false, "Allow join requests")
	setCmd.Flags().StringVar(&policiesSetVisibility, "visibility", "", "Group visibility (authenticated, private)")
	setCmd.Flags().StringVar(&policiesSetMembersVisibility, "members-visibility", "", "Group members visibility (members, managers)")
	setCmd.Flags().StringSliceVar(&policiesSetSignupFields, "signup-fields", nil, "Comma-separated list of fields required from users applying for membership (empty string to require none)")
	setCmd.Flags().IntVar(&policiesSetAuthTimeout, "authentication-timeout", 0, "Time in seconds before a user must re-authenticate to access a high-assurance group")

	return setCmd
}

func runPoliciesSet(cmd *cobra.Command, args []string) error {
	groupID := args[0]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Fetch existing policies and apply only the flags the user changed. The
	// PUT replaces the entire policies document, so we must preserve the
	// values of flags left unset.
	policies, err := groupsClient.GetGroupPolicies(ctx, groupID)
	if err != nil {
		return fmt.Errorf("error getting current group policies: %w", err)
	}

	if cmd.Flags().Changed("high-assurance") {
		policies.IsHighAssurance = policiesSetHighAssurance
	}
	if cmd.Flags().Changed("join-requests") {
		policies.JoinRequests = policiesSetJoinRequests
	}
	if cmd.Flags().Changed("visibility") {
		policies.GroupVisibility = policiesSetVisibility
	}
	if cmd.Flags().Changed("members-visibility") {
		policies.GroupMembersVisibility = policiesSetMembersVisibility
	}
	if cmd.Flags().Changed("signup-fields") {
		policies.SignupFields = policiesSetSignupFields
	}
	if cmd.Flags().Changed("authentication-timeout") {
		policies.AuthenticationAssuranceTimeout = &policiesSetAuthTimeout
	}

	if err := groupsClient.SetGroupPolicies(ctx, groupID, policies); err != nil {
		return fmt.Errorf("error setting group policies: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully updated policies for group %s.\n", groupID)

	return nil
}
