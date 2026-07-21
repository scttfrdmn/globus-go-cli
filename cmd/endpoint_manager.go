// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-cli/cmd/transfer"
)

// getEndpointManagerCommand returns the `endpoint-manager` admin command group.
func getEndpointManagerCommand() *cobra.Command {
	return transfer.EndpointManagerCmd()
}
