// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

var (
	// Options for ls command
	recursiveFlag  bool
	longFormatFlag bool
	showHiddenFlag bool
)

// LsCmd returns the ls command
func LsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls ENDPOINT_ID[:PATH]",
		Short: "List contents of a Globus endpoint directory",
		Long: `List the contents of a directory on a Globus endpoint.

This command lists files and directories at the specified path on a Globus endpoint.
If no path is provided, the endpoint's default directory is used.

Examples:
  globus transfer ls endpoint_id                # List contents of the default directory
  globus transfer ls endpoint_id:/path/to/dir   # List contents of a specific directory
  globus transfer ls endpoint_id:/path -l       # List in long format
  globus transfer ls endpoint_id:/path -r       # List recursively`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listFiles(cmd, args[0])
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "List directories recursively")
	cmd.Flags().BoolVarP(&longFormatFlag, "long", "l", false, "Use long listing format")
	cmd.Flags().BoolVarP(&showHiddenFlag, "all", "a", false, "Show hidden files")

	return cmd
}

// listFiles lists files and directories at the specified path on a Globus endpoint
func listFiles(cmd *cobra.Command, arg string) error {
	// Parse the endpoint ID and path
	endpointID, path, err := parseEndpointPath(arg)
	if err != nil {
		return err
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Prepare options
	options := &pkg.ListFileOptions{
		ShowHidden: showHiddenFlag,
		Recursive:  recursiveFlag,
	}

	// List the files
	fileList, err := transferClient.ListFiles(ctx, endpointID, path, options)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Get output format
	format := viper.GetString("format")
	if format == "" {
		format = "text"
	}

	// Display the results based on format
	switch strings.ToLower(format) {
	case "json":
		// Output as JSON
		jsonData, err := json.MarshalIndent(fileList, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case "csv":
		// Output as CSV
		fmt.Println("name,type,size,last_modified,permissions")
		for _, file := range fileList.Data {
			lastModified := file.LastModified
			if lastModified == "" {
				lastModified = "N/A"
			}
			fmt.Printf("%s,%s,%d,%s,%s\n",
				strings.ReplaceAll(file.Name, ",", " "),
				file.Type,
				file.Size,
				lastModified,
				file.Permissions,
			)
		}
	default:
		// Output as text
		fmt.Printf("Contents of %s:%s\n\n", endpointID, fileList.Path)

		if longFormatFlag {
			// Long format listing
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "Type\tPermissions\tSize\tLast Modified\tName")
			fmt.Fprintln(w, "----\t-----------\t----\t-------------\t----")

			for _, file := range fileList.Data {
				// Format size and last modified time
				sizeStr := formatSize(file.Size)
				lastModified := formatTime(file.LastModified)

				// Colorize the output based on file type
				nameStr := file.Name
				if file.Type == "dir" {
					nameStr = color.BlueString(nameStr)
				} else if isExecutable(file.Permissions) {
					nameStr = color.GreenString(nameStr)
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					file.Type,
					file.Permissions,
					sizeStr,
					lastModified,
					nameStr,
				)
			}
			w.Flush()
		} else {
			// Simple listing
			for _, file := range fileList.Data {
				nameStr := file.Name
				if file.Type == "dir" {
					nameStr = color.BlueString(nameStr + "/")
				} else if isExecutable(file.Permissions) {
					nameStr = color.GreenString(nameStr + "*")
				}
				fmt.Println(nameStr)
			}
		}

		// Display count
		fmt.Printf("\nTotal: %d items\n", len(fileList.Data))
	}

	return nil
}

// parseEndpointPath parses an endpoint ID and path from a string in the format "endpoint_id[:path]"
func parseEndpointPath(arg string) (endpointID, path string, err error) {
	// Split the endpoint ID and path
	parts := strings.SplitN(arg, ":", 2)
	endpointID = parts[0]

	// Validate the endpoint ID
	if endpointID == "" {
		return "", "", fmt.Errorf("invalid endpoint ID")
	}

	// Get the path
	if len(parts) > 1 {
		path = parts[1]
	} else {
		path = "/"
	}

	return endpointID, path, nil
}

// formatSize formats a size in bytes as a human-readable string
func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
}

// formatTime formats a time string as a human-readable string
func formatTime(timeStr string) string {
	if timeStr == "" {
		return "N/A"
	}

	// Try to parse the time
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05",
	} {
		if t, err := time.Parse(layout, timeStr); err == nil {
			// Format the time
			return t.Format("Jan 02, 2006 15:04:05")
		}
	}

	// If we couldn't parse it, return as is
	return timeStr
}

// isExecutable checks if a file is executable based on its permissions
func isExecutable(permissions string) bool {
	// Check if the permissions string contains an 'x'
	return strings.Contains(permissions, "x")
}