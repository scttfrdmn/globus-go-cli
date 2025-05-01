// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"
	
	"github.com/scttfrdmn/globus-go-cli/cmd/transfer"
)

// getTransferCommand returns the transfer command
func getTransferCommand() *cobra.Command {
	// transferCmd represents the transfer command
	transferCmd := &cobra.Command{
		Use:   "transfer",
		Short: "Commands for Globus Transfer",
		Long: `Commands for working with Globus Transfer including file operations,
endpoint management, and transfer task handling.`,
	}

	// Import transfer commands
	transferCmd.AddCommand(
		transfer.EndpointCmd(),
		transfer.LsCmd(),
		transfer.MkdirCmd(),
		transfer.CpCmd(),
		transfer.RmCmd(),
		transfer.TaskCmd(),
	)

	return transferCmd
}