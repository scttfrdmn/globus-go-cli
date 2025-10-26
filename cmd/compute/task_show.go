// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

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
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TaskShowCmd represents the compute task show command
var TaskShowCmd = &cobra.Command{
	Use:   "show TASK_ID",
	Short: "Show status and results of a task",
	Long: `Display detailed information about a task execution.

This includes the task's status, result (if completed), and any error
information if the task failed.

Examples:
  # Show task status
  globus compute task show TASK_ID

  # Show task with JSON output
  globus compute task show TASK_ID --format json`,
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

	// Create compute client
	computeClient, err := compute.NewClient(
		compute.WithAuthorizer(coreAuthorizer),
	)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get task status
	taskStatus, err := computeClient.GetTaskStatus(ctx, taskID)
	if err != nil {
		return fmt.Errorf("error getting task status: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable
		fmt.Printf("Task Details\n")
		fmt.Printf("============\n\n")

		fmt.Printf("Task ID:       %s\n", taskStatus.TaskID)
		fmt.Printf("Status:        %s\n", taskStatus.Status)

		if !taskStatus.CompletedAt.IsZero() {
			fmt.Printf("Completed:     %s\n", taskStatus.CompletedAt.Format(time.RFC3339))
		}

		// Display result if available
		if taskStatus.Result != nil {
			fmt.Printf("\nResult:\n")
			resultJSON, _ := json.MarshalIndent(taskStatus.Result, "  ", "  ")
			fmt.Printf("%s\n", string(resultJSON))
		}

		// Display exception if task failed
		if taskStatus.Exception != "" {
			fmt.Printf("\nException:\n%s\n", taskStatus.Exception)
		}
	} else {
		// JSON or CSV output
		formatter := output.NewFormatter(format, os.Stdout)
		headers := []string{"TaskID", "Status", "CompletedAt"}
		if err := formatter.FormatOutput(taskStatus, headers); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}
