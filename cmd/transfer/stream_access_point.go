// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// StreamAccessPointCmd returns the stream-access-point subcommand group.
// Stream Access Points provide real-time data stream access via Globus Streams.
// Added in Python SDK v4.3.0.
func StreamAccessPointCmd() *cobra.Command {
	sapCmd := &cobra.Command{
		Use:   "stream-access-point",
		Short: "Commands for Globus Streams access points",
		Long: `Commands for working with Globus Streams access points.

Stream Access Points provide URL-based access to real-time data streams
associated with a Globus Streams tunnel.

Examples:
  globus transfer stream-access-point show ACCESS_POINT_ID`,
	}

	sapCmd.AddCommand(streamAccessPointShowCmd())
	return sapCmd
}

// streamAccessPointShowCmd returns the stream-access-point show command.
func streamAccessPointShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ACCESS_POINT_ID",
		Short: "Show details of a Globus Streams access point",
		Long: `Show details for a specific Globus Streams access point.

An access point provides a URL for accessing a real-time data stream
from a Globus Streams tunnel.

Examples:
  globus transfer stream-access-point show ACCESS_POINT_ID`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accessPointID := args[0]
			profile := viper.GetString("profile")
			client, err := newTransferClient(profile)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			ap, err := client.GetStreamAccessPoint(ctx, accessPointID)
			if err != nil {
				return fmt.Errorf("failed to get stream access point: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "ID:           %s\n", ap.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Tunnel ID:    %s\n", ap.TunnelID)
			fmt.Fprintf(cmd.OutOrStdout(), "Endpoint ID:  %s\n", ap.EndpointID)
			fmt.Fprintf(cmd.OutOrStdout(), "Path:         %s\n", ap.Path)
			fmt.Fprintf(cmd.OutOrStdout(), "Access URL:   %s\n", ap.AccessURL)
			if ap.ExpiresAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Expires At:   %s\n", ap.ExpiresAt.Format(time.RFC3339))
			}
			return nil
		},
	}
}
