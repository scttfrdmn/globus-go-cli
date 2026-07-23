// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/scttfrdmn/globus-go-cli/cmd/project"
)

// getProjectCommand returns the `project` command group: management of Globus
// Auth projects, their registered clients (service accounts), and client secret
// credentials — the developer/console administrative surface. The `client` and
// `credential` subgroups are attached here; the `admin` subgroup is attached by
// ProjectCmd itself.
func getProjectCommand() *cobra.Command {
	projectCmd := project.ProjectCmd()
	projectCmd.AddCommand(project.ProjectClientCmd())
	projectCmd.AddCommand(project.CredentialCmd())
	return projectCmd
}
