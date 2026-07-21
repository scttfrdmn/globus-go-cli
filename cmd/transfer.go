// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-cli/cmd/transfer"
)

// addTransferCommands wires the transfer commands directly onto the root
// command as flat top-level commands, matching the Python Globus CLI (globus
// ls, globus mkdir, globus rm, globus transfer, globus task ..., globus
// endpoint ...).
func addTransferCommands(root *cobra.Command) {
	root.AddCommand(
		transfer.LsCmd(),
		transfer.MkdirCmd(),
		transfer.RmCmd(),
		transfer.RenameCmd(),
		transfer.StatCmd(),
		transfer.DeleteCmd(),
		transfer.CpCmd(), // exposed as the top-level `transfer` verb
		transfer.TaskCmd(),
		transfer.EndpointCmd(),
		transfer.BookmarkCmd(),
		transfer.TunnelCmd(),
		transfer.StreamAccessPointCmd(),
	)
}
