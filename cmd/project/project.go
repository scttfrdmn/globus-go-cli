// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package project

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// Flags for project create/update.
var (
	projectDisplayName  string
	projectContactEmail string
	projectAdmins       []string
	projectAdminGroups  []string
)

// ProjectCmd returns the project command tree for managing Globus Auth
// projects. Every subcommand obtains a manage_projects-scoped auth client via
// getClient (escalating consent on first use).
func ProjectCmd() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Commands for managing Globus Auth projects",
		Long: `Manage Globus Auth projects — the developer-console administrative surface
for grouping registered clients (service accounts / app registrations).

These operations require the auth.globus.org manage_projects scope, which is
requested via a dedicated consent on first use.`,
	}

	projectCmd.AddCommand(
		projectListCmd(),
		projectShowCmd(),
		projectCreateCmd(),
		projectUpdateCmd(),
		projectDeleteCmd(),
		adminCmd(),
	)

	return projectCmd
}

// projectListCmd returns the project list command.
func projectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your Globus Auth projects",
		Long:  `List the Globus Auth projects on which the current user is an administrator.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listProjects(cmd)
		},
	}
}

// projectShowCmd returns the project show command.
func projectShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show PROJECT_ID",
		Short: "Show project details",
		Long:  `Show detailed information about a specific Globus Auth project.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showProject(cmd, args[0])
		},
	}
}

// projectCreateCmd returns the project create command.
func projectCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create a new Globus Auth project.

At least one administrator must be provided via --admin or --admin-group.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProject(cmd)
		},
	}

	cmd.Flags().StringVar(&projectDisplayName, "display-name", "", "Display name for the project (required)")
	cmd.Flags().StringVar(&projectContactEmail, "contact-email", "", "Contact email for the project")
	cmd.Flags().StringSliceVar(&projectAdmins, "admin", nil, "Identity ID of an admin (repeatable)")
	cmd.Flags().StringSliceVar(&projectAdminGroups, "admin-group", nil, "Group ID of an admin group (repeatable)")
	_ = cmd.MarkFlagRequired("display-name")

	return cmd
}

// projectUpdateCmd returns the project update command.
func projectUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update PROJECT_ID",
		Short: "Update a project",
		Long: `Update a Globus Auth project.

Only the flags you provide are changed; unset fields are left as-is.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateProject(cmd, args[0])
		},
	}

	cmd.Flags().StringVar(&projectDisplayName, "display-name", "", "New display name for the project")
	cmd.Flags().StringVar(&projectContactEmail, "contact-email", "", "New contact email for the project")

	return cmd
}

// projectDeleteCmd returns the project delete command.
func projectDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete PROJECT_ID",
		Short: "Delete a project",
		Long:  `Delete a Globus Auth project.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteProject(cmd, args[0])
		},
	}
}

// listProjects lists the current user's projects.
func listProjects(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	projects, err := client.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(projects, nil)
	}

	type projectRow struct {
		ID           string
		DisplayName  string
		ContactEmail string
	}
	rows := make([]projectRow, 0, len(projects))
	for _, p := range projects {
		rows = append(rows, projectRow{
			ID: p.ID, DisplayName: p.DisplayName, ContactEmail: p.ContactEmail,
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "DisplayName", "ContactEmail"})
}

// showProject shows details for a single project.
func showProject(cmd *cobra.Command, projectID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(project, nil)
	}

	fmt.Println("Project Details:")
	fmt.Printf("  ID:            %s\n", project.ID)
	fmt.Printf("  Display Name:  %s\n", project.DisplayName)
	fmt.Printf("  Contact Email: %s\n", project.ContactEmail)
	fmt.Printf("  Admin IDs:     %v\n", project.AdminIDs)
	fmt.Printf("  Created At:    %s\n", project.CreatedAt.Format(time.RFC3339))

	return nil
}

// createProject creates a new project.
func createProject(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	create := &auth.ProjectCreate{
		DisplayName:   projectDisplayName,
		ContactEmail:  projectContactEmail,
		AdminIDs:      projectAdmins,
		AdminGroupIDs: projectAdminGroups,
	}

	var project *auth.Project
	if err := withProjectRetry(ctx, func(client *auth.Client) error {
		p, err := client.CreateProject(ctx, create)
		if err != nil {
			return err
		}
		project = p
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Created project %s (%s)\n", project.ID, project.DisplayName)
	return nil
}

// updateProject updates a project, sending only the flags that were set.
func updateProject(cmd *cobra.Command, projectID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	update := &auth.ProjectUpdate{}
	if cmd.Flags().Changed("display-name") {
		update.DisplayName = projectDisplayName
	}
	if cmd.Flags().Changed("contact-email") {
		update.ContactEmail = projectContactEmail
	}

	if err := withProjectRetry(ctx, func(client *auth.Client) error {
		_, err := client.UpdateProject(ctx, projectID, update)
		return err
	}); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	fmt.Printf("Updated project %s\n", projectID)
	return nil
}

// deleteProject deletes a project.
func deleteProject(cmd *cobra.Command, projectID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := withProjectRetry(ctx, func(client *auth.Client) error {
		return client.DeleteProject(ctx, projectID)
	}); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Printf("Deleted project %s\n", projectID)
	return nil
}
