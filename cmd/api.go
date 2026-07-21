// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-cli/cmd/api"
)

// getAPICommand returns the `api` raw-passthrough command group.
func getAPICommand() *cobra.Command {
	return api.APICmd()
}
