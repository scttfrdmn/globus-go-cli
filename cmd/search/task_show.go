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

// TaskShowCmd represents the search task show command
var TaskShowCmd = &cobra.Command{
	Use:   "show TASK_ID",
	Short: "Show status of a Globus Search task",
	Long: `Display the status and details of a Globus Search indexing task.

Tasks are created when you ingest or delete documents. Use this command
to monitor the progress and check for errors.

Examples:
  # Show task status
  globus search task show TASK_ID

  # Show with JSON output
  globus search task show TASK_ID --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runTaskShow,
}

func runTaskShow(cmd *cobra.Command, args []string) error {
	taskID := args[0]

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

	// Get task status
	taskStatus, err := searchClient.GetTaskStatus(ctx, taskID)
	if err != nil {
		return fmt.Errorf("error getting task status: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Task Information\n")
		fmt.Printf("================\n\n")
		fmt.Printf("Task ID:    %s\n", taskStatus.TaskID)
		fmt.Printf("State:      %s\n", taskStatus.State)
		fmt.Printf("Created At: %s\n", taskStatus.CreatedAt)
		if taskStatus.CompletedAt != "" {
			fmt.Printf("Completed:  %s\n", taskStatus.CompletedAt)
		}

		fmt.Printf("\nDocument Status\n")
		fmt.Printf("---------------\n")
		fmt.Printf("Total:      %d documents\n", taskStatus.TotalDocuments)
		fmt.Printf("Succeeded:  %d documents\n", taskStatus.SuccessDocuments)
		fmt.Printf("Failed:     %d documents\n", taskStatus.FailedDocuments)
		fmt.Printf("Errors:     %d\n", taskStatus.ErrorCount)

		// Show failed subjects if any
		if len(taskStatus.FailedSubjects) > 0 {
			fmt.Printf("\nFailed Subjects:\n")
			for _, subject := range taskStatus.FailedSubjects {
				fmt.Printf("  - %s\n", subject)
			}
		}

		if taskStatus.DetailLocation != "" {
			fmt.Printf("\nDetails: %s\n", taskStatus.DetailLocation)
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"TaskID", "State", "TotalDocuments", "SuccessDocuments", "FailedDocuments", "ErrorCount"}
		if err := formatter.FormatOutput(taskStatus, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
