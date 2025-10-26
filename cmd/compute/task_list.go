// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	taskListLimit int
)

// TaskListCmd represents the compute task list command
var TaskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List recent task executions.

Examples:
  # List recent tasks
  globus compute task list

  # Limit results
  globus compute task list --limit 50

  # JSON output for scripting
  globus compute task list --format json`,
	RunE: runTaskList,
}

func init() {
	TaskListCmd.Flags().IntVar(&taskListLimit, "limit", 25, "Maximum number of tasks to return")
}

func runTaskList(cmd *cobra.Command, args []string) error {
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

	// Build list options
	options := &compute.TaskListOptions{
		PerPage: taskListLimit,
	}

	// List tasks
	taskList, err := computeClient.ListTasks(ctx, options)
	if err != nil {
		return fmt.Errorf("error listing tasks: %w", err)
	}

	// Format output
	format := viper.GetString("format")

	if format == "text" {
		// Text output - human readable table
		if len(taskList.Tasks) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}

		fmt.Printf("Task IDs:\n")
		fmt.Printf("=========\n\n")

		for i, taskID := range taskList.Tasks {
			fmt.Printf("%d. %s\n", i+1, taskID)
		}

		fmt.Printf("\nTotal: %d task(s)\n", taskList.Total)
	} else {
		// JSON output - output task list structure
		if format == "json" {
			output, _ := json.MarshalIndent(taskList, "", "  ")
			fmt.Println(string(output))
		} else {
			// CSV output - just the task IDs
			fmt.Println("TaskID")
			for _, taskID := range taskList.Tasks {
				fmt.Println(taskID)
			}
		}
	}

	return nil
}
