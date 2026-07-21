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

var (
	updateName        string
	updateDescription string
)

// UpdateCmd represents the group update command
var UpdateCmd = &cobra.Command{
	Use:   "update GROUP_ID",
	Short: "Update a Globus group",
	Long: `Update settings for an existing Globus group.

You must be an administrator of the group to update it.

Examples:
  # Update group name
  globus group update GROUP_ID --name "New Group Name"

  # Update description
  globus group update GROUP_ID --description "Updated description"

  # Update both name and description
  globus group update GROUP_ID --name "New Name" --description "New description"`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdateGroup,
}

func init() {
	UpdateCmd.Flags().StringVar(&updateName, "name", "", "New name for the group")
	UpdateCmd.Flags().StringVar(&updateDescription, "description", "", "New description for the group")
}

func runUpdateGroup(cmd *cobra.Command, args []string) error {
	groupID := args[0]

	// Check if at least one flag is provided
	if updateName == "" && updateDescription == "" {
		return fmt.Errorf("at least one of --name or --description must be provided")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Prepare update request
	update := &groups.GroupUpdate{}

	if updateName != "" {
		update.Name = updateName
	}

	if updateDescription != "" {
		update.Description = updateDescription
	}

	// Update the group
	updatedGroup, err := groupsClient.UpdateGroup(ctx, groupID, update)
	if err != nil {
		return fmt.Errorf("error updating group: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Group updated successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Group ID:    %s\n", updatedGroup.ID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", updatedGroup.Name)
	fmt.Fprintf(os.Stdout, "Description: %s\n", updatedGroup.Description)

	return nil
}
