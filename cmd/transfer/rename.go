// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// RenameCmd returns the rename command
func RenameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename ENDPOINT_ID:OLD_PATH NEW_PATH",
		Short: "Rename a file or directory on an endpoint",
		Long: `Rename a file or directory on a Globus endpoint.

This command renames a path on the specified Globus endpoint. The first
argument is the endpoint and current path; the second argument is the new
path on the same endpoint.

Examples:
  globus transfer rename ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/old.txt /path/new.txt
  globus transfer rename ddb59aef-6d04-11e5-ba46-22000b92c6ec:/dir/old /dir/new`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoint ID and the current path; the second arg is the
			// new path on the same endpoint.
			endpointID, oldPath := parseEndpointAndPath(args[0])
			newPath := args[1]

			if oldPath == "/" {
				return fmt.Errorf("a source path must be specified for rename command")
			}

			return renamePath(cmd, endpointID, oldPath, newPath)
		},
	}

	return cmd
}

// renamePath renames a file or directory on an endpoint
func renamePath(cmd *cobra.Command, endpointID, oldPath, newPath string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Rename the path. The v4 SDK takes endpoint/old-path/new-path/local-user
	// positionally; local-user is optional (pass "").
	if _, err := transferClient.Rename(ctx, endpointID, oldPath, newPath, ""); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	fmt.Printf("Successfully renamed %s:%s to %s\n", endpointID, oldPath, newPath)
	return nil
}
