// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// IndexRoleDeleteCmd represents the search index role delete command
var IndexRoleDeleteCmd = &cobra.Command{
	Use:   "delete INDEX_ID ROLE_ID",
	Short: "Delete a role from a Globus Search index",
	Long: `Delete a role assignment from a Globus Search index.

Examples:
  # Delete a role
  globus search index role delete INDEX_ID ROLE_ID`,
	Args: cobra.ExactArgs(2),
	RunE: runIndexRoleDelete,
}

func runIndexRoleDelete(cmd *cobra.Command, args []string) error {
	indexID := args[0]
	roleID := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	if err := searchClient.RemoveRole(ctx, indexID, roleID); err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}

	fmt.Printf("Deleted role %s from index %s\n", roleID, indexID)
	return nil
}
