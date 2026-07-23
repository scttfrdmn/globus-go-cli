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

// Flag-backed options for the client subcommands. Prefixed to avoid collisions
// with sibling files in the project package.
var (
	clientCreateProject     string
	clientCreateName        string
	clientCreatePublic      bool
	clientCreateVisibility  string
	clientCreateRedirectURI []string
	clientCreateType        string

	clientUpdateName       string
	clientUpdateVisibility string

	clientRedirectURIs []string

	clientPrivacyPolicy      string
	clientTermsAndConditions string
)

// ProjectClientCmd returns the `client` command tree for managing the Globus
// Auth clients (app / service-account registrations) within a project.
func ProjectClientCmd() *cobra.Command {
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Commands for managing Globus Auth clients",
		Long: `Manage the Globus Auth clients (application and service-account
registrations) belonging to your projects.

These operations require the auth.globus.org manage_projects scope.`,
	}

	clientCmd.AddCommand(
		pcListCmd(),
		pcShowCmd(),
		pcCreateCmd(),
		pcUpdateCmd(),
		pcDeleteCmd(),
		pcUpdateRedirectURIsCmd(),
		pcUpdateMetadataCmd(),
	)

	return clientCmd
}

// pcListCmd returns the `client list` command.
func pcListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list PROJECT_ID",
		Short: "List clients in a project",
		Long: `List the Globus Auth clients belonging to the given project.

The Auth API returns all clients across your projects; this command filters
them to the requested project ID.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcList(cmd, args[0])
		},
	}
}

// pcShowCmd returns the `client show` command.
func pcShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show CLIENT_ID",
		Short: "Show client details",
		Long:  `Show detailed information about a single Globus Auth client.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcShow(cmd, args[0])
		},
	}
}

// pcCreateCmd returns the `client create` command.
func pcCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a client",
		Long:  `Create a new Globus Auth client in a project.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcCreate(cmd)
		},
	}

	cmd.Flags().StringVar(&clientCreateProject, "project", "", "Project ID to create the client in (required)")
	cmd.Flags().StringVar(&clientCreateName, "name", "", "Display name for the client (required)")
	cmd.Flags().BoolVar(&clientCreatePublic, "public", false, "Register the client as a public (native app) client")
	cmd.Flags().StringVar(&clientCreateVisibility, "visibility", "", "Client visibility (public or private)")
	cmd.Flags().StringSliceVar(&clientCreateRedirectURI, "redirect-uri", nil, "Redirect URI (repeatable)")
	cmd.Flags().StringVar(&clientCreateType, "client-type", "", "Client type")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// pcUpdateCmd returns the `client update` command.
func pcUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update CLIENT_ID",
		Short: "Update a client",
		Long:  `Update the mutable fields (name, visibility) of a Globus Auth client.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcUpdate(cmd, args[0])
		},
	}

	cmd.Flags().StringVar(&clientUpdateName, "name", "", "New display name for the client")
	cmd.Flags().StringVar(&clientUpdateVisibility, "visibility", "", "New client visibility (public or private)")

	return cmd
}

// pcDeleteCmd returns the `client delete` command.
func pcDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete CLIENT_ID",
		Short: "Delete a client",
		Long:  `Delete a Globus Auth client.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcDelete(cmd, args[0])
		},
	}
}

// pcUpdateRedirectURIsCmd returns the `client update-redirect-uris` command.
func pcUpdateRedirectURIsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-redirect-uris CLIENT_ID",
		Short: "Update a client's redirect URIs",
		Long:  `Replace the set of redirect URIs registered on a Globus Auth client.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcUpdateRedirectURIs(cmd, args[0])
		},
	}

	cmd.Flags().StringSliceVar(&clientRedirectURIs, "redirect-uri", nil, "Redirect URI (repeatable, required)")
	_ = cmd.MarkFlagRequired("redirect-uri")

	return cmd
}

