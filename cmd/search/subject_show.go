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

// SubjectShowCmd represents the search subject show command
var SubjectShowCmd = &cobra.Command{
	Use:   "show INDEX_ID SUBJECT",
	Short: "Show data for a subject in a Globus Search index",
	Long: `Display all entries for a specific subject in a Search index.

Only entries visible to you will be shown, based on the visible_to
access control list. If no entries are visible, an error is returned.

Examples:
  # Show subject details
  globus search subject show INDEX_ID my-document-id

  # Show with JSON output
  globus search subject show INDEX_ID doc123 --format json`,
	Args: cobra.ExactArgs(2),
	RunE: runSubjectShow,
}

func runSubjectShow(cmd *cobra.Command, args []string) error {
	indexID := args[0]
	subject := args[1]

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

	// Search for the specific subject
	searchRequest := &search.SearchRequest{
		IndexID: indexID,
		Query:   fmt.Sprintf("subject:%s", subject),
	}

	response, err := searchClient.Search(ctx, searchRequest)
	if err != nil {
		return fmt.Errorf("error searching for subject: %w", err)
	}

	// Check if subject was found
	if len(response.Results) == 0 {
		return fmt.Errorf("subject '%s' not found or you don't have permission to view it", subject)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		result := response.Results[0]

		fmt.Printf("Subject: %s\n", result.Subject)
		fmt.Printf("========================================\n\n")

		// Display content
		if result.Content != nil {
			fmt.Printf("Content:\n")
			contentJSON, _ := json.MarshalIndent(result.Content, "  ", "  ")
			fmt.Printf("  %s\n\n", string(contentJSON))
		}

		// Display score
		if result.Score > 0 {
			fmt.Printf("Relevance Score: %.4f\n", result.Score)
		}

		// Display highlights if any
		if len(result.Highlight) > 0 {
			fmt.Printf("\nHighlights:\n")
			for field, highlights := range result.Highlight {
				fmt.Printf("  %s:\n", field)
				for _, hl := range highlights {
					fmt.Printf("    - %s\n", hl)
				}
			}
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"Subject", "Content", "Score"}
		if err := formatter.FormatOutput(response.Results[0], headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
