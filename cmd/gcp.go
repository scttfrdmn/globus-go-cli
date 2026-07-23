// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-cli/cmd/gcp"
)

// getGCPCommand returns the `gcp` command group (Globus Connect Personal
// endpoint/collection management via the Globus service API).
func getGCPCommand() *cobra.Command {
	return gcp.GCPCmd()
}