// pcUpdateMetadataCmd returns the `client update-metadata` command.
func pcUpdateMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-metadata CLIENT_ID",
		Short: "Update a client's metadata links",
		Long:  `Update the privacy policy and terms-and-conditions links on a Globus Auth client.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcUpdateMetadata(cmd, args[0])
		},
	}

	cmd.Flags().StringVar(&clientPrivacyPolicy, "privacy-policy", "", "URL of the client's privacy policy")
	cmd.Flags().StringVar(&clientTermsAndConditions, "terms-and-conditions", "", "URL of the client's terms and conditions")

	return cmd
}

// pcList lists the clients belonging to a project.
func pcList(cmd *cobra.Command, projectID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	all, err := client.GetClients(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	// GetClients returns every client across the caller's projects; filter to
	// the requested project (there is no per-project list method).
	filtered := make([]auth.AuthClientInfo, 0, len(all))
	for _, c := range all {
		if c.Project == projectID {
			filtered = append(filtered, c)
		}
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(filtered, nil)
	}

	type clientRow struct {
		ID           string
		Name         string
		ClientType   string
		PublicClient bool
	}
	rows := make([]clientRow, 0, len(filtered))
	for _, c := range filtered {
		rows = append(rows, clientRow{
			ID: c.ID, Name: c.Name, ClientType: c.ClientType, PublicClient: c.PublicClient,
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "Name", "ClientType", "PublicClient"})
}

// pcShow shows details for a single client.
func pcShow(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	info, err := client.GetClient(ctx, &auth.GetClientOptions{ClientID: clientID})
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(info, nil)
	}

	fmt.Println("Client Details:")
	fmt.Printf("  ID:            %s\n", info.ID)
	fmt.Printf("  Name:          %s\n", info.Name)
	fmt.Printf("  Project:       %s\n", info.Project)
	fmt.Printf("  Client Type:   %s\n", info.ClientType)
	fmt.Printf("  Visibility:    %s\n", info.Visibility)
	fmt.Printf("  Redirect URIs: %v\n", info.RedirectURIs)
	fmt.Printf("  Public Client: %t\n", info.PublicClient)

	return nil
}

// pcCreate creates a new client.
func pcCreate(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	req := &auth.ClientCreate{
		Name:         clientCreateName,
		Project:      clientCreateProject,
		Visibility:   clientCreateVisibility,
		RedirectURIs: clientCreateRedirectURI,
		ClientType:   clientCreateType,
	}
	// Only send public_client when the flag was explicitly set.
	if cmd.Flags().Changed("public") {
		public := clientCreatePublic
		req.PublicClient = &public
	}

	created, err := client.CreateClient(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	fmt.Printf("Created client %s (%s)\n", created.ID, created.Name)
	return nil
}

// pcUpdate updates the mutable fields of a client.
func pcUpdate(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	update := &auth.ClientUpdate{}
	if cmd.Flags().Changed("name") {
		update.Name = clientUpdateName
	}
	if cmd.Flags().Changed("visibility") {
		update.Visibility = clientUpdateVisibility
	}

	if _, err := client.UpdateClient(ctx, clientID, update); err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	fmt.Printf("Updated client %s\n", clientID)
	return nil
}

// pcDelete deletes a client.
func pcDelete(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	if err := client.DeleteClient(ctx, clientID); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	fmt.Printf("Deleted client %s\n", clientID)
	return nil
}

// pcUpdateRedirectURIs replaces a client's redirect URIs.
func pcUpdateRedirectURIs(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	update := &auth.ClientUpdate{RedirectURIs: clientRedirectURIs}
	if _, err := client.UpdateClient(ctx, clientID, update); err != nil {
		return fmt.Errorf("failed to update redirect URIs: %w", err)
	}

	fmt.Printf("Updated redirect URIs for client %s\n", clientID)
	return nil
}

// pcUpdateMetadata updates a client's terms/privacy metadata links.
func pcUpdateMetadata(cmd *cobra.Command, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	links := &auth.ClientLinksInput{}
	if cmd.Flags().Changed("privacy-policy") {
		links.PrivacyPolicy = clientPrivacyPolicy
	}
	if cmd.Flags().Changed("terms-and-conditions") {
		links.TermsAndConditions = clientTermsAndConditions
	}

	update := &auth.ClientUpdate{Links: links}
	if _, err := client.UpdateClient(ctx, clientID, update); err != nil {
		return fmt.Errorf("failed to update client metadata: %w", err)
	}

	fmt.Printf("Updated metadata for client %s\n", clientID)
	return nil
}
