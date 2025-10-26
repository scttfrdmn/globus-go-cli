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

	IndexCreateCmd.MarkFlagRequired("display-name")
}

func runIndexCreate(cmd *cobra.Command, args []string) error {
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

	// Build create request
	createRequest := &search.IndexCreateRequest{
		DisplayName: indexCreateDisplayName,
		Description: indexCreateDescription,
		IsMonitored: indexCreateMonitored,
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
	fmt.Fprintf(os.Stdout, "Active:       %t\n", index.IsActive)
	fmt.Fprintf(os.Stdout, "Created By:   %s\n", index.CreatedBy)
	fmt.Fprintf(os.Stdout, "Created At:   %s\n", index.CreatedAt.Format(time.RFC3339))

	return nil
}
