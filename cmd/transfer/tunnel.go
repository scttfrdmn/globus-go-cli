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

// Flag variables for tunnel subcommands.
var (
	tunnelLimit        int
	tunnelListener     string
	tunnelInitiator    string
	tunnelLabel        string
	tunnelListenerPort int
	tunnelListenerIP   string
	tunnelLifetimeMins int
	tunnelRestartable  bool
)

// TunnelCmd returns the tunnel subcommand group.
//
// Globus Streams tunnels (Python SDK v4.3.0+) are backed by the v4 Transfer
// client. A tunnel connects a listener stream access point to an initiator
// stream access point, enabling low-latency streaming data movement.
func TunnelCmd() *cobra.Command {
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Commands for managing Globus Streams tunnels",
		Long: `Commands for managing Globus Streams tunnels.

A tunnel connects a listener stream access point to an initiator stream access
point, enabling streaming data movement between Globus collections.`,
	}

	tunnelCmd.AddCommand(
		tunnelListCmd(),
		tunnelShowCmd(),
		tunnelCreateCmd(),
		tunnelUpdateCmd(),
		tunnelDeleteCmd(),
		tunnelEventsCmd(),
	)

	return tunnelCmd
}

// tunnelListCmd returns the tunnel list command.
func tunnelListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Globus Streams tunnels",
		Long:  `List Globus Streams tunnels visible to the current user.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.ListTunnels(ctx, &transfer.ListTunnelsOptions{Limit: tunnelLimit})
			if err != nil {
				return fmt.Errorf("failed to list tunnels: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON {
				// Emit the enveloped service document ({"DATA":[...],...}).
				return formatter.FormatOutput(resp, nil)
			}

			type tunnelRow struct {
				ID     string
				Label  string
				Status string
				Source string
				Owner  string
			}
			rows := make([]tunnelRow, 0, len(resp.Tunnels))
			for _, t := range resp.Tunnels {
				rows = append(rows, tunnelRow{
					ID:     t.ID,
					Label:  t.DisplayName,
					Status: t.Status,
					Source: t.SourceEndpointID + ":" + t.SourcePath,
					Owner:  t.Owner,
				})
			}
			return formatter.FormatOutput(rows, []string{"ID", "Label", "Status", "Source", "Owner"})
		},
	}

	cmd.Flags().IntVar(&tunnelLimit, "limit", 25, "Maximum number of tunnels to return")

	return cmd
}

// tunnelShowCmd returns the tunnel show command.
func tunnelShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show TUNNEL_ID",
		Short: "Show a Globus Streams tunnel",
		Long:  `Show detailed information about a specific Globus Streams tunnel.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			tunnel, err := client.GetTunnel(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get tunnel: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
				return formatter.FormatOutput(tunnel, nil)
			}

			fmt.Println("Tunnel Details:")
			fmt.Printf("  ID:             %s\n", tunnel.ID)
			fmt.Printf("  Display Name:   %s\n", tunnel.DisplayName)
			fmt.Printf("  Status:         %s\n", tunnel.Status)
			fmt.Printf("  Source:         %s:%s\n", tunnel.SourceEndpointID, tunnel.SourcePath)
			fmt.Printf("  Owner:          %s\n", tunnel.Owner)
			if tunnel.ExpiresAt != nil {
				fmt.Printf("  Expires At:     %s\n", tunnel.ExpiresAt.Format(time.RFC3339))
			}

			return nil
		},
	}
}

