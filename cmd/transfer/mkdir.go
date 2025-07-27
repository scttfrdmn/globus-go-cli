// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

var (
	mkdirRecursive bool
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

	return cmd
}

// createDirectory creates a directory on an endpoint
func createDirectory(cmd *cobra.Command, endpointID, path string) error {
	// Get current profile
	profile := viper.GetString("profile")

	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration - not used with direct client initialization in v0.9.17
	// We still load it for future use cases
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create a simple static token authorizer for v0.9.17
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)

	// Create a core authorizer adapter for v0.9.17 compatibility
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create transfer client with v0.9.17 compatible options
	transferOptions := []transfer.Option{
		transfer.WithAuthorizer(coreAuthorizer),
	}

	transferClient, err := transfer.NewClient(transferOptions...)
	if err != nil {
		return fmt.Errorf("failed to create transfer client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create directory options
	options := &transfer.CreateDirectoryOptions{
		EndpointID: endpointID,
		Path:       path,
	}

	// Create the directory
	err = transferClient.CreateDirectory(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("Successfully created directory %s:%s\n", endpointID, path)
	return nil
}
