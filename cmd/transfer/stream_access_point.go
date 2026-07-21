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
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// sapLimit is the --limit flag value for stream-access-point list.
var sapLimit int

// StreamAccessPointCmd returns the stream-access-point subcommand group.
//
// Globus Streams access points (Python SDK v4.3.0+) are backed by the v4
// Transfer client. A stream access point exposes a collection path for use as a
// listener or initiator endpoint of a Globus Streams tunnel.
func StreamAccessPointCmd() *cobra.Command {
	sapCmd := &cobra.Command{
		Use:   "stream-access-point",
		Short: "Commands for Globus Streams access points",
		Long: `Commands for working with Globus Streams access points.

A stream access point exposes a collection path for use as the listener or
initiator endpoint of a Globus Streams tunnel.`,
	}

	sapCmd.AddCommand(
		streamAccessPointShowCmd(),
		streamAccessPointListCmd(),
	)
	return sapCmd
}

// streamAccessPointShowCmd returns the stream-access-point show command.
func streamAccessPointShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ACCESS_POINT_ID",
		Short: "Show details of a Globus Streams access point",
		Long:  `Show detailed information about a specific Globus Streams access point.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			sap, err := client.GetStreamAccessPoint(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get stream access point: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
				return formatter.FormatOutput(sap, nil)
			}

			fmt.Println("Stream Access Point Details:")
			fmt.Printf("  ID:           %s\n", sap.ID)
			fmt.Printf("  Tunnel ID:    %s\n", sap.TunnelID)
			fmt.Printf("  Endpoint ID:  %s\n", sap.EndpointID)
			fmt.Printf("  Path:         %s\n", sap.Path)
			fmt.Printf("  Access URL:   %s\n", sap.AccessURL)

			return nil
		},
	}
}

// streamAccessPointListCmd returns the stream-access-point list command.
func streamAccessPointListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Globus Streams access points",
		Long:  `List Globus Streams access points visible to the current user.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.ListStreamAccessPoints(ctx, &transfer.ListTunnelsOptions{Limit: sapLimit})
			if err != nil {
				return fmt.Errorf("failed to list stream access points: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON {
				// Emit the enveloped service document ({"DATA_TYPE","DATA":[...]}).
				return formatter.FormatOutput(resp, nil)
			}

			type sapRow struct {
				ID         string
				TunnelID   string
				EndpointID string
				Path       string
			}
			rows := make([]sapRow, 0, len(resp.Data))
			for _, s := range resp.Data {
				rows = append(rows, sapRow{
					ID:         s.ID,
					TunnelID:   s.TunnelID,
					EndpointID: s.EndpointID,
					Path:       s.Path,
				})
			}
			return formatter.FormatOutput(rows, []string{"ID", "TunnelID", "EndpointID", "Path"})
		},
	}

	cmd.Flags().IntVar(&sapLimit, "limit", 25, "Maximum number of stream access points to return")

	return cmd
}
