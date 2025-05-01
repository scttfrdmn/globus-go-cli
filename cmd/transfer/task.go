// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	authcmd "github.com/scttfrdmn/globus-go-cli/cmd/auth"
	"github.com/scttfrdmn/globus-go-cli/pkg/config"
)

var (
	// Filter options for task listing
	taskFilterStatus string
	taskFilterType   string
	taskFilterLabel  string
	taskLimit        int
	taskOffset       int
)

// TaskCmd returns the task command
func TaskCmd() *cobra.Command {
	// taskCmd represents the task command
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Commands for managing Globus transfer tasks",
		Long: `Commands for managing Globus Transfer tasks, including listing,
showing details, cancelling, and waiting for tasks.`,
	}

	// Add task subcommands
	taskCmd.AddCommand(
		taskListCmd(),
		taskShowCmd(),
		taskCancelCmd(),
		taskWaitCmd(),
	)

	return taskCmd
}

// taskListCmd returns the task list command
func taskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Globus transfer tasks",
		Long: `List Globus Transfer tasks.

This command lists transfer tasks, filtered by various criteria.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTasks(cmd)
		},
	}

	// Add flags for filtering
	cmd.Flags().StringVar(&taskFilterStatus, "status", "", "Filter by status (ACTIVE, SUCCEEDED, FAILED)")
	cmd.Flags().StringVar(&taskFilterType, "type", "", "Filter by type (TRANSFER, DELETE)")
	cmd.Flags().StringVar(&taskFilterLabel, "label", "", "Filter by label")
	cmd.Flags().IntVar(&taskLimit, "limit", 25, "Maximum number of tasks to return")
	cmd.Flags().IntVar(&taskOffset, "offset", 0, "Offset for pagination")

	return cmd
}

// taskShowCmd returns the task show command
func taskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [TASK_ID]",
		Short: "Show Globus transfer task details",
		Long: `Show details for a specific Globus Transfer task.

This command displays detailed information about a transfer task.
If no task ID is provided, it shows the most recent task.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var taskID string
			if len(args) > 0 {
				taskID = args[0]
			} else {
				var err error
				taskID, err = getLastTaskID()
				if err != nil {
					return err
				}
			}
			return showTask(cmd, taskID)
		},
	}

	return cmd
}

// taskCancelCmd returns the task cancel command
func taskCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [TASK_ID]",
		Short: "Cancel a Globus transfer task",
		Long: `Cancel a Globus Transfer task.

This command cancels a running transfer task.
If no task ID is provided, it cancels the most recent task.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var taskID string
			if len(args) > 0 {
				taskID = args[0]
			} else {
				var err error
				taskID, err = getLastTaskID()
				if err != nil {
					return err
				}
			}
			return cancelTask(cmd, taskID)
		},
	}

	return cmd
}

// taskWaitCmd returns the task wait command
func taskWaitCmd() *cobra.Command {
	var progressFlag bool

	cmd := &cobra.Command{
		Use:   "wait [TASK_ID]",
		Short: "Wait for a Globus transfer task to complete",
		Long: `Wait for a Globus Transfer task to complete.

This command polls a transfer task until it completes (succeeds or fails).
If no task ID is provided, it waits for the most recent task.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var taskID string
			if len(args) > 0 {
				taskID = args[0]
			} else {
				var err error
				taskID, err = getLastTaskID()
				if err != nil {
					return err
				}
			}

			// Get the transfer client
			transferClient, err := getTransferClient()
			if err != nil {
				return err
			}

			// Wait for the task to complete
			return waitForTask(transferClient, taskID, progressFlag)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&progressFlag, "progress", true, "Show progress bar")

	return cmd
}

