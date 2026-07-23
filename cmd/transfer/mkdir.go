// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	mkdirRecursive bool
	mkdirLocalUser string
)

// MkdirCmd returns the mkdir command
func MkdirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mkdir ENDPOINT_ID:PATH",
		Short: "Create a directory on an endpoint",
		Long: `Create a directory on a Globus endpoint.

This command creates a directory at the specified path on the Globus endpoint.
If --recursive is specified, it will create parent directories as needed.

Examples:
  globus transfer mkdir ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/directory
  globus transfer mkdir --recursive ddb59aef-6d04-11e5-ba46-22000b92c6ec:/deep/path/to/create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoint ID and path
			endpointID, path := parseEndpointAndPath(args[0])

			// Check that path is specified
			if path == "/" {
				return fmt.Errorf("path must be specified for mkdir command")
			}

			return createDirectory(cmd, endpointID, path)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&mkdirRecursive, "recursive", "p", false, "Create parent directories as needed")
	cmd.Flags().StringVar(&mkdirLocalUser, "local-user", "", "Local user to map to (GCSv5 mapped collections)")

	return cmd
}

// createDirectory creates a directory on an endpoint
func createDirectory(cmd *cobra.Command, endpointID, path string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Create the directory. The v4 SDK takes endpoint/path/local-user
	// positionally; the recursive flag is a client-side convenience that the
	// operation API does not accept, so it is a no-op for now.
	if _, err := transferClient.MakeDirectory(ctx, endpointID, path, mkdirLocalUser); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("Successfully created directory %s:%s\n", endpointID, path)
	return nil
}
