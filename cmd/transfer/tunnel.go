// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"io"
	"strconv"
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

// tunnelStopState is the lifecycle state "streams tunnel stop" requests,
// matching the Python CLI.
const tunnelStopState = "STOPPING"

// StreamsCmd returns the "streams" command group, matching the Python
// globus-cli (3.41.0+): "streams tunnel ..." and "streams access-point ...".
func StreamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "streams",
		Short: "Manage Globus Streams tunnels and access points",
		Long: `Manage Globus Streams tunnels and stream access points.

A tunnel connects an initiator stream access point to a listener stream access
point, carrying streaming data between them.`,
	}
	cmd.AddCommand(streamsTunnelCmd(), streamsAccessPointCmd())
	return cmd
}

// streamsTunnelCmd returns "streams tunnel" with the Python CLI's argument
// shapes (e.g. "create INITIATOR_STREAM_AP_ID LISTENER_STREAM_AP_ID").
func streamsTunnelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Manage Globus Streams tunnels",
	}
	cmd.AddCommand(
		tunnelListCmd(),
		tunnelShowCmd(),
		streamsTunnelCreateCmd(),
		streamsTunnelUpdateCmd(),
		tunnelStopCmd(),
		tunnelDeleteCmd(),
		tunnelEventsCmd(),
	)
	return cmd
}

// TunnelCmd returns the legacy top-level "tunnel" group, kept for
// compatibility with earlier releases of this CLI. New scripts should use
// "globus streams tunnel", which matches the Python CLI.
func TunnelCmd() *cobra.Command {
	tunnelCmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Commands for managing Globus Streams tunnels (use 'streams tunnel')",
		Long: `Commands for managing Globus Streams tunnels.

This group predates "globus streams tunnel", which matches the Python Globus
CLI; prefer that. Note "tunnel create" takes --initiator/--listener flags,
whereas "streams tunnel create" takes them as positional arguments.`,
		Deprecated: "use 'globus streams tunnel' instead",
	}

	tunnelCmd.AddCommand(
		tunnelListCmd(),
		tunnelShowCmd(),
		tunnelCreateCmd(),
		tunnelUpdateCmd(),
		tunnelStopCmd(),
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
		Args:  cobra.NoArgs,
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

			// Python CLI TUNNEL_LIST_FIELDS: ID, Label, State, Created Time.
			type tunnelRow struct {
				ID          string
				Label       string
				State       string
				CreatedTime string
			}
			rows := make([]tunnelRow, 0, len(resp.Tunnels))
			for _, t := range resp.Tunnels {
				rows = append(rows, tunnelRow{
					ID:          t.ID,
					Label:       t.Label,
					State:       t.State,
					CreatedTime: formatTime(t.CreatedAt),
				})
			}
			return formatter.FormatOutput(rows, []string{"ID", "Label", "State", "CreatedTime"})
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
		Long:  `Show the attributes of a Globus Streams tunnel.`,
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
			return displayTunnel(cmd.OutOrStdout(), tunnel)
		},
	}
}

// displayTunnel renders one tunnel: JSON/unix through the formatter, text as
// the Python CLI's TUNNEL_STANDARD_FIELDS record.
func displayTunnel(w io.Writer, t *transfer.Tunnel) error {
	formatter := output.NewFormatter(viper.GetString("format"), w)
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(t, nil)
	}
	printRecord(w, [][2]string{
		{"ID", t.ID},
		{"Label", t.Label},
		{"State", t.State},
		{"Status", t.Status},
		{"Listener Access Point", t.ListenerAccessPointID},
		{"Listener IP", t.ListenerIPAddress},
		{"Listener Port", formatIntPtr(t.ListenerPort)},
		{"Initiator Access Point", t.InitiatorAccessPointID},
		{"Initiator IP", t.InitiatorIPAddress},
		{"Initiator Port", formatIntPtr(t.InitiatorPort)},
	})
	return nil
}

// streamsTunnelCreateCmd returns "streams tunnel create" with the Python
// CLI's shape: INITIATOR_STREAM_AP_ID LISTENER_STREAM_AP_ID [--label]
// [--lifetime-minutes] [--restartable].
func streamsTunnelCreateCmd() *cobra.Command {
	var (
		label       string
		lifetime    int
		restartable bool
	)
	cmd := &cobra.Command{
		Use:   "create INITIATOR_STREAM_AP_ID LISTENER_STREAM_AP_ID",
		Short: "Create a Globus Streams tunnel",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			create := &transfer.TunnelCreate{
				InitiatorStreamAccessPoint: args[0],
				ListenerStreamAccessPoint:  args[1],
				Label:                      label,
			}
			if cmd.Flags().Changed("lifetime-minutes") {
				create.LifetimeMins = &lifetime
			}
			if restartable {
				create.Restartable = &restartable
			}
			return createTunnel(cmd, create)
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "A friendly label to associate with the tunnel")
	cmd.Flags().IntVar(&lifetime, "lifetime-minutes", 0, "The amount of time in minutes that the tunnel will exist")
	cmd.Flags().BoolVar(&restartable, "restartable", false, "Restart the tunnel in the event of an error")
	return cmd
}

