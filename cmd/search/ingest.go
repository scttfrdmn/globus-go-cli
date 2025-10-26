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
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ingestFile      string
	ingestData      string
	ingestBatchSize int
)

// IngestCmd represents the search ingest command
var IngestCmd = &cobra.Command{
	Use:   "ingest INDEX_ID",
	Short: "Ingest documents into a Globus Search index",
	Long: `Ingest documents into a Globus Search index.

Documents can be provided via a JSON file or as a JSON string. Each document
must have a 'subject' field and 'entries' containing the searchable data.

Document Format:
{
  "subject": "unique-identifier",
  "visible_to": ["public"],
  "entries": [
    {
      "entry_id": "entry1",
      "content": {
        "field1": "value1",
        "field2": "value2"
      }
    }
  ]
}

Examples:
  # Ingest from a file
  globus search ingest INDEX_ID --file documents.json

  # Ingest a single document
  globus search ingest INDEX_ID --data '{"subject":"doc1","visible_to":["public"],"entries":[{"content":{"title":"Test"}}]}'

  # Batch ingest with custom batch size
  globus search ingest INDEX_ID --file large-dataset.json --batch-size 100`,
	Args: cobra.ExactArgs(1),
	RunE: runSearchIngest,
}

func init() {
	IngestCmd.Flags().StringVar(&ingestFile, "file", "", "JSON file containing documents to ingest")
	IngestCmd.Flags().StringVar(&ingestData, "data", "", "JSON string containing documents to ingest")
	IngestCmd.Flags().IntVar(&ingestBatchSize, "batch-size", 50, "Batch size for ingestion")
}

func runSearchIngest(cmd *cobra.Command, args []string) error {
	indexID := args[0]

	// Validate input
	if ingestFile == "" && ingestData == "" {
		return fmt.Errorf("either --file or --data must be provided")
	}
	if ingestFile != "" && ingestData != "" {
		return fmt.Errorf("cannot specify both --file and --data")
	}

	// Read documents
	var documentsJSON []byte
	var err error

	if ingestFile != "" {
		documentsJSON, err = os.ReadFile(ingestFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		documentsJSON = []byte(ingestData)
	}

	// Parse documents
	var documents []search.SearchDocument
	if err := json.Unmarshal(documentsJSON, &documents); err != nil {
		// Try parsing as a single document
		var singleDoc search.SearchDocument
		if err2 := json.Unmarshal(documentsJSON, &singleDoc); err2 != nil {
			return fmt.Errorf("failed to parse documents (tried both array and single document): %w / %w", err, err2)
		}
		documents = []search.SearchDocument{singleDoc}
	}

	if len(documents) == 0 {
		return fmt.Errorf("no documents found in input")
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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Build ingest request
	ingestRequest := &search.IngestRequest{
		IndexID:   indexID,
		Documents: documents,
	}

	// Execute ingest
	response, err := searchClient.IngestDocuments(ctx, ingestRequest)
	if err != nil {
		return fmt.Errorf("error ingesting documents: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Documents ingested successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Task ID:      %s\n", response.Task.TaskID)
	fmt.Fprintf(os.Stdout, "Total:        %d documents\n", response.Total)
	fmt.Fprintf(os.Stdout, "Succeeded:    %d documents\n", response.Succeeded)
	fmt.Fprintf(os.Stdout, "Failed:       %d documents\n", response.Failed)
	if response.Task.TaskID != "" {
		fmt.Fprintf(os.Stdout, "\nCheck task status with: globus search task show %s\n", response.Task.TaskID)
	}

	return nil
}
