// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

// EndpointCmd returns the endpoint command
func EndpointCmd() *cobra.Command {
	// endpointCmd represents the endpoint command
	endpointCmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Commands for managing Globus endpoints",
		Long: `Commands for managing Globus Transfer endpoints including listing,
searching, and displaying endpoint details.`,
	}

	// Add endpoint subcommands
	endpointCmd.AddCommand(
		endpointListCmd(),
		endpointShowCmd(),
		endpointSearchCmd(),
		endpointUpdateCmd(),
		endpointDeleteCmd(),
		endpointRoleCmd(),
		endpointPermissionCmd(),
		endpointSetSubscriptionIDCmd(),
		endpointMySharedEndpointListCmd(),
	)

	return endpointCmd
}

// Filter options for endpoint listing
var (
	filterOwner        string
	filterRecentlyUsed bool
	filterIsManagedBy  string
	filterOrganization string
	filterRole         string
	filterSubscribeID  string
	filterMyTasksOnly  bool
	searchText         string
	limit              int
)

// endpointListCmd returns the endpoint list command
func endpointListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Globus endpoints",
		Long: `List Globus Transfer endpoints visible to the current user.

This command lists endpoints that the current user has access to,
with filtering options to narrow down the results.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listEndpoints(cmd)
		},
	}

	// Add flags for filtering
	cmd.Flags().StringVar(&filterOwner, "owner", "", "Filter by owner ID")
	cmd.Flags().BoolVar(&filterRecentlyUsed, "recently-used", false, "Show only recently used endpoints")
	cmd.Flags().StringVar(&filterIsManagedBy, "managed-by", "", "Filter by managing entity")
	cmd.Flags().StringVar(&filterOrganization, "organization", "", "Filter by organization")
	cmd.Flags().StringVar(&filterRole, "role", "", "Filter by role (manager, administrator, etc.)")
	cmd.Flags().StringVar(&filterSubscribeID, "subscription", "", "Filter by subscription ID")
	cmd.Flags().BoolVar(&filterMyTasksOnly, "my-tasks", false, "Show only endpoints with my tasks")
	cmd.Flags().StringVar(&searchText, "search", "", "Search text to filter endpoints")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of endpoints to return")

	return cmd
}

// endpointShowCmd returns the endpoint show command
func endpointShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show ENDPOINT_ID",
		Short: "Show endpoint details",
		Long: `Show detailed information about a specific Globus endpoint.

This command displays all available details about the specified endpoint,
including server configuration, access rules, and more.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showEndpoint(cmd, args[0])
		},
	}
}

// endpointSearchCmd returns the endpoint search command
func endpointSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search SEARCH_TEXT",
		Short: "Search for Globus endpoints",
		Long: `Search for Globus endpoints by name, description, or other attributes.

This command performs a search across all endpoints visible to the current user,
returning matches based on the provided search text.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			searchText = args[0]
			return listEndpoints(cmd)
		},
	}

	// Add flags for filtering
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of endpoints to return")

	return cmd
}

// listEndpoints lists Globus endpoints
func listEndpoints(cmd *cobra.Command) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Prepare options for endpoint search.
	options := &transfer.EndpointSearchOptions{
		Limit: limit,
	}

	if filterOwner != "" {
		options.FilterOwnerID = filterOwner
	}

	// Filter scope handling.
	if filterRecentlyUsed {
		options.FilterScope = "recently-used"
	} else if filterMyTasksOnly {
		options.FilterScope = "in-use"
	}

	// Search text for full text search.
	if searchText != "" {
		options.FilterFulltext = searchText
	}

	// Get the endpoints
	endpoints, err := transferClient.EndpointSearch(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to list endpoints: %w", err)
	}

	// Route all formats through the shared formatter so -F (text/json/unix) and
	// --jmespath/--jq work uniformly. For JSON/JMESPath, emit the raw endpoint
	// documents; for text/unix, a projected row set.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())

	if formatter.Format == output.FormatJSON {
		// Emit the enveloped service document ({"DATA_TYPE","DATA":[...],...}),
		// matching the Python CLI's JSON output shape.
		return formatter.FormatOutput(endpoints, nil)
	}

	type endpointRow struct {
		ID        string
		Name      string
		Owner     string
		Activated bool
		Public    bool
	}
	rows := make([]endpointRow, 0, len(endpoints.Data))
	for _, e := range endpoints.Data {
		rows = append(rows, endpointRow{
			ID: e.ID, Name: e.DisplayName, Owner: e.Owner,
			Activated: e.Activated, Public: e.Public,
		})
	}
	return formatter.FormatOutput(rows, []string{"ID", "Name", "Owner", "Activated", "Public"})
}

// showEndpoint shows details for a specific endpoint
func showEndpoint(cmd *cobra.Command, endpointID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get the endpoint
	endpoint, err := transferClient.GetEndpoint(ctx, endpointID)
	if err != nil {
		return fmt.Errorf("failed to get endpoint: %w", err)
	}

	// For json/unix or a --jmespath/--jq expression, route through the shared
	// formatter (emitting the raw endpoint document). Otherwise render the text
	// detail view below.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(endpoint, nil)
	}

	{
		// Output as text
		fmt.Println("Endpoint Details:")
		fmt.Printf("  ID:             %s\n", endpoint.ID)
		fmt.Printf("  Display Name:   %s\n", endpoint.DisplayName)
		fmt.Printf("  Owner:          %s\n", endpoint.Owner)
		fmt.Printf("  Description:    %s\n", endpoint.Description)
		fmt.Printf("  Activated:      %t\n", endpoint.Activated)
		fmt.Printf("  Public:         %t\n", endpoint.Public)

		// Organization and department if available
		if endpoint.Organization != "" {
			fmt.Printf("  Organization:   %s\n", endpoint.Organization)
		}
		if endpoint.Department != "" {
			fmt.Printf("  Department:     %s\n", endpoint.Department)
		}

		if endpoint.SubscriptionID != "" {
			fmt.Printf("  Subscription:   %s\n", endpoint.SubscriptionID)
		}
	}

	return nil
}