// tunnelCreateCmd returns the legacy flag-based "tunnel create" command.
func tunnelCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Globus Streams tunnel",
		Long: `Create a Globus Streams tunnel connecting a listener stream access
point to an initiator stream access point.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			return createTunnel(cmd, create)
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

func createTunnel(cmd *cobra.Command, create *transfer.TunnelCreate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	tunnel, err := client.CreateTunnel(ctx, create)
	if err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}
	return displayTunnel(cmd.OutOrStdout(), tunnel)
}

// streamsTunnelUpdateCmd returns "streams tunnel update TUNNEL_ID [--label]".
func streamsTunnelUpdateCmd() *cobra.Command {
	var label string
	cmd := &cobra.Command{
		Use:   "update TUNNEL_ID",
		Short: "Update a Globus Streams tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateTunnel(cmd, args[0], &transfer.TunnelUpdate{Label: label}, "update")
		},
	}
	cmd.Flags().StringVar(&label, "label", "", "A friendly label to associate with the tunnel")
	return cmd
}

// tunnelUpdateCmd returns the legacy "tunnel update" command.
func tunnelUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update TUNNEL_ID",
		Short: "Update a Globus Streams tunnel",
		Long:  `Update the mutable fields of a Globus Streams tunnel.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			update := &transfer.TunnelUpdate{
				Label:             tunnelLabel,
				ListenerIPAddress: tunnelListenerIP,
			}
			if cmd.Flags().Changed("listener-port") {
				port := tunnelListenerPort
				update.ListenerPort = &port
			}
			return updateTunnel(cmd, args[0], update, "update")
		},
	}

	cmd.Flags().StringVar(&tunnelLabel, "label", "", "Human-readable label for the tunnel")
	cmd.Flags().IntVar(&tunnelListenerPort, "listener-port", 0, "Listener port")
	cmd.Flags().StringVar(&tunnelListenerIP, "listener-ip", "", "Listener IP address")

	return cmd
}

// tunnelStopCmd returns "tunnel stop": it stops a tunnel but leaves its data
// in the system (PATCH attributes.state=STOPPING), as in the Python CLI.
func tunnelStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop TUNNEL_ID",
		Short: "Stop a Globus Streams tunnel",
		Long:  `Stop a tunnel but leave its data in the system.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateTunnel(cmd, args[0], &transfer.TunnelUpdate{State: tunnelStopState}, "stop")
		},
	}
}

func updateTunnel(cmd *cobra.Command, tunnelID string, update *transfer.TunnelUpdate, verb string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}
	tunnel, err := client.UpdateTunnel(ctx, tunnelID, update)
	if err != nil {
		return fmt.Errorf("failed to %s tunnel: %w", verb, err)
	}
	return displayTunnel(cmd.OutOrStdout(), tunnel)
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

			fmt.Fprintln(cmd.OutOrStdout(), "Tunnel deleted successfully")
			return nil
		},
	}
}

// tunnelEventsCmd returns the tunnel events command.
func tunnelEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events TUNNEL_ID",
		Short: "List events for a Globus Streams tunnel",
		Long:  `Show the events that have occurred on a Globus Streams tunnel.`,
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

			// Python CLI TUNNEL_EVENT_LIST_FIELDS: Code, Error, Details, Time.
			type eventRow struct {
				Code    string
				Error   bool
				Details string
				Time    string
			}
			rows := make([]eventRow, 0, len(resp.Events))
			for _, e := range resp.Events {
				rows = append(rows, eventRow{
					Code:    e.Code,
					Error:   e.IsError,
					Details: e.DetailText,
					Time:    formatTime(e.OccurredAt),
				})
			}
			return formatter.FormatOutput(rows, []string{"Code", "Error", "Details", "Time"})
		},
	}

	cmd.Flags().IntVar(&tunnelLimit, "limit", 25, "Maximum number of events to return")

	return cmd
}

// printRecord prints key/value pairs as an aligned record, like the Python
// CLI's RECORD text mode.
func printRecord(w io.Writer, fields [][2]string) {
	width := 0
	for _, f := range fields {
		if len(f[0]) > width {
			width = len(f[0])
		}
	}
	for _, f := range fields {
		fmt.Fprintf(w, "%-*s %s\n", width+1, f[0]+":", f[1])
	}
}

// formatTime renders t as "YYYY-MM-DD HH:MM:SS" UTC, or "" when nil.
func formatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

// formatIntPtr renders an optional integer, or "" when nil.
func formatIntPtr(n *int) string {
	if n == nil {
		return ""
	}
	return strconv.Itoa(*n)
}
