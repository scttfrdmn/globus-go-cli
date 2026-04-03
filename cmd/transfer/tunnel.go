// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

// TunnelCmd returns the tunnel subcommand group.
// Globus Streams tunnels provide persistent channels for streaming data between
// endpoints. Added in Python SDK v4.3.0 / Go SDK v4.3.0-1.
func TunnelCmd() *cobra.Command {
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Commands for managing Globus Streams tunnels",
		Long: `Commands for managing Globus Streams tunnels.

Tunnels provide persistent channels for streaming data between Globus endpoints.
They are the foundation of the Globus Streams service, enabling real-time data
access without staging files through Transfer tasks.

Examples:
  globus transfer tunnel list
  globus transfer tunnel create --name "My Tunnel" --source-endpoint EP_ID
  globus transfer tunnel show TUNNEL_ID
  globus transfer tunnel events TUNNEL_ID`,
	}

	tunnelCmd.AddCommand(
		tunnelListCmd(),
		tunnelCreateCmd(),
		tunnelShowCmd(),
		tunnelUpdateCmd(),
		tunnelDeleteCmd(),
		tunnelEventsCmd(),
	)

	return tunnelCmd
}

// newTransferClient is a helper that loads credentials and returns a ready transfer client.
func newTransferClient(profile string) (*transfer.Client, error) {
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return nil, fmt.Errorf("not logged in: %w", err)
	}
	if !authcmd.IsTokenValid(tokenInfo) {
		return nil, fmt.Errorf("token is expired, please login again")
	}
	if _, err = config.LoadClientConfig(); err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}
	coreAuthorizer := authorizers.ToCore(authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken))
	client, err := transfer.NewClient(transfer.WithAuthorizer(coreAuthorizer))
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer client: %w", err)
	}
	return client, nil
}

