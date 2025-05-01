// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

var (
	rmRecursive bool
	rmForce     bool
)

// RmCmd returns the rm command
func RmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm ENDPOINT_ID:PATH",
		Short: "Remove a file or directory on an endpoint",
		Long: `Remove a file or directory on a Globus endpoint.

This command deletes a file or directory on the specified Globus endpoint.
If --recursive is specified, it will delete directories and their contents.

Examples:
  globus transfer rm ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/file
  globus transfer rm --recursive ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoint ID and path
			endpointID, path := parseEndpointAndPath(args[0])
			
			// Check that path is specified
			if path == "/" {
				return fmt.Errorf("path must be specified for rm command")
			}
			
			return removeItem(cmd, endpointID, path)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&rmRecursive, "recursive", "r", false, "Remove directories and their contents recursively")
	cmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Force removal without confirmation")

	return cmd
}

// removeItem removes a file or directory on an endpoint
func removeItem(cmd *cobra.Command, endpointID, path string) error {
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

	// Check if we need to prompt for confirmation
	if !rmForce {
		// Get file/directory info
		options := &pkg.ListOptions{
			Path: path,
		}
		
		listing, err := transferClient.ListDirectory(ctx, endpointID, options)
		if err != nil {
			// If we can't get info, still prompt
			prompt := fmt.Sprintf("Are you sure you want to delete %s:%s?", endpointID, path)
			if !confirmAction(prompt) {
				fmt.Println("Operation canceled.")
				return nil
			}
		} else {
			// Check if it's a directory
			isDir := false
			for _, item := range listing.DATA {
				if item.Type == "dir" && item.Name == "." {
					isDir = true
					break
				}
			}
			
			if isDir {
				if !rmRecursive {
					return fmt.Errorf("%s is a directory. Use --recursive to remove directories", path)
				}
				
				// Count items in the directory
				count := len(listing.DATA)
				if count > 2 { // Accounting for "." and ".."
					prompt := fmt.Sprintf("Are you sure you want to delete directory %s:%s and all its contents (%d items)?", 
						endpointID, path, count-2)
					if !confirmAction(prompt) {
						fmt.Println("Operation canceled.")
						return nil
					}
				}
			} else {
				// It's a file
				prompt := fmt.Sprintf("Are you sure you want to delete file %s:%s?", endpointID, path)
				if !confirmAction(prompt) {
					fmt.Println("Operation canceled.")
					return nil
				}
			}
		}
	}

	// Delete the item
	err = transferClient.Delete(ctx, endpointID, path, rmRecursive)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	fmt.Printf("Successfully deleted %s:%s\n", endpointID, path)
	return nil
}

// confirmAction asks the user for confirmation
func confirmAction(prompt string) bool {
	confirm := promptui.Prompt{
		Label:     prompt,
		IsConfirm: true,
	}

	result, err := confirm.Run()
	if err != nil {
		return false
	}

	return strings.ToLower(result) == "y"
}