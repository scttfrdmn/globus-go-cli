// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

var (
	taskWait     bool
	taskWaitTime int
	taskFilter   string
)

// TaskCmd returns the task command
func TaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Commands for managing Globus Transfer tasks",
		Long: `Commands for managing Globus Transfer tasks including listing,
showing details, canceling, and waiting for completion.`,
	}

	// Add task subcommands
	taskCmd.AddCommand(
		taskListCmd(),
		taskShowCmd(),
		taskCancelCmd(),
		taskWaitCmd(),
		taskEventListCmd(),
		taskPauseInfoCmd(),
		taskUpdateCmd(),
	)

	return taskCmd
}

// taskListCmd returns the task list command
func taskListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Globus Transfer tasks",
		Long: `List Globus Transfer tasks for the current user.

This command lists transfer tasks with filtering options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTasks(cmd)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&taskFilter, "filter", "", "Filter tasks by status (active, inactive, completed, failed)")
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of tasks to return")

	return cmd
}

// taskShowCmd returns the task show command
func taskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show TASK_ID",
		Short: "Show Globus Transfer task details",
		Long: `Show details for a specific Globus Transfer task.

This command displays detailed information about a transfer task
including status, timing, and file details.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showTask(cmd, args[0])
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&taskWait, "wait", false, "Wait for the task to complete")
	cmd.Flags().IntVar(&taskWaitTime, "timeout", 300, "Maximum time to wait in seconds")

	return cmd
}

// taskCancelCmd returns the task cancel command
func taskCancelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel TASK_ID",
		Short: "Cancel a Globus Transfer task",
		Long: `Cancel a Globus Transfer task.

This command cancels a running transfer task. It cannot cancel 
completed or already canceled tasks.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cancelTask(cmd, args[0])
		},
	}
}

// taskWaitCmd returns the task wait command
func taskWaitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait TASK_ID",
		Short: "Wait for a Globus Transfer task to complete",
		Long: `Wait for a Globus Transfer task to complete.

This command polls the task status until it completes or fails,
showing progress information while waiting.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return waitForTask(cmd, args[0], taskWaitTime)
		},
	}

	// Add flags
	cmd.Flags().IntVar(&taskWaitTime, "timeout", 300, "Maximum time to wait in seconds")

	return cmd
}

// listTasks lists transfer tasks
func listTasks(cmd *cobra.Command) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Prepare options for listing tasks
	options := &transfer.ListTasksOptions{
		Limit: limit,
	}

	// Add filters based on flags (v4 accepts a list of statuses)
	if taskFilter != "" {
		options.FilterStatus = []string{taskFilter}
	}

	// Get the tasks
	tasks, err := transferClient.ListTasks(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Get output format
	format := viper.GetString("format")

	// Format and display the results
	formatter := output.NewFormatter(format, cmd.OutOrStdout())

	// For JSON, emit the enveloped service document ({"DATA_TYPE","DATA":[...]}),
	// matching the Python CLI's JSON output shape.
	if formatter.Format == output.FormatJSON {
		return formatter.FormatOutput(tasks, nil)
	}

	// Define the headers
	headers := []string{"TaskID", "Status", "Type", "Source", "Destination", "Label"}

	// Create a slice of task entries for formatting
	type taskEntry struct {
		TaskID      string
		Status      string
		Type        string
		Source      string
		Destination string
		Label       string
	}

	entries := make([]taskEntry, 0, len(tasks.Data))

	for _, task := range tasks.Data {
		source := "N/A"
		if task.SourceEndpoint != "" {
			source = task.SourceEndpoint
		}

		destination := "N/A"
		if task.DestinationEndpoint != "" {
			destination = task.DestinationEndpoint
		}

		entry := taskEntry{
			TaskID:      task.TaskID,
			Status:      task.Status,
			Type:        task.Type,
			Source:      source,
			Destination: destination,
			Label:       task.Label,
		}

		entries = append(entries, entry)
	}

	// Display the results using the formatter
	if err := formatter.FormatOutput(entries, headers); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}

