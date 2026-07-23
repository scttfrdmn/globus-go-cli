// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// adminCmd returns the admin subgroup for managing a project's administrators.
func adminCmd() *cobra.Command {
	admin := &cobra.Command{
		Use:   "admin",
		Short: "Commands for managing project administrators",
		Long:  `List, add, and remove administrators on a Globus Auth project.`,
	}

	admin.AddCommand(
		adminListCmd(),
		adminAddCmd(),
		adminRemoveCmd(),
	)

	return admin
}

// adminListCmd returns the admin list command.
func adminListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list PROJECT_ID",
		Short: "List a project's administrators",
		Long:  `List the administrator identities and admin groups on a Globus Auth project.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAdmins(cmd, args[0])
		},
	}
}

// adminAddCmd returns the admin add command.
func adminAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add PROJECT_ID IDENTITY_OR_USERNAME",
		Short: "Add an administrator to a project",
		Long: `Add an administrator to a Globus Auth project.

The second argument may be an identity ID (UUID) or a username; a username is
resolved to its identity ID via Globus Auth (provisioning if necessary).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addAdmin(cmd, args[0], args[1])
		},
	}
}

// adminRemoveCmd returns the admin remove command.
func adminRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove PROJECT_ID IDENTITY_ID",
		Short: "Remove an administrator from a project",
		Long:  `Remove an administrator identity from a Globus Auth project.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeAdmin(cmd, args[0], args[1])
		},
	}
}

// listAdmins displays a project's administrator identities and groups.
func listAdmins(cmd *cobra.Command, projectID string) error {
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

	// Collect admin identity IDs from both the flat list and the expanded
	// admins object (deduped).
	identities := dedupeStrings(project.AdminIDs)
	groups := append([]string(nil), project.AdminGroupIDs...)
	if project.Admins != nil {
		identities = dedupeStrings(append(identities, project.Admins.Identities...))
		groups = dedupeStrings(append(groups, project.Admins.Groups...))
	}
	groups = dedupeStrings(groups)

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(project, nil)
	}

	type adminRow struct {
		IdentityID string
	}
	rows := make([]adminRow, 0, len(identities))
	for _, id := range identities {
		rows = append(rows, adminRow{IdentityID: id})
	}
	if err := formatter.FormatOutput(rows, []string{"IdentityID"}); err != nil {
		return err
	}

	if len(groups) > 0 {
		fmt.Println("\nAdmin Groups:")
		for _, g := range groups {
			fmt.Printf("  %s\n", g)
		}
	}

	return nil
}

// addAdmin resolves the given identity-or-username to an identity ID and adds
// it to the project's admin list (preserving other fields).
func addAdmin(cmd *cobra.Command, projectID, identityOrUsername string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	identityID, err := resolveIdentityID(ctx, client, identityOrUsername)
	if err != nil {
		return err
	}

	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	merged := dedupeStrings(append(project.AdminIDs, identityID))

	update := &auth.ProjectUpdate{
		DisplayName:  project.DisplayName,
		ContactEmail: project.ContactEmail,
		AdminIDs:     merged,
	}
	if _, err := client.UpdateProject(ctx, projectID, update); err != nil {
		return fmt.Errorf("failed to update project admins: %w", err)
	}

	fmt.Printf("Added admin %s to project %s\n", identityID, projectID)
	return nil
}

// removeAdmin removes an identity ID from the project's admin list (preserving
// other fields).
func removeAdmin(cmd *cobra.Command, projectID, identityID string) error {
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

	reduced := make([]string, 0, len(project.AdminIDs))
	for _, id := range project.AdminIDs {
		if id != identityID {
			reduced = append(reduced, id)
		}
	}

	update := &auth.ProjectUpdate{
		DisplayName:  project.DisplayName,
		ContactEmail: project.ContactEmail,
		AdminIDs:     reduced,
	}
	if _, err := client.UpdateProject(ctx, projectID, update); err != nil {
		return fmt.Errorf("failed to update project admins: %w", err)
	}

	fmt.Printf("Removed admin %s from project %s\n", identityID, projectID)
	return nil
}

// resolveIdentityID returns an identity ID for the given argument. If the
// argument looks like an ID (no "@"), it is used as-is; otherwise it is treated
// as a username and looked up (provisioning if necessary).
func resolveIdentityID(ctx context.Context, client *auth.Client, arg string) (string, error) {
	if !strings.Contains(arg, "@") {
		return arg, nil
	}

	identities, err := client.GetIdentities(ctx, &auth.GetIdentitiesOptions{
		Usernames: []string{arg},
		Provision: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to resolve identity %q: %w", arg, err)
	}
	if len(identities) == 0 {
		return "", fmt.Errorf("no identity found for username %q", arg)
	}
	return identities[0].ID, nil
}

// dedupeStrings returns the input with duplicate and empty entries removed,
// preserving order.
func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
