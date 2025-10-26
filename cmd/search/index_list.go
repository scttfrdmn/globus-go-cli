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

var (
	indexListLimit  int
	indexListOffset int
)

// IndexListCmd represents the search index list command
var IndexListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Globus Search indices",
	Long: `List all Globus Search indices where you have some permissions.

This shows indices you own, administer, or have access to.

Examples:
  # List all your indices
  globus search index list

  # Limit results
  globus search index list --limit 20

  # JSON output for scripting
  globus search index list --format json`,
	RunE: runIndexList,
}

func init() {
	IndexListCmd.Flags().IntVar(&indexListLimit, "limit", 100, "Maximum number of indices to return")
	IndexListCmd.Flags().IntVar(&indexListOffset, "offset", 0, "Offset for pagination")
}

func runIndexList(cmd *cobra.Command, args []string) error {
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

	// List indices
	options := &search.ListIndexesOptions{
		Limit:  indexListLimit,
		Offset: indexListOffset,
	}

	indexList, err := searchClient.ListIndexes(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing indices: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(indexList.Indexes) == 0 {
			fmt.Println("No indices found.")
			return nil
		}

		fmt.Printf("%-36s  %-40s  %-10s  %-10s\n", "Index ID", "Display Name", "Active", "Public")
		fmt.Printf("%s  %s  %s  %s\n",
			"------------------------------------",
			"----------------------------------------",
			"----------",
			"----------")

		for _, index := range indexList.Indexes {
			displayName := index.DisplayName
			if len(displayName) > 40 {
				displayName = displayName[:37] + "..."
			}

			active := "No"
			if index.IsActive {
				active = "Yes"
			}

			public := "No"
			if index.IsPublic {
				public = "Yes"
			}

			fmt.Printf("%-36s  %-40s  %-10s  %-10s\n",
				index.ID,
				displayName,
				active,
				public)
		}

		fmt.Printf("\nTotal: %d index(es)\n", len(indexList.Indexes))

		if indexList.HasMore {
			nextOffset := indexListOffset + indexListLimit
			fmt.Printf("More indices available. Use --offset %d to see next page.\n", nextOffset)
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"ID", "DisplayName", "Description", "IsActive", "IsPublic", "CreatedBy"}
		if err := formatter.FormatOutput(indexList.Indexes, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
