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
	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// IndexShowCmd represents the search index show command
var IndexShowCmd = &cobra.Command{
	Use:   "show INDEX_ID",
	Short: "Show details for a Globus Search index",
	Long: `Display detailed information about a Globus Search index.

This shows the index configuration, status, and metadata.

Examples:
  # Show index details
  globus search index show INDEX_ID

  # Show with JSON output
  globus search index show INDEX_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexShow,
}

func runIndexShow(cmd *cobra.Command, args []string) error {
	indexID := args[0]

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

	// Get index details
	index, err := searchClient.GetIndex(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error getting index: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Index Information\n")
		fmt.Printf("=================\n\n")
		fmt.Printf("Index ID:     %s\n", index.ID)
		fmt.Printf("Display Name: %s\n", index.DisplayName)
		if index.Description != "" {
			fmt.Printf("Description:  %s\n", index.Description)
		}
		fmt.Printf("Active:       %t\n", index.IsActive)
		fmt.Printf("Public:       %t\n", index.IsPublic)
		fmt.Printf("Monitored:    %t\n", index.IsMonitored)
		if index.MonitoringFrequency > 0 {
			fmt.Printf("Monitoring Frequency: %d minutes\n", index.MonitoringFrequency)
		}
		fmt.Printf("Max Size:     %d MB\n", index.MaxSize)

		fmt.Printf("\nMetadata\n")
		fmt.Printf("--------\n")
		fmt.Printf("Created By:   %s\n", index.CreatedBy)
		fmt.Printf("Created At:   %s\n", index.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated At:   %s\n", index.UpdatedAt.Format(time.RFC3339))
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "DisplayName", "Description", "IsActive", "IsPublic", "CreatedBy"}
		if err := formatter.FormatOutput(index, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
