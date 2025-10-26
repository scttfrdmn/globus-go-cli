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

// SubjectDeleteCmd represents the search subject delete command
var SubjectDeleteCmd = &cobra.Command{
	Use:   "delete INDEX_ID SUBJECT",
	Short: "Delete a subject from a Globus Search index",
	Long: `Submit a delete task to remove a subject from a Search index.

This requires writer or stronger privileges on the index. Deletions are
queued as tasks and are not guaranteed to be immediate.

Examples:
  # Delete a subject
  globus search subject delete INDEX_ID my-document-id

  # The command returns a task ID for monitoring progress
  globus search subject delete INDEX_ID doc123`,
	Args: cobra.ExactArgs(2),
	RunE: runSubjectDelete,
}

func runSubjectDelete(cmd *cobra.Command, args []string) error {
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

	// Build delete request
	deleteRequest := &search.DeleteDocumentsRequest{
		IndexID:  indexID,
		Subjects: []string{subject},
	}

	// Execute delete
	response, err := searchClient.DeleteDocuments(ctx, deleteRequest)
	if err != nil {
		return fmt.Errorf("error deleting subject: %w", err)
	}

	// Display success message
	fmt.Fprintf(os.Stdout, "Subject deletion task submitted!\n\n")
	fmt.Fprintf(os.Stdout, "Task ID:    %s\n", response.Task.TaskID)
	fmt.Fprintf(os.Stdout, "Subject:    %s\n", subject)
	fmt.Fprintf(os.Stdout, "Status:     %s\n", response.Task.ProcessingState)

	if response.Task.TaskID != "" {
		fmt.Fprintf(os.Stdout, "\nCheck task status with: globus search task show %s\n", response.Task.TaskID)
	}

	return nil
}
