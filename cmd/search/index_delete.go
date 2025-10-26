// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package search

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var indexDeleteConfirm bool

// IndexDeleteCmd represents the search index delete command
var IndexDeleteCmd = &cobra.Command{
	Use:   "delete INDEX_ID",
	Short: "Delete a Globus Search index",
	Long: `Delete a Globus Search index permanently.

This will delete the index and all its documents. This action cannot be undone.

Examples:
  # Delete with confirmation prompt
  globus search index delete INDEX_ID

  # Delete without confirmation
  globus search index delete INDEX_ID --confirm`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexDelete,
}

func init() {
	IndexDeleteCmd.Flags().BoolVar(&indexDeleteConfirm, "confirm", false, "Skip confirmation prompt")
}

func runIndexDelete(cmd *cobra.Command, args []string) error {
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

	// Get index first to display name and confirm
	index, err := searchClient.GetIndex(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error getting index: %w", err)
	}

	// Confirmation prompt unless --confirm flag is set
	if !indexDeleteConfirm {
		fmt.Printf("Are you sure you want to delete index '%s' (%s)? [y/N]: ", index.DisplayName, indexID)
		fmt.Println("\nWarning: This will permanently delete the index and all its documents!")
		fmt.Print("Type 'yes' to confirm: ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading confirmation: %w", err)
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "yes" {
			fmt.Println("Index deletion cancelled.")
			return nil
		}
	}

	// Delete the index
	err = searchClient.DeleteIndex(ctx, indexID)
	if err != nil {
		return fmt.Errorf("error deleting index: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Index deleted successfully!\n\n")
	fmt.Fprintf(os.Stdout, "Index ID:     %s\n", indexID)
	fmt.Fprintf(os.Stdout, "Display Name: %s\n", index.DisplayName)

	return nil
}
