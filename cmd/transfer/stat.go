// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
)

// StatCmd returns the stat command
func StatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stat ENDPOINT_ID:PATH",
		Short: "Show metadata for a single path on an endpoint",
		Long: `Show metadata for a single file or directory on a Globus endpoint.

This command stats a single path and reports its name, type, size, last
modified time, and permissions.

Examples:
  globus transfer stat ddb59aef-6d04-11e5-ba46-22000b92c6ec:/path/to/file`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoint ID and path
			endpointID, path := parseEndpointAndPath(args[0])

			if path == "/" {
				return fmt.Errorf("path must be specified for stat command")
			}

			return statPath(cmd, endpointID, path)
		},
	}

	return cmd
}

// statPath stats a single path on an endpoint
func statPath(cmd *cobra.Command, endpointID, path string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Stat the path. GenericResponse is a map[string]interface{} passthrough.
	resp, err := transferClient.OperationStat(ctx, endpointID, path, "")
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	// For json/unix or a --jmespath/--jq expression, route through the shared
	// formatter (raw stat document). Otherwise render the text detail view.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(resp, nil)
	}

	// Output as text, guarding for keys that may be absent.
	fmt.Println("Path Details:")
	if v, ok := resp["name"]; ok {
		fmt.Printf("  Name:          %v\n", v)
	}
	if v, ok := resp["type"]; ok {
		fmt.Printf("  Type:          %v\n", v)
	}
	if v, ok := resp["size"]; ok {
		fmt.Printf("  Size:          %v\n", v)
	}
	if v, ok := resp["last_modified"]; ok {
		fmt.Printf("  Last Modified: %v\n", v)
	}
	if v, ok := resp["permissions"]; ok {
		fmt.Printf("  Permissions:   %v\n", v)
	}

	return nil
}