// showTask shows details for a specific task
func showTask(cmd *cobra.Command, taskID string) error {
	// If wait flag is specified, wait for the task
	if taskWait {
		return waitForTask(cmd, taskID, taskWaitTime)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get the task
	task, err := transferClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// For json/unix or a --jmespath/--jq expression, route through the shared
	// formatter (raw task document). Otherwise render the text detail view.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON || formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(task, nil)
	}

	{
		// Output as text
		fmt.Println("Task Details:")
		fmt.Printf("  Task ID:        %s\n", task.TaskID)
		fmt.Printf("  Status:         %s\n", task.Status)
		fmt.Printf("  Type:           %s\n", task.Type)
		fmt.Printf("  Label:          %s\n", task.Label)

		// Format dates - RequestTime is time.Time in the v4 SDK.
		fmt.Printf("  Request Time:   %s\n", task.RequestTime.Format("2006-01-02 15:04:05"))

		// CompletionTime is time.Time in v4; zero means not yet complete.
		if !task.CompletionTime.IsZero() {
			fmt.Printf("  Completion Time: %s\n", task.CompletionTime.Format("2006-01-02 15:04:05"))
		}

		// Format task status with color
		switch task.Status {
		case "SUCCEEDED":
			color.Green("  Task succeeded")
		case "FAILED":
			if task.NiceStatus != "" {
				color.Red("  Task failed: %s", task.NiceStatus)
			} else {
				color.Red("  Task failed")
			}
		case "ACTIVE":
			color.Yellow("  Task is active")
		}

		// Show endpoint information
		fmt.Println("\nEndpoints:")
		fmt.Printf("  Source:      %s\n", task.SourceEndpoint)
		fmt.Printf("  Destination: %s\n", task.DestinationEndpoint)

		// Show transfer stats
		fmt.Println("\nTransfer Stats:")
		fmt.Printf("  Files:          %d\n", task.FilesTransferred)
		fmt.Printf("  Directories:    %d\n", task.Directories)
		fmt.Printf("  Files Skipped:  %d\n", task.FilesSkipped)
		fmt.Printf("  Bytes Transferred: %d\n", task.BytesTransferred)
	}

	return nil
}

// cancelTask cancels a task
func cancelTask(cmd *cobra.Command, taskID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Get the current task status first
	task, err := transferClient.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task status: %w", err)
	}

	// Check if the task can be canceled
	if task.Status != "ACTIVE" {
		return fmt.Errorf("task %s is not active (status: %s), cannot cancel", taskID, task.Status)
	}

	// Cancel the task - CancelTask returns OperationResult and error in v0.9.17
	result, err := transferClient.CancelTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}
	_ = result // Using result to avoid unused variable warning

	fmt.Printf("Successfully canceled task %s\n", taskID)
	return nil
}

// waitForTask waits for a task to complete
func waitForTask(cmd *cobra.Command, taskID string, timeout int) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Start spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = fmt.Sprintf(" Waiting for task %s to complete...", taskID)
	s.Start()
	defer s.Stop()

	// Poll for task completion
	pollInterval := 5 * time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for task completion")
		case <-ticker.C:
			// Get the task status
			task, err := transferClient.GetTask(ctx, taskID)
			if err != nil {
				return fmt.Errorf("failed to get task status: %w", err)
			}

			// Update spinner message
			if task.FilesTransferred > 0 || task.BytesTransferred > 0 {
				s.Suffix = fmt.Sprintf(" Waiting for task %s: %d files, %d bytes transferred",
					taskID, task.FilesTransferred, task.BytesTransferred)
			}

			// Check if the task has completed
			if task.Status != "ACTIVE" {
				s.Stop()

				// Display final status
				if task.Status == "SUCCEEDED" {
					color.Green("Task %s completed successfully", taskID)
					fmt.Printf("Transferred %d files (%d bytes)\n", task.FilesTransferred, task.BytesTransferred)
				} else if task.Status == "FAILED" {
					// NiceStatus not available in v0.9.17
					color.Red("Task %s failed", taskID)
				} else {
					fmt.Printf("Task %s status: %s\n", taskID, task.Status)
				}

				return nil
			}
		}
	}
}
