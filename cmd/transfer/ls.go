// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
)

var (
	lsRecursive bool
	lsLongFormat bool
	lsShowHidden bool
)

// LsCmd returns the ls command
func LsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls ENDPOINT_ID[:PATH]",
		Short: "List directory contents on an endpoint",
		Long: `List directory contents on a Globus endpoint.

This command lists the contents of a directory on the specified Globus endpoint.
The PATH is optional and defaults to the home directory or root of the endpoint.

Examples:
  globus transfer ls ddb59aef-6d04-11e5-ba46-22000b92c6ec
  globus transfer ls ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoint ID and path
			endpointID, path := parseEndpointAndPath(args[0])
			
			return listDirectory(cmd, endpointID, path)
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&lsRecursive, "recursive", "r", false, "List directories recursively")
	cmd.Flags().BoolVarP(&lsLongFormat, "long", "l", false, "List in long format with details")
	cmd.Flags().BoolVarP(&lsShowHidden, "all", "a", false, "Show hidden files")

	return cmd
}

// parseEndpointAndPath parses an endpoint ID and path from a string
func parseEndpointAndPath(s string) (endpointID, path string) {
	parts := strings.SplitN(s, ":", 2)
	endpointID = parts[0]
	
	if len(parts) > 1 {
		path = parts[1]
	} else {
		path = "/"
	}
	
	return endpointID, path
}

// listDirectory lists the contents of a directory on an endpoint
func listDirectory(cmd *cobra.Command, endpointID, path string) error {
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

	// Load client configuration - not used with direct client initialization in v0.9.10
	// We still load it for future use cases
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create a simple static token authorizer for v0.9.10
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	
	// Create a core authorizer adapter for v0.9.10 compatibility
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create transfer client with v0.9.10 compatible options
	transferOptions := []transfer.Option{
		transfer.WithAuthorizer(coreAuthorizer),
	}
	
	transferClient, err := transfer.NewClient(transferOptions...)
	if err != nil {
		return fmt.Errorf("failed to create transfer client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Prepare listing options
	options := &transfer.ListDirectoryOptions{
		EndpointID: endpointID,
		Path:       path,
		ShowHidden: lsShowHidden,
	}

	// Get the directory listing
	listing, err := transferClient.ListDirectory(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to list directory: %w", err)
	}

	// Get output format
	format := viper.GetString("format")
	
	// Format and display the results
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	
	// Define the headers based on format
	var headers []string
	if lsLongFormat {
		headers = []string{"Type", "Permissions", "User", "Group", "Size", "LastModified", "Name"}
	} else {
		headers = []string{"Type", "Name"}
	}
	
	// Create a slice of file entries for formatting
	type fileEntry struct {
		Type        string
		Permissions string
		User        string
		Group       string
		Size        int64
		LastModified string
		Name        string
	}
	
	// In SDK v0.9.10, the field is named Data instead of DATA
	entries := make([]fileEntry, 0, len(listing.Data))
	
	for _, item := range listing.Data {
		entry := fileEntry{
			Type:        getFileType(item.Type),
			Name:        item.Name,
		}
		
		if lsLongFormat {
			entry.Permissions = item.Permissions
			entry.User = item.User
			entry.Group = item.Group
			entry.Size = item.Size
			
			// Format last modified time
			t, err := time.Parse(time.RFC3339, item.LastModified)
			if err == nil {
				entry.LastModified = t.Format("Jan 02 15:04")
			} else {
				entry.LastModified = item.LastModified
			}
		}
		
		entries = append(entries, entry)
	}
	
	// Display the results using the formatter
	if err := formatter.FormatOutput(entries, headers); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}
	
	// Output the directory path
	fmt.Printf("\nDirectory: %s:%s\n", endpointID, path)
	fmt.Printf("Total: %d items\n", len(listing.Data))
	
	return nil
}

// getFileType returns a string representation of the file type
func getFileType(t string) string {
	switch t {
	case "dir":
		return "d"
	case "file":
		return "f"
	case "link":
		return "l"
	default:
		return "-"
	}
}