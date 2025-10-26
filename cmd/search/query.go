// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"context"
	"encoding/json"
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
	queryString string
	queryLimit  int
	queryOffset int
	queryAdvanced bool
)

// QueryCmd represents the search query command
var QueryCmd = &cobra.Command{
	Use:   "query INDEX_ID",
	Short: "Query a Globus Search index",
	Long: `Query a Globus Search index by ID using a search query string.

The query command searches for documents in a Search index and returns
matching results. You can use simple keyword searches or advanced query
syntax for more complex searches.

Examples:
  # Simple keyword search
  globus search query INDEX_ID --query "climate data"

  # Limit results
  globus search query INDEX_ID --query "research" --limit 10

  # Advanced query with offset
  globus search query INDEX_ID --query "subject:biology" --limit 20 --offset 40

  # JSON output for scripting
  globus search query INDEX_ID --query "data" --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runSearchQuery,
}

func init() {
	QueryCmd.Flags().StringVarP(&queryString, "query", "q", "", "Search query string (required)")
	QueryCmd.Flags().IntVar(&queryLimit, "limit", 10, "Maximum number of results to return")
	QueryCmd.Flags().IntVar(&queryOffset, "offset", 0, "Offset for pagination")
	QueryCmd.Flags().BoolVar(&queryAdvanced, "advanced", false, "Use advanced query syntax")

	QueryCmd.MarkFlagRequired("query")
}

func runSearchQuery(cmd *cobra.Command, args []string) error {
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

	// Build search request
	searchRequest := &search.SearchRequest{
		IndexID: indexID,
		Query:   queryString,
		Options: &search.SearchOptions{
			Limit:  queryLimit,
			Offset: queryOffset,
		},
	}

	// Execute search
	response, err := searchClient.Search(ctx, searchRequest)
	if err != nil {
		return fmt.Errorf("error executing search: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		if len(response.Results) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		fmt.Printf("Search Results (showing %d of %d total)\n", len(response.Results), response.Total)
		fmt.Printf("========================================\n\n")

		for i, result := range response.Results {
			fmt.Printf("Result %d:\n", i+1)
			fmt.Printf("  Subject: %s\n", result.Subject)
			fmt.Printf("  Score:   %.4f\n", result.Score)

			// Display content
			if result.Content != nil {
				contentJSON, _ := json.MarshalIndent(result.Content, "    ", "  ")
				fmt.Printf("  Content:\n    %s\n", string(contentJSON))
			}

			// Display highlights
			if len(result.Highlight) > 0 {
				fmt.Printf("  Highlights:\n")
				for field, highlights := range result.Highlight {
					fmt.Printf("    %s: %v\n", field, highlights)
				}
			}
			fmt.Println()
		}

		if response.HasMore {
			nextOffset := queryOffset + queryLimit
			fmt.Printf("More results available. Use --offset %d to see next page.\n", nextOffset)
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"Subject", "Content", "Score"}
		if err := formatter.FormatOutput(response.Results, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
