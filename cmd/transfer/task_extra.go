// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/scttfrdmn/globus-go-cli/pkg/output"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

var (
	taskUpdateLabel    string
	taskUpdateDeadline string
)

// TaskExtraSubcommands returns the additional task subcommands (event-list,
// pause-info, update) so the task group can attach them.
func TaskExtraSubcommands() []*cobra.Command {
	return []*cobra.Command{
		taskEventListCmd(),
		taskPauseInfoCmd(),
		taskUpdateCmd(),
	}
}

// taskEventListCmd returns the task event-list command
func taskEventListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event-list TASK_ID",
		Short: "List events for a Globus Transfer task",
		Long: `List events for a specific Globus Transfer task.

This command displays the event log for a transfer task, including
timestamps, event codes, and descriptions.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTaskEvents(cmd, args[0])
		},
	}

	// Add flags
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of events to return")

	return cmd
}

// taskPauseInfoCmd returns the task pause-info command
func taskPauseInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause-info TASK_ID",
		Short: "Show why a Globus Transfer task is paused",
		Long: `Show pause information for a specific Globus Transfer task.

This command reports the rules and reasons a transfer task is currently
paused, if any.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showTaskPauseInfo(cmd, args[0])
		},
	}
}

// taskUpdateCmd returns the task update command
func taskUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update TASK_ID",
		Short: "Update a Globus Transfer task's label or deadline",
		Long: `Update a Globus Transfer task's label and/or deadline.

This command updates the mutable fields of a transfer task. Only the flags
you set are sent to the service.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateTask(cmd, args[0])
		},
	}

	// Add flags
	cmd.Flags().StringVar(&taskUpdateLabel, "label", "", "New label for the task")
	cmd.Flags().StringVar(&taskUpdateDeadline, "deadline", "", "New deadline for the task")

	return cmd
}

// listTaskEvents lists events for a task
func listTaskEvents(cmd *cobra.Command, taskID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	resp, err := transferClient.TaskEventList(ctx, taskID, &transfer.ListTaskEventsOptions{Limit: limit})
	if err != nil {
		return fmt.Errorf("failed to list task events: %w", err)
	}

	// For a --jmespath/--jq expression or JSON, emit the enveloped service
	// document ({"DATA_TYPE","DATA":[...]}), matching the Python CLI. For unix
	// (tab-delimited) emit the flat event rows.
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	if formatter.Format == output.FormatJSON {
		return formatter.FormatOutput(resp, nil)
	}
	if formatter.Format == output.FormatUnix {
		return formatter.FormatOutput(resp.Data, nil)
	}

	fmt.Printf("Events for task %s:\n", taskID)
	for _, event := range resp.Data {
		fmt.Println()
		if v, ok := event["time"]; ok {
			fmt.Printf("  Time:        %v\n", v)
		}
		if v, ok := event["code"]; ok {
			fmt.Printf("  Code:        %v\n", v)
		}
		if v, ok := event["description"]; ok {
			fmt.Printf("  Description: %v\n", v)
		}
		if v, ok := event["is_error"]; ok {
			fmt.Printf("  Is Error:    %v\n", v)
		}
	}

	return nil
}

// showTaskPauseInfo shows pause information for a task
func showTaskPauseInfo(cmd *cobra.Command, taskID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	resp, err := transferClient.TaskPauseInfo(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task pause info: %w", err)
	}

	// Route through the shared formatter (GenericResponse is a map).
	format := viper.GetString("format")
	formatter := output.NewFormatter(format, cmd.OutOrStdout())
	return formatter.FormatOutput(resp, nil)
}

// updateTask updates a task's label and/or deadline
func updateTask(cmd *cobra.Command, taskID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build a v4 Transfer client authorized for the current profile.
	transferClient, err := getClient(ctx)
	if err != nil {
		return err
	}

	// Build the update document with only the fields that were set.
	doc := map[string]interface{}{"DATA_TYPE": "task"}
	if cmd.Flags().Changed("label") {
		doc["label"] = taskUpdateLabel
	}
	if cmd.Flags().Changed("deadline") {
		doc["deadline"] = taskUpdateDeadline
	}

	if _, err := transferClient.UpdateTask(ctx, taskID, doc); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("Successfully updated task %s\n", taskID)
	return nil
}
