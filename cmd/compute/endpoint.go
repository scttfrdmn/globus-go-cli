// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"github.com/spf13/cobra"
)

// GetEndpointCmd returns the endpoint subcommand for compute
func GetEndpointCmd() *cobra.Command {
	endpointCmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage Globus Compute endpoints",
		Long: `Commands for managing Globus Compute endpoints.

Endpoints are execution environments where your functions run. They can be
local machines, clusters, or cloud resources configured with Globus Compute.`,
	}

	// Add endpoint subcommands
	endpointCmd.AddCommand(EndpointListCmd)
	endpointCmd.AddCommand(EndpointShowCmd)

	return endpointCmd
}