// tunnelListCmd returns the tunnel list command.
func tunnelListCmd() *cobra.Command {
	var tunnelLimit int
	var tunnelMarker string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Globus Streams tunnels",
		Long: `List Globus Streams tunnels owned by the current user.

Examples:
  globus transfer tunnel list
  globus transfer tunnel list --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := viper.GetString("profile")
			client, err := newTransferClient(profile)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			opts := &transfer.ListTunnelsOptions{Limit: tunnelLimit, Marker: tunnelMarker}
			list, err := client.ListTunnels(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to list tunnels: %w", err)
			}

			format := viper.GetString("format")
			formatter := output.NewFormatter(format, cmd.OutOrStdout())
			headers := []string{"ID", "DisplayName", "Status", "SourceEndpointID", "SourcePath"}

			type tunnelEntry struct {
				ID               string
				DisplayName      string
				Status           string
				SourceEndpointID string
				SourcePath       string
			}
			entries := make([]tunnelEntry, 0, len(list.Tunnels))
			for _, t := range list.Tunnels {
				entries = append(entries, tunnelEntry{
					ID:               t.ID,
					DisplayName:      t.DisplayName,
					Status:           t.Status,
					SourceEndpointID: t.SourceEndpointID,
					SourcePath:       t.SourcePath,
				})
			}
			return formatter.FormatOutput(entries, headers)
		},
	}

	cmd.Flags().IntVar(&tunnelLimit, "limit", 25, "Maximum number of tunnels to return")
	cmd.Flags().StringVar(&tunnelMarker, "marker", "", "Pagination marker for next page")
	return cmd
}

// tunnelCreateCmd returns the tunnel create command.
func tunnelCreateCmd() *cobra.Command {
	var name string
	var sourceEndpoint string
	var sourcePath string
	var expiresIn int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Globus Streams tunnel",
		Long: `Create a new Globus Streams tunnel.

A tunnel provides a persistent channel for streaming data from a source endpoint.

Examples:
  globus transfer tunnel create --name "My Tunnel" --source-endpoint EP_ID
  globus transfer tunnel create --name "My Tunnel" --source-endpoint EP_ID \
    --source-path /data/streams --expires-in 3600`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := viper.GetString("profile")
			client, err := newTransferClient(profile)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			data := &transfer.CreateTunnelData{
				DisplayName:      name,
				SourceEndpointID: sourceEndpoint,
				SourcePath:       sourcePath,
			}
			if expiresIn > 0 {
				data.ExpiresIn = &expiresIn
			}

			tunnel, err := client.CreateTunnel(ctx, data)
			if err != nil {
				return fmt.Errorf("failed to create tunnel: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Tunnel created successfully!\n\n")
			fmt.Fprintf(cmd.OutOrStdout(), "ID:               %s\n", tunnel.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Display Name:     %s\n", tunnel.DisplayName)
			fmt.Fprintf(cmd.OutOrStdout(), "Status:           %s\n", tunnel.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Source Endpoint:  %s\n", tunnel.SourceEndpointID)
			if tunnel.SourcePath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Source Path:      %s\n", tunnel.SourcePath)
			}
			if tunnel.ExpiresAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Expires At:       %s\n", tunnel.ExpiresAt.Format(time.RFC3339))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Display name for the tunnel (required)")
	cmd.Flags().StringVar(&sourceEndpoint, "source-endpoint", "", "Source endpoint ID (required)")
	cmd.Flags().StringVar(&sourcePath, "source-path", "", "Path on the source endpoint")
	cmd.Flags().IntVar(&expiresIn, "expires-in", 0, "Tunnel lifetime in seconds (0 = no expiration)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("source-endpoint")
	return cmd
}

// tunnelShowCmd returns the tunnel show command.
func tunnelShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show TUNNEL_ID",
		Short: "Show details of a Globus Streams tunnel",
		Long: `Show details for a specific Globus Streams tunnel.

Examples:
  globus transfer tunnel show TUNNEL_ID`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelID := args[0]
			profile := viper.GetString("profile")
			client, err := newTransferClient(profile)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			tunnel, err := client.GetTunnel(ctx, tunnelID)
			if err != nil {
				return fmt.Errorf("failed to get tunnel: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "ID:               %s\n", tunnel.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Display Name:     %s\n", tunnel.DisplayName)
			fmt.Fprintf(cmd.OutOrStdout(), "Owner:            %s\n", tunnel.Owner)
			fmt.Fprintf(cmd.OutOrStdout(), "Status:           %s\n", tunnel.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "Source Endpoint:  %s\n", tunnel.SourceEndpointID)
			if tunnel.SourcePath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Source Path:      %s\n", tunnel.SourcePath)
			}
			if tunnel.CreatedAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Created At:       %s\n", tunnel.CreatedAt.Format(time.RFC3339))
			}
			if tunnel.UpdatedAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Updated At:       %s\n", tunnel.UpdatedAt.Format(time.RFC3339))
			}
			if tunnel.ExpiresAt != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Expires At:       %s\n", tunnel.ExpiresAt.Format(time.RFC3339))
			}
			return nil
		},
	}
}

// tunnelUpdateCmd returns the tunnel update command.
func tunnelUpdateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update TUNNEL_ID",
		Short: "Update a Globus Streams tunnel",
		Long: `Update the display name of a Globus Streams tunnel.

Examples:
  globus transfer tunnel update TUNNEL_ID --name "New Name"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelID := args[0]
			profile := viper.GetString("profile")
			client, err := newTransferClient(profile)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			data := &transfer.UpdateTunnelData{}
			if cmd.Flags().Changed("name") {
				data.DisplayName = name
			}

			tunnel, err := client.UpdateTunnel(ctx, tunnelID, data)
			if err != nil {
				return fmt.Errorf("failed to update tunnel: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Tunnel updated successfully!\n\n")
			fmt.Fprintf(cmd.OutOrStdout(), "ID:           %s\n", tunnel.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", tunnel.DisplayName)
			fmt.Fprintf(cmd.OutOrStdout(), "Status:       %s\n", tunnel.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New display name for the tunnel")
	return cmd
}

// tunnelDeleteCmd returns the tunnel delete command.
func tunnelDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete TUNNEL_ID",
		Short: "Delete a Globus Streams tunnel",
		Long: `Delete a Globus Streams tunnel permanently.

Examples:
  globus transfer tunnel delete TUNNEL_ID`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelID := args[0]
			profile := viper.GetString("profile")
			client, err := newTransferClient(profile)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := client.DeleteTunnel(ctx, tunnelID); err != nil {
				return fmt.Errorf("failed to delete tunnel: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Tunnel %s deleted successfully.\n", tunnelID)
			return nil
		},
	}
}

// tunnelEventsCmd returns the tunnel events command.
func tunnelEventsCmd() *cobra.Command {
	var eventsLimit int
	var eventsMarker string

	cmd := &cobra.Command{
		Use:   "events TUNNEL_ID",
		Short: "List events for a Globus Streams tunnel",
		Long: `List events associated with a Globus Streams tunnel.

Events record state changes and activity for a tunnel over its lifetime.
Added in Python SDK v4.4.0.

Examples:
  globus transfer tunnel events TUNNEL_ID
  globus transfer tunnel events TUNNEL_ID --limit 100`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tunnelID := args[0]
			profile := viper.GetString("profile")
			client, err := newTransferClient(profile)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			opts := &transfer.ListTunnelEventsOptions{Limit: eventsLimit, Marker: eventsMarker}
			eventList, err := client.GetTunnelEvents(ctx, tunnelID, opts)
			if err != nil {
				return fmt.Errorf("failed to get tunnel events: %w", err)
			}

			format := viper.GetString("format")
			formatter := output.NewFormatter(format, cmd.OutOrStdout())
			headers := []string{"ID", "Code", "Description", "OccurredAt"}

			type eventEntry struct {
				ID          string
				Code        string
				Description string
				OccurredAt  string
			}
			entries := make([]eventEntry, 0, len(eventList.Events))
			for _, e := range eventList.Events {
				occurredAt := ""
				if e.OccurredAt != nil {
					occurredAt = e.OccurredAt.Format(time.RFC3339)
				}
				entries = append(entries, eventEntry{
					ID:          e.ID,
					Code:        e.Code,
					Description: e.Description,
					OccurredAt:  occurredAt,
				})
			}
			return formatter.FormatOutput(entries, headers)
		},
	}

	cmd.Flags().IntVar(&eventsLimit, "limit", 25, "Maximum number of events to return")
	cmd.Flags().StringVar(&eventsMarker, "marker", "", "Pagination marker for next page")
	return cmd
}
