// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// readJSONArg resolves a [JSON_FILE|JSON|file:JSON_FILE] argument (the form the
// Python CLI accepts for query documents) into raw JSON bytes. A "file:" prefix
// or an existing path is read from disk; otherwise the value is treated as
// inline JSON.
func readJSONArg(arg string) ([]byte, error) {
	if strings.HasPrefix(arg, "file:") {
		return os.ReadFile(strings.TrimPrefix(arg, "file:"))
	}
	trimmed := strings.TrimSpace(arg)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return []byte(arg), nil
	}
	if data, err := os.ReadFile(arg); err == nil {
		return data, nil
	}
	// Fall back to treating the value as inline JSON.
	return []byte(arg), nil
}

var (
	queryString   string
	queryDocument string
	queryLimit    int
	queryOffset   int
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
	QueryCmd.Flags().StringVarP(&queryString, "query", "q", "", "Search query string")
	QueryCmd.Flags().StringVar(&queryDocument, "query-document", "", "A complete query document (inline JSON, a path, or file:PATH). At least one of -q or --query-document is required")
	QueryCmd.Flags().IntVar(&queryLimit, "limit", 10, "Maximum number of results to return")
	QueryCmd.Flags().IntVar(&queryOffset, "offset", 0, "Offset for pagination")
	QueryCmd.Flags().BoolVar(&queryAdvanced, "advanced", false, "Use advanced query syntax")
}

func runSearchQuery(cmd *cobra.Command, args []string) error {
	indexID := args[0]

	if queryString == "" && queryDocument == "" {
		return fmt.Errorf("at least one of -q/--query or --query-document must be provided")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Search client authorized for the current profile.
	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	var response *search.SearchResults

	if queryDocument != "" {
		// A complete query document is posted via POST /index/{id}/search.
		// Command-line options, when present, override the document's fields.
		docData, derr := readJSONArg(queryDocument)
		if derr != nil {
			return fmt.Errorf("failed to read query document: %w", derr)
		}
		var q search.SearchQuery
		if uerr := json.Unmarshal(docData, &q); uerr != nil {
			return fmt.Errorf("failed to parse query document JSON: %w", uerr)
		}
		if cmd.Flags().Changed("query") {
			q.Q = queryString
		}
		if cmd.Flags().Changed("limit") {
			q.Limit = queryLimit
		}
		if cmd.Flags().Changed("offset") {
			q.Offset = queryOffset
		}
		if cmd.Flags().Changed("advanced") {
			q.AdvancedQuery = queryAdvanced
		}
		response, err = searchClient.Search(ctx, indexID, &q)
	} else {
		// A simple query string is modeled as GET /v1/index/{id}/search with
		// q/offset/limit/advanced query params (SearchGet).
		response, err = searchClient.SearchGet(ctx, indexID, &search.SearchGetOptions{
			Q:        queryString,
			Offset:   queryOffset,
			Limit:    queryLimit,
			Advanced: queryAdvanced,
		})
	}
	if err != nil {
		return fmt.Errorf("error executing search: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		if len(response.GMeta) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		fmt.Printf("Search Results (showing %d of %d total)\n", len(response.GMeta), response.Total)
		fmt.Printf("========================================\n\n")

		for i, result := range response.GMeta {
			fmt.Printf("Result %d:\n", i+1)
			fmt.Printf("  Subject: %s\n", result.Subject)

			// Display content
			if len(result.Content) > 0 {
				contentJSON, _ := json.MarshalIndent(result.Content, "    ", "  ")
				fmt.Printf("  Content:\n    %s\n", string(contentJSON))
			}
			fmt.Println()
		}

		if response.HasNextPage {
			nextOffset := queryOffset + queryLimit
			fmt.Printf("More results available. Use --offset %d to see next page.\n", nextOffset)
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"Subject", "Content", "Entries"}
		if err := formatter.FormatOutput(response.GMeta, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
