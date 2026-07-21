// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/search"
	"github.com/spf13/cobra"
)

var (
	indexCreateDisplayName string
	indexCreateDescription string
	indexCreateMonitored   bool
)

// IndexCreateCmd represents the search index create command
var IndexCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Globus Search index",
	Long: `Create a new Globus Search index for storing searchable documents.

After creating an index, you can ingest documents and grant permissions
to other users via roles.

Examples:
  # Create a simple index
  globus search index create --display-name "My Research Data"

  # Create with description
  globus search index create \
    --display-name "Climate Data" \
    --description "Global climate research datasets"

  # Create with monitoring enabled
  globus search index create \
    --display-name "Production Index" \
    --monitored`,
	RunE: runIndexCreate,
}

func init() {
	IndexCreateCmd.Flags().StringVar(&indexCreateDisplayName, "display-name", "", "Display name for the index (required)")
	IndexCreateCmd.Flags().StringVar(&indexCreateDescription, "description", "", "Description of the index")
	IndexCreateCmd.Flags().BoolVar(&indexCreateMonitored, "monitored", false, "Enable monitoring for the index")

	_ = IndexCreateCmd.MarkFlagRequired("display-name")
}

func runIndexCreate(cmd *cobra.Command, args []string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Search client authorized for the current profile.
	searchClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Build create request. Upstream create_index accepts only display_name and
	// description; the --monitored flag is retained but has no effect.
	createRequest := &search.IndexCreate{
		DisplayName: indexCreateDisplayName,
		Description: indexCreateDescription,
	}

	// Create index
	index, err := searchClient.CreateIndex(ctx, createRequest)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Index created successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Index ID:     %s\n", index.ID)
	fmt.Fprintf(os.Stdout, "Display Name: %s\n", index.DisplayName)
	if index.Description != "" {
		fmt.Fprintf(os.Stdout, "Description:  %s\n", index.Description)
	}
	fmt.Fprintf(os.Stdout, "Status:       %s\n", index.Status)
	if !index.Created.IsZero() {
		fmt.Fprintf(os.Stdout, "Created At:   %s\n", index.Created.Format(time.RFC3339))
	}

	return nil
}
