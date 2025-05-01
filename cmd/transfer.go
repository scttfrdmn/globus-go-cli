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
endpoint management, and transfer task handling.

Available Commands:
  endpoint    Manage Globus Transfer endpoints
  ls          List directory contents on an endpoint
  mkdir       Create a directory on an endpoint
  rm          Remove files and directories from an endpoint
  cp          Transfer files between endpoints
  task        Manage transfer tasks

Examples:
  globus transfer endpoint list               List your endpoints
  globus transfer endpoint search "my data"   Search for endpoints
  globus transfer ls ENDPOINT_ID:/path        List files on an endpoint
  globus transfer mkdir ENDPOINT_ID:/newdir   Create a directory
  globus transfer cp SOURCE_EP:/file DEST_EP:/path   Transfer a file
  globus transfer task show TASK_ID           Check transfer status`,
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