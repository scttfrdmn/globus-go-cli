// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// sapLimit is the --limit flag value for stream access point list.
var sapLimit int

// streamsAccessPointCmd returns "streams access-point", matching the Python
// CLI's "show" plus this CLI's "list" extension.
func streamsAccessPointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access-point",
		Short: "Inspect Globus Streams access points",
	}
	cmd.AddCommand(streamAccessPointShowCmd(), streamAccessPointListCmd())
	return cmd
}

// StreamAccessPointCmd returns the legacy top-level "stream-access-point"
// group, kept for compatibility. Prefer "globus streams access-point".
func StreamAccessPointCmd() *cobra.Command {
	sapCmd := &cobra.Command{
		Use:   "stream-access-point",
		Short: "Commands for Globus Streams access points (use 'streams access-point')",
		Long: `Commands for working with Globus Streams access points.

This group predates "globus streams access-point", which matches the Python
Globus CLI; prefer that.`,
		Deprecated: "use 'globus streams access-point' instead",
	}

	sapCmd.AddCommand(
		streamAccessPointShowCmd(),
		streamAccessPointListCmd(),
	)
	return sapCmd
}

// streamAccessPointShowCmd returns the stream access point show command.
func streamAccessPointShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show STREAM_ACCESS_POINT_ID",
		Short: "Show the attributes of a stream access point",
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
			renderAccessPoint(cmd.OutOrStdout(), sap)
			return nil
		},
	}
}

// renderAccessPoint prints the Python CLI's access point STANDARD_FIELDS,
// plus the host endpoint this CLI has always shown.
func renderAccessPoint(w io.Writer, sap *transfer.StreamAccessPoint) {
	printRecord(w, [][2]string{
		{"ID", sap.ID},
		{"Display Name", sap.DisplayName},
		{"Contact Email", sap.ContactEmail},
		{"Contact Info", sap.ContactInfo},
		{"Department", sap.Department},
		{"Description", sap.Description},
		{"Info Link", sap.InfoLink},
		{"Organization", sap.Organization},
		{"Host Endpoint", sap.EndpointID},
	})
}

// streamAccessPointListCmd returns the stream access point list command
// (a Go CLI extension; the Python CLI has no list).
func streamAccessPointListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Globus Streams access points",
		Long:  `List Globus Streams access points visible to the current user.`,
		Args:  cobra.NoArgs,
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
				ID           string
				DisplayName  string
				Organization string
				Domain       string
				HostEndpoint string
			}
			rows := make([]sapRow, 0, len(resp.Data))
			for _, s := range resp.Data {
				domain := ""
				if s.TLSFTPServer != "" {
					domain = endpointDomain(s.TLSFTPServer, "")
				}
				rows = append(rows, sapRow{
					ID:           s.ID,
					DisplayName:  strings.TrimSpace(s.DisplayName),
					Organization: s.Organization,
					Domain:       domain,
					HostEndpoint: s.EndpointID,
				})
			}
			return formatter.FormatOutput(rows, []string{"ID", "DisplayName", "Organization", "Domain", "HostEndpoint"})
		},
	}

	cmd.Flags().IntVar(&sapLimit, "limit", 25, "Maximum number of stream access points to return")

	return cmd
}
