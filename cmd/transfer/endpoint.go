// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
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
	filterSubscribeId  string
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
	cmd.Flags().StringVar(&filterSubscribeId, "subscription", "", "Filter by subscription ID")
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
	// Get current profile
	profile := viper.GetString("profile")
	
	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration - not used with direct client initialization in v0.9.10
	// We still load it for future use cases
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create a simple static token authorizer for v0.9.10
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	
	// Create a core authorizer adapter for v0.9.10 compatibility
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create transfer client with v0.9.10 compatible options
	transferOptions := []transfer.Option{
		transfer.WithAuthorizer(coreAuthorizer),
	}
	
	transferClient, err := transfer.NewClient(transferOptions...)
	if err != nil {
		return fmt.Errorf("failed to create transfer client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Prepare options for listing endpoints
	options := &transfer.ListEndpointsOptions{
		Limit: limit,
	}

	// Add filters based on flags
	if filterOwner != "" {
		options.FilterOwnerID = filterOwner
	}
	if filterRecentlyUsed {
		options.FilterRecentlyUsed = true
	}
	if filterIsManagedBy != "" {
		options.FilterIsManagedBy = filterIsManagedBy
	}
	if filterOrganization != "" {
		options.FilterOrganization = filterOrganization
	}
	if filterRole != "" {
		options.FilterRole = filterRole
	}
	if filterSubscribeId != "" {
		options.FilterSubscriptionID = filterSubscribeId
	}
	if filterMyTasksOnly {
		options.FilterMyTasksOnly = true
	}
	if searchText != "" {
		options.SearchText = searchText
	}

	// Get the endpoints
	endpoints, err := transferClient.ListEndpoints(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to list endpoints: %w", err)
	}

	// Get output format
	format := viper.GetString("format")
	if format == "" {
		format = "text"
	}

	// Display the results based on format
	switch strings.ToLower(format) {
	case "json":
		// Output as JSON
		jsonData, err := json.MarshalIndent(endpoints, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case "csv":
		// Output as CSV
		fmt.Println("id,display_name,owner_string,activated,gcp_connected")
		for _, endpoint := range endpoints.Data {
			fmt.Printf("%s,%s,%s,%t,%t\n",
				endpoint.ID,
				strings.ReplaceAll(endpoint.DisplayName, ",", " "),
				strings.ReplaceAll(endpoint.OwnerString, ",", " "),
				endpoint.Activated,
				endpoint.GCPConnected,
			)
		}
	default:
		// Output as text table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tName\tOwner\tActivated\tConnected\t")
		fmt.Fprintln(w, "---\t----\t-----\t---------\t---------\t")

		for _, endpoint := range endpoints.Data {
			fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%t\t\n",
				endpoint.ID,
				endpoint.DisplayName,
				endpoint.OwnerString,
				endpoint.Activated,
				endpoint.GCPConnected,
			)
		}
		w.Flush()

		// Display count
		fmt.Printf("\nShowing %d of %d endpoints\n", len(endpoints.Data), endpoints.Length)
	}

	return nil
}

// showEndpoint shows details for a specific endpoint
func showEndpoint(cmd *cobra.Command, endpointID string) error {
	// Get current profile
	profile := viper.GetString("profile")
	
	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration - not used with direct client initialization in v0.9.10
	// We still load it for future use cases
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create a simple static token authorizer for v0.9.10
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	
	// Create a core authorizer adapter for v0.9.10 compatibility
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create transfer client with v0.9.10 compatible options
	transferOptions := []transfer.Option{
		transfer.WithAuthorizer(coreAuthorizer),
	}
	
	transferClient, err := transfer.NewClient(transferOptions...)
	if err != nil {
		return fmt.Errorf("failed to create transfer client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the endpoint
	endpoint, err := transferClient.GetEndpoint(ctx, endpointID)
	if err != nil {
		return fmt.Errorf("failed to get endpoint: %w", err)
	}

	// Get output format
	format := viper.GetString("format")
	if format == "" {
		format = "text"
	}

	// Display the results based on format
	switch strings.ToLower(format) {
	case "json":
		// Output as JSON
		jsonData, err := json.MarshalIndent(endpoint, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	default:
		// Output as text
		fmt.Println("Endpoint Details:")
		fmt.Printf("  ID:             %s\n", endpoint.ID)
		fmt.Printf("  Display Name:   %s\n", endpoint.DisplayName)
		fmt.Printf("  Owner:          %s\n", endpoint.OwnerString)
		fmt.Printf("  Description:    %s\n", endpoint.Description)
		fmt.Printf("  Activated:      %t\n", endpoint.Activated)
		fmt.Printf("  Connected:      %t\n", endpoint.GCPConnected)
		fmt.Printf("  Default Dir:    %s\n", endpoint.DefaultDirectory)
		
		if endpoint.Activated {
			fmt.Printf("  Expires In:     %s\n", formatDuration(endpoint.ExpiresIn))
		}
		
		// Show server information if available
		if len(endpoint.Data) > 0 {
			fmt.Println("\nServer Configuration:")
			for i, server := range endpoint.Data {
				fmt.Printf("  Server #%d:\n", i+1)
				fmt.Printf("    Hostname:       %s\n", server.Hostname)
				fmt.Printf("    Scheme:         %s\n", server.Scheme)
				fmt.Printf("    Port:           %d\n", server.Port)
				fmt.Printf("    Subject:        %s\n", server.Subject)
			}
		}
		
		// Show network use if available
		if endpoint.NetworkUse != nil {
			fmt.Println("\nNetwork Use:")
			fmt.Printf("  Max Concurrency:    %d\n", endpoint.NetworkUse.MaxConcurrency)
			fmt.Printf("  Max Parallelism:    %d\n", endpoint.NetworkUse.MaxParallelism)
			fmt.Printf("  Preferred Concurrency: %d\n", endpoint.NetworkUse.PreferredConcurrency)
			fmt.Printf("  Preferred Parallelism: %d\n", endpoint.NetworkUse.PreferredParallelism)
		}
	}

	return nil
}

// formatDuration formats a duration in seconds as a human-readable string
func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "Expired"
	}
	
	duration := time.Duration(seconds) * time.Second
	
	if duration.Hours() > 24 {
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24
		return fmt.Sprintf("%d days, %d hours", days, hours)
	}
	
	if duration.Hours() >= 1 {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	}
	
	if duration.Minutes() >= 1 {
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		return fmt.Sprintf("%d minutes, %d seconds", minutes, seconds)
	}
	
	return fmt.Sprintf("%d seconds", seconds)
}