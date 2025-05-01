// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
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

	// Load client configuration
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create SDK config
	sdkConfig := pkg.NewConfig().
		WithClientID(clientCfg.ClientID).
		WithClientSecret(clientCfg.ClientSecret)

	// Create transfer client
	transferClient := sdkConfig.NewTransferClient(tokenInfo.AccessToken)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the directory
	err = transferClient.MakeDir(ctx, endpointID, path, mkdirRecursive)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fmt.Printf("Successfully created directory %s:%s\n", endpointID, path)
	return nil
}