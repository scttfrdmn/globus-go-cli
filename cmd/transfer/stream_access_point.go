// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"fmt"

	"github.com/spf13/cobra"
)

// StreamAccessPointCmd returns the stream-access-point subcommand group.
//
// Globus Streams (tunnels and stream access points) is a Python SDK v4.3.0+
// feature and is not part of the frozen v3 SDK this CLI builds against, so these
// commands are not available.
func StreamAccessPointCmd() *cobra.Command {
	sapCmd := &cobra.Command{
		Use:   "stream-access-point",
		Short: "Commands for Globus Streams access points (unavailable)",
		Long: `Commands for working with Globus Streams access points.

NOTE: Globus Streams is a newer API not included in the SDK version this CLI
builds against, so these commands are not available.`,
	}

	sapCmd.AddCommand(streamAccessPointShowCmd())
	return sapCmd
}

// streamAccessPointShowCmd returns the stream-access-point show command.
func streamAccessPointShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ACCESS_POINT_ID",
		Short: "Show details of a Globus Streams access point (unavailable)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("this SDK version does not support Globus Streams")
		},
	}
}
