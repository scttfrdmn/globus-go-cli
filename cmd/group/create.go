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
	createName        string
	createDescription string
	createPublic      bool
	createParentID    string
)

// CreateCmd represents the group create command
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Globus group",
	Long: `Create a new Globus group with specified settings.

Groups provide collaborative access control and membership management
for Globus resources.

Examples:
  # Create a basic group
  globus group create --name "My Research Group" --description "Group for our research project"

  # Create a public group
  globus group create --name "Public Data Group" --description "Open access group" --public

Output:
  Displays the newly created group ID and details.`,
	RunE: runCreateGroup,
}

func init() {
	CreateCmd.Flags().StringVar(&createName, "name", "", "Name for the new group (required)")
	CreateCmd.Flags().StringVar(&createDescription, "description", "", "Description of the group")
	CreateCmd.Flags().BoolVar(&createPublic, "public", false, "Make the group publicly visible")
	CreateCmd.Flags().StringVar(&createParentID, "parent-id", "", "Make the new group a subgroup of the specified parent group")
	_ = CreateCmd.MarkFlagRequired("name")
}

func runCreateGroup(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Groups client authorized for the current profile.
	groupsClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Prepare group creation request
	newGroup := &groups.GroupCreate{
		Name:        createName,
		Description: createDescription,
		PublicGroup: createPublic,
		ParentID:    createParentID,
	}

	// Create the group
	createdGroup, err := groupsClient.CreateGroup(ctx, newGroup)
	if err != nil {
		return fmt.Errorf("error creating group: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Group created successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Group ID:    %s\n", createdGroup.ID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", createdGroup.Name)
	fmt.Fprintf(os.Stdout, "Description: %s\n", createdGroup.Description)
	fmt.Fprintf(os.Stdout, "Identity ID: %s\n", createdGroup.IdentityID)
	fmt.Fprintf(os.Stdout, "\nYou can now add members and configure the group.\n")

	return nil
}