// listTasks lists Globus transfer tasks
func listTasks(cmd *cobra.Command) error {
	// Get the transfer client
	transferClient, err := getTransferClient()
	if err != nil {
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Prepare options for listing tasks
	options := &pkg.TaskListOptions{
		Limit:  taskLimit,
		Offset: taskOffset,
	}

	// Add filters based on flags
	if taskFilterStatus != "" {
		options.FilterStatus = taskFilterStatus
	}
	if taskFilterType != "" {
		options.FilterType = taskFilterType
	}
	if taskFilterLabel != "" {
		options.FilterLabel = taskFilterLabel
	}

	// Get the tasks
	taskList, err := transferClient.ListTasks(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Get output format
	format := viper.GetString("format")
	if format == "" {
		format = "text"
	}

	// Display the results based on format
	switch strings.ToLower(format) {
	case "json":
		// Output as JSON
		jsonData, err := json.MarshalIndent(taskList, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case "csv":
		// Output as CSV
		fmt.Println("task_id,status,type,label,source_endpoint,destination_endpoint,requested_time")
		for _, task := range taskList.Tasks {
			fmt.Printf("%s,%s,%s,%s,%s,%s,%s\n",
				task.TaskID,
				task.Status,
				task.Type,
				strings.ReplaceAll(task.Label, ",", " "),
				task.SourceEndpointDisplayName,
				task.DestinationEndpointDisplayName,
				task.RequestTime,
			)
		}
	default:
		// Output as text table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Task ID\tStatus\tType\tLabel\tRequested Time")
		fmt.Fprintln(w, "-------\t------\t----\t-----\t--------------")

		for _, task := range taskList.Tasks {
			// Format the request time
			requestTime, _ := time.Parse(time.RFC3339, task.RequestTime)
			requestTimeStr := requestTime.Format("2006-01-02 15:04:05")

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				task.TaskID,
				task.Status,
				task.Type,
				task.Label,
				requestTimeStr,
			)
		}
		w.Flush()

		// Display count
		fmt.Printf("\nShowing %d of %d tasks\n", len(taskList.Tasks), taskList.Total)
	}

	return nil
}

// showTask shows details for a specific Globus transfer task
func showTask(cmd *cobra.Command, taskID string) error {
	// Get the transfer client
	transferClient, err := getTransferClient()
	if err != nil {
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the task
	task, err := transferClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get output format
	format := viper.GetString("format")
	if format == "" {
		format = "text"
	}

	// Display the results based on format
	switch strings.ToLower(format) {
	case "json":
		// Output as JSON
		jsonData, err := json.MarshalIndent(task, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	default:
		// Output as text
		fmt.Println("Task Details:")
		fmt.Printf("  Task ID:      %s\n", task.TaskID)
		fmt.Printf("  Status:       %s\n", task.Status)
		fmt.Printf("  Type:         %s\n", task.Type)
		fmt.Printf("  Label:        %s\n", task.Label)
		fmt.Printf("  Owner:        %s\n", task.Owner)
		
		// Format times
		requestTime, _ := time.Parse(time.RFC3339, task.RequestTime)
		fmt.Printf("  Requested:    %s\n", requestTime.Format("2006-01-02 15:04:05"))
		
		if task.CompletionTime != "" {
			completionTime, _ := time.Parse(time.RFC3339, task.CompletionTime)
			fmt.Printf("  Completed:    %s\n", completionTime.Format("2006-01-02 15:04:05"))
			
			// Calculate duration
			duration := completionTime.Sub(requestTime)
			fmt.Printf("  Duration:     %s\n", formatDuration(int(duration.Seconds())))
		}
		
		// Print transfer specific information
		if task.Type == "TRANSFER" {
			fmt.Printf("\nTransfer Info:\n")
			fmt.Printf("  Source:       %s (%s)\n", task.SourceEndpointDisplayName, task.SourceEndpoint)
			fmt.Printf("  Destination:  %s (%s)\n", task.DestinationEndpointDisplayName, task.DestEndpoint)
			fmt.Printf("  Files:        %d transferred, %d skipped, %d failed\n",
				task.FilesTransferred, task.FilesSkipped, task.FilesSkippedFail)
			fmt.Printf("  Directories:  %d created\n", task.DirectoriesCreated)
			fmt.Printf("  Size:         %s transferred\n", formatSize(task.BytesTransferred))
			
			// Calculate progress
			if task.Status == "ACTIVE" {
				percent := percentComplete(task.BytesTransferred, task.BytesExpected)
				fmt.Printf("  Progress:     %.1f%% (%s of %s)\n",
					percent,
					formatSize(task.BytesTransferred),
					formatSize(task.BytesExpected),
				)
			}
			
			// Show sync level if available
			if task.SyncLevel != 0 {
				fmt.Printf("  Sync Level:   %s\n", getSyncLevelString(strconv.Itoa(task.SyncLevel)))
			}
		} else if task.Type == "DELETE" {
			fmt.Printf("\nDelete Info:\n")
			fmt.Printf("  Endpoint:     %s (%s)\n", task.SourceEndpointDisplayName, task.SourceEndpoint)
			fmt.Printf("  Files:        %d deleted\n", task.FilesTransferred)
			fmt.Printf("  Directories:  %d deleted\n", task.DirectoriesCreated) // Repurposed for directories deleted
		}
		
		// Print additional information based on status
		if task.Status == "SUCCEEDED" {
			fmt.Printf("\nResult: Transfer completed successfully\n")
		} else if task.Status == "FAILED" {
			fmt.Printf("\nResult: Transfer failed: %s\n", task.NiceStatusShortDescription)
			if task.FatalErrorDescription != "" {
				fmt.Printf("Error details: %s\n", task.FatalErrorDescription)
			}
		}
	}

	return nil
}

// cancelTask cancels a Globus transfer task
func cancelTask(cmd *cobra.Command, taskID string) error {
	// Get the transfer client
	transferClient, err := getTransferClient()
	if err != nil {
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First, check if the task exists and is cancellable
	task, err := transferClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if the task is already complete
	if task.Status != "ACTIVE" {
		return fmt.Errorf("task is not active (status: %s), cannot cancel", task.Status)
	}

	// Show task info
	fmt.Printf("Cancelling task:\n")
	fmt.Printf("  Task ID: %s\n", task.TaskID)
	fmt.Printf("  Type:    %s\n", task.Type)
	fmt.Printf("  Label:   %s\n", task.Label)

	// Confirm with the user
	fmt.Print("\nAre you sure you want to cancel this task? [y/N] ")
	var confirm string
	fmt.Scanln(&confirm)
	if strings.ToLower(confirm) != "y" {
		fmt.Println("Cancelled. The task will continue.")
		return nil
	}

	// Cancel the task
	err = transferClient.CancelTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}

	fmt.Printf("Task %s has been cancelled.\n", taskID)
	return nil
}

// getLastTaskID gets the ID of the last executed task
func getLastTaskID() (string, error) {
	// Try to get the last task ID from the file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	taskFile := filepath.Join(homeDir, ".globus-cli", "last-task-id")
	data, err := os.ReadFile(taskFile)
	if err != nil {
		return "", fmt.Errorf("no task ID found. Please specify a task ID")
	}

	taskID := strings.TrimSpace(string(data))
	if taskID == "" {
		return "", fmt.Errorf("empty task ID found. Please specify a task ID")
	}

	return taskID, nil
}

// getTransferClient gets a configured transfer client
func getTransferClient() (*pkg.TransferClient, error) {
	// Get current profile
	profile := viper.GetString("profile")
	
	// Load token
	tokenInfo, err := authcmd.LoadToken(profile)
	if err != nil {
		return nil, fmt.Errorf("not logged in: %w", err)
	}

	// Check if token is valid
	if !authcmd.IsTokenValid(tokenInfo) {
		return nil, fmt.Errorf("token is expired, please login again")
	}

	// Load client configuration
	clientCfg, err := config.LoadClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load client configuration: %w", err)
	}

	// Create SDK config
	sdkConfig := pkg.NewConfig().
		WithClientID(clientCfg.ClientID).
		WithClientSecret(clientCfg.ClientSecret)

	// Create transfer client
	transferClient := sdkConfig.NewTransferClient(tokenInfo.AccessToken)

	return transferClient, nil
}

// getSyncLevelString converts a sync level value to a human-readable string
func getSyncLevelString(syncLevel string) string {
	switch syncLevel {
	case "0":
		return "None"
	case "1":
		return "Exists"
	case "2":
		return "Size"
	case "3":
		return "Modification Time"
	case "4":
		return "Checksum"
	default:
		return fmt.Sprintf("Unknown (%s)", syncLevel)
	}
}