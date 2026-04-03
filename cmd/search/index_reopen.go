// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
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

// IndexReopenCmd represents the search index reopen command.
// Added in Python SDK v4.0.0.
var IndexReopenCmd = &cobra.Command{
	Use:   "reopen INDEX_ID",
	Short: "Reopen a previously deleted Globus Search index",
	Long: `Reopen a previously deleted Globus Search index.

A deleted index can be reopened to restore access to its documents.
Added in Python SDK v4.0.0.

Examples:
  globus search index reopen INDEX_ID`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexReopen,
}

func runIndexReopen(cmd *cobra.Command, args []string) error {
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

	// Reopen the index
	index, err := searchClient.ReopenIndex(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error reopening index: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Index reopened successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Index ID:     %s\n", index.ID)
	fmt.Fprintf(os.Stdout, "Display Name: %s\n", index.DisplayName)
	fmt.Fprintf(os.Stdout, "Is Active:    %t\n", index.IsActive)

	return nil
}
