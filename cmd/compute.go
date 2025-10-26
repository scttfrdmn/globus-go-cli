// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/scttfrdmn/globus-go-cli/cmd/compute"
	"github.com/spf13/cobra"
)

func getComputeCommand() *cobra.Command {
	computeCmd := &cobra.Command{
		Use:   "compute",
		Short: "Commands for Globus Compute",
		Long: `Commands for interacting with the Globus Compute service.

Globus Compute is a distributed Function as a Service (FaaS) platform that
enables flexible, scalable, and high-performance remote function execution.

You can:
- List and view compute endpoints
- Register and manage functions
- Execute functions remotely and monitor tasks
- Manage dependencies, containers, and environments

Note: This Go CLI provides Compute support not available in the Python CLI!

Examples:
  # List available endpoints
  globus compute endpoint list

  # Register a function
  globus compute function register --name "my_func" --file function.py

  # Run a function
  globus compute task run FUNCTION_ID ENDPOINT_ID --input input.json

  # Check task status
  globus compute task show TASK_ID`,
	}

	// Add subcommands
	computeCmd.AddCommand(compute.GetEndpointCmd())
	computeCmd.AddCommand(compute.GetFunctionCmd())
	computeCmd.AddCommand(compute.GetTaskCmd())

	return computeCmd
}