// tunnelCreateCmd returns the tunnel create command.
func tunnelCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Globus Streams tunnel",
		Long: `Create a Globus Streams tunnel connecting a listener stream access
point to an initiator stream access point.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			create := &transfer.TunnelCreate{
				ListenerStreamAccessPoint:  tunnelListener,
				InitiatorStreamAccessPoint: tunnelInitiator,
				Label:                      tunnelLabel,
				ListenerIPAddress:          tunnelListenerIP,
			}
			if cmd.Flags().Changed("listener-port") {
				port := tunnelListenerPort
				create.ListenerPort = &port
			}
			if cmd.Flags().Changed("lifetime-mins") {
				mins := tunnelLifetimeMins
				create.LifetimeMins = &mins
			}
			if cmd.Flags().Changed("restartable") {
				restartable := tunnelRestartable
				create.Restartable = &restartable
			}

			tunnel, err := client.CreateTunnel(ctx, create)
			if err != nil {
				return fmt.Errorf("failed to create tunnel: %w", err)
			}

			fmt.Printf("Created tunnel %s (status: %s)\n", tunnel.ID, tunnel.Status)
			return nil
		},
	}

	cmd.Flags().StringVar(&tunnelListener, "listener", "", "Listener stream access point ID (required)")
	cmd.Flags().StringVar(&tunnelInitiator, "initiator", "", "Initiator stream access point ID (required)")
	cmd.Flags().StringVar(&tunnelLabel, "label", "", "Human-readable label for the tunnel")
	cmd.Flags().IntVar(&tunnelListenerPort, "listener-port", 0, "Listener port")
	cmd.Flags().StringVar(&tunnelListenerIP, "listener-ip", "", "Listener IP address")
	cmd.Flags().IntVar(&tunnelLifetimeMins, "lifetime-mins", 0, "Tunnel lifetime in minutes")
	cmd.Flags().BoolVar(&tunnelRestartable, "restartable", false, "Whether the tunnel is restartable")

	_ = cmd.MarkFlagRequired("listener")
	_ = cmd.MarkFlagRequired("initiator")

	return cmd
}

// tunnelUpdateCmd returns the tunnel update command.
func tunnelUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update TUNNEL_ID",
		Short: "Update a Globus Streams tunnel",
		Long:  `Update the mutable fields of a Globus Streams tunnel.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			update := &transfer.TunnelUpdate{
				Label:             tunnelLabel,
				ListenerIPAddress: tunnelListenerIP,
			}
			if cmd.Flags().Changed("listener-port") {
				port := tunnelListenerPort
				update.ListenerPort = &port
			}

			if _, err := client.UpdateTunnel(ctx, args[0], update); err != nil {
				return fmt.Errorf("failed to update tunnel: %w", err)
			}

			fmt.Printf("Updated tunnel %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&tunnelLabel, "label", "", "Human-readable label for the tunnel")
	cmd.Flags().IntVar(&tunnelListenerPort, "listener-port", 0, "Listener port")
	cmd.Flags().StringVar(&tunnelListenerIP, "listener-ip", "", "Listener IP address")

	return cmd
}

// tunnelDeleteCmd returns the tunnel delete command.
func tunnelDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete TUNNEL_ID",
		Short: "Delete a Globus Streams tunnel",
		Long:  `Delete a Globus Streams tunnel.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			if err := client.DeleteTunnel(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete tunnel: %w", err)
			}

			fmt.Printf("Deleted tunnel %s\n", args[0])
			return nil
		},
	}
}

// tunnelEventsCmd returns the tunnel events command.
func tunnelEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events TUNNEL_ID",
		Short: "Show events for a Globus Streams tunnel",
		Long:  `Show the event history for a specific Globus Streams tunnel.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client, err := getClient(ctx)
			if err != nil {
				return err
			}

			resp, err := client.GetTunnelEvents(ctx, args[0], &transfer.ListTunnelEventsOptions{Limit: tunnelLimit})
			if err != nil {
				return fmt.Errorf("failed to get tunnel events: %w", err)
			}

			formatter := output.NewFormatter(viper.GetString("format"), cmd.OutOrStdout())
			if formatter.Format == output.FormatJSON {
				return formatter.FormatOutput(resp, nil)
			}

			type eventRow struct {
				ID          string
				Code        string
				Description string
				OccurredAt  string
			}
			rows := make([]eventRow, 0, len(resp.Events))
			for _, e := range resp.Events {
				occurred := ""
				if e.OccurredAt != nil {
					occurred = e.OccurredAt.Format(time.RFC3339)
				}
				rows = append(rows, eventRow{
					ID:          e.ID,
					Code:        e.Code,
					Description: e.Description,
					OccurredAt:  occurred,
				})
			}
			return formatter.FormatOutput(rows, []string{"ID", "Code", "Description", "OccurredAt"})
		},
	}

	cmd.Flags().IntVar(&tunnelLimit, "limit", 25, "Maximum number of events to return")

	return cmd
}
