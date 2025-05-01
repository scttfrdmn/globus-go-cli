// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
)

// MkdirCmd returns the mkdir command
func MkdirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mkdir ENDPOINT_ID[:PATH]",
		Short: "Create a directory on a Globus endpoint",
		Long: `Create a directory on a Globus endpoint.

This command creates a new directory at the specified path on a Globus endpoint.

Examples:
  globus transfer mkdir endpoint_id:/path/to/new_dir    # Create a new directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return makeDirectory(cmd, args[0])
		},
	}

	return cmd
}

// makeDirectory creates a directory on a Globus endpoint
func makeDirectory(cmd *cobra.Command, arg string) error {
	// Parse the endpoint ID and path
	endpointID, path, err := parseEndpointPath(arg)
	if err != nil {
		return fmt.Errorf("invalid endpoint path: %w", err)
	}

	// Ensure the path is provided
	if path == "" || path == "/" {
		return fmt.Errorf("a directory path must be provided")
	}

	// Get the transfer client
	transferClient, err := getTransferClient()
	if err != nil {
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the directory
	fmt.Printf("Creating directory on %s:%s\n", endpointID, path)
	
	err = transferClient.MakeDirectory(ctx, endpointID, path)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("Directory created successfully: %s:%s\n", endpointID, path)
	return nil
}