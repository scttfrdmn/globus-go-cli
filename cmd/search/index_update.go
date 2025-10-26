// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"os"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	indexUpdateDisplayName string
	indexUpdateDescription string
	indexUpdateActive      bool
	indexUpdateMonitored   bool
)

// IndexUpdateCmd represents the search index update command
var IndexUpdateCmd = &cobra.Command{
	Use:   "update INDEX_ID",
	Short: "Update a Globus Search index",
	Long: `Update settings for a Globus Search index.

You can update the display name, description, monitoring status,
and active status of an index you own or administer.

Examples:
  # Update display name
  globus search index update INDEX_ID --display-name "New Name"

  # Update description
  globus search index update INDEX_ID --description "Updated description"

  # Enable monitoring
  globus search index update INDEX_ID --monitored

  # Deactivate an index
  globus search index update INDEX_ID --active=false`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexUpdate,
}

func init() {
	IndexUpdateCmd.Flags().StringVar(&indexUpdateDisplayName, "display-name", "", "New display name for the index")
	IndexUpdateCmd.Flags().StringVar(&indexUpdateDescription, "description", "", "New description for the index")
	IndexUpdateCmd.Flags().BoolVar(&indexUpdateActive, "active", true, "Set index active status")
	IndexUpdateCmd.Flags().BoolVar(&indexUpdateMonitored, "monitored", false, "Enable monitoring for the index")
}

func runIndexUpdate(cmd *cobra.Command, args []string) error {
	indexID := args[0]

	// Check if any update flags were provided
	if !cmd.Flags().Changed("display-name") &&
		!cmd.Flags().Changed("description") &&
		!cmd.Flags().Changed("active") &&
		!cmd.Flags().Changed("monitored") {
		return fmt.Errorf("at least one update flag must be provided")
	}

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

	// Load client configuration
	_, err = config.LoadClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create authorizer
	tokenAuthorizer := authorizers.NewStaticTokenAuthorizer(tokenInfo.AccessToken)
	coreAuthorizer := authorizers.ToCore(tokenAuthorizer)

	// Create search client
	searchClient, err := search.NewClient(
		search.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create search client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build update request
	updateRequest := &search.IndexUpdateRequest{}

	if cmd.Flags().Changed("display-name") {
		updateRequest.DisplayName = indexUpdateDisplayName
	}
	if cmd.Flags().Changed("description") {
		updateRequest.Description = indexUpdateDescription
	}
	if cmd.Flags().Changed("active") {
		updateRequest.IsActive = indexUpdateActive
	}
	if cmd.Flags().Changed("monitored") {
		updateRequest.IsMonitored = indexUpdateMonitored
	}

	// Update index
	index, err := searchClient.UpdateIndex(ctx, indexID, updateRequest)
	if err != nil {
		return fmt.Errorf("error updating index: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Index updated successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Index ID:     %s\n", index.ID)
	fmt.Fprintf(os.Stdout, "Display Name: %s\n", index.DisplayName)
	if index.Description != "" {
		fmt.Fprintf(os.Stdout, "Description:  %s\n", index.Description)
	}
	fmt.Fprintf(os.Stdout, "Active:       %t\n", index.IsActive)
	fmt.Fprintf(os.Stdout, "Monitored:    %t\n", index.IsMonitored)
	fmt.Fprintf(os.Stdout, "Updated At:   %s\n", index.UpdatedAt.Format(time.RFC3339))

	return nil
}
