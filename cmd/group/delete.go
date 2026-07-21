// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package group

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var deleteConfirm bool

// DeleteCmd represents the group delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete GROUP_ID",
	Short: "Delete a Globus group",
	Long: `Delete a Globus group permanently.

WARNING: This action cannot be undone. All group data, memberships,
and associated resources will be removed.

You must be an administrator of the group to delete it.

Examples:
  # Delete a group (requires confirmation)
  globus group delete GROUP_ID

  # Delete without confirmation prompt
  globus group delete GROUP_ID --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteGroup,
}

func init() {
	DeleteCmd.Flags().BoolVar(&deleteConfirm, "confirm", false, "Skip confirmation prompt")
}

func runDeleteGroup(cmd *cobra.Command, args []string) error {
	groupID := args[0]

	// Confirm deletion unless --confirm flag is used
	if !deleteConfirm {
		fmt.Fprintf(os.Stdout, "WARNING: This will permanently delete the group and all its data.\n")
		fmt.Fprintf(os.Stdout, "Group ID: %s\n\n", groupID)
		fmt.Fprintf(os.Stdout, "Are you sure you want to delete this group? (yes/no): ")

		var response string
		_, _ = fmt.Scanln(&response)

		if response != "yes" && response != "y" {
			fmt.Fprintf(os.Stdout, "Deletion cancelled.\n")
			return nil
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Delete the group
	err = groupsClient.DeleteGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("error deleting group: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Group %s deleted successfully.\n", groupID)

	return nil
}
