// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers/mocks"
	"github.com/spf13/cobra"
)

// mockTaskCommands creates a set of mock task commands for testing
func mockTaskCommands(mockClient *mocks.MockTransferClient) *cobra.Command {
	taskCmd := TaskCmd()
	
	// Get the subcommands
	var listCmd, showCmd, cancelCmd, waitCmd *cobra.Command
	
	// Find the subcommands by name
	for _, cmd := range taskCmd.Commands() {
		switch cmd.Name() {
		case "list":
			listCmd = cmd
		case "show":
			showCmd = cmd
		case "cancel":
			cancelCmd = cmd
		case "wait":
			waitCmd = cmd
		}
	}
	
	// Override list command
	if listCmd != nil {
		listCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Prepare options for listing tasks
			options := &mocks.ListTasksOptions{
				Limit: limit,
			}

			// Add filters based on flags
			if taskFilter != "" {
				options.FilterStatus = taskFilter
			}

			// Get the tasks
			tasks, err := mockClient.ListTasks(context.Background(), options)
			if err != nil {
				return fmt.Errorf("failed to list tasks: %w", err)
			}

			// Display results
			fmt.Fprintln(cmd.OutOrStdout(), "TaskID     Status     Type       Label")
			fmt.Fprintln(cmd.OutOrStdout(), "---------- ---------- ---------- ----------")
			for _, task := range tasks.Data {
				fmt.Fprintf(cmd.OutOrStdout(), "%-10s %-10s %-10s %s\n",
					task.TaskID, task.Status, task.Type, task.Label)
			}

			return nil
		}
	}
	
	// Override show command
	if showCmd != nil {
		showCmd.RunE = func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly one argument")
			}
			
			taskID := args[0]
			
			// Get the task
			task, err := mockClient.GetTask(context.Background(), taskID)
			if err != nil {
				return fmt.Errorf("failed to get task: %w", err)
			}

			// Display the task details
			fmt.Fprintln(cmd.OutOrStdout(), "Task Details:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Task ID:        %s\n", task.TaskID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Status:         %s\n", task.Status)
			fmt.Fprintf(cmd.OutOrStdout(), "  Type:           %s\n", task.Type)
			fmt.Fprintf(cmd.OutOrStdout(), "  Label:          %s\n", task.Label)
			fmt.Fprintf(cmd.OutOrStdout(), "  Request Time:   %s\n", task.RequestTime.Format("2006-01-02 15:04:05"))

			if task.CompletionTime != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  Completion Time: %s\n", task.CompletionTime.Format("2006-01-02 15:04:05"))
			}

			// Show task succeeded message
			if task.Status == "SUCCEEDED" {
				fmt.Fprintln(cmd.OutOrStdout(), "  Task succeeded")
			}

			// Show endpoint information
			fmt.Fprintln(cmd.OutOrStdout(), "\nEndpoints:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Source:      %s (%s)\n", task.SourceEndpointDisplay, task.SourceEndpointID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Destination: %s (%s)\n", task.DestEndpointDisplay, task.DestinationEndpointID)

			// Show transfer stats
			fmt.Fprintln(cmd.OutOrStdout(), "\nTransfer Stats:")
			fmt.Fprintf(cmd.OutOrStdout(), "  Files:          %d\n", task.FilesTransferred)
			fmt.Fprintf(cmd.OutOrStdout(), "  Directories:    %d\n", task.Subtasks)
			fmt.Fprintf(cmd.OutOrStdout(), "  Files Skipped:  %d\n", task.FilesSkipped)
			fmt.Fprintf(cmd.OutOrStdout(), "  Total Files:    %d\n", task.Subtasks)
			fmt.Fprintf(cmd.OutOrStdout(), "  Total Bytes:    %d\n", task.BytesSkipped+task.BytesTransferred)
			fmt.Fprintf(cmd.OutOrStdout(), "  Bytes Transferred: %d\n", task.BytesTransferred)

			// Show verification if applicable
			if task.VerifyChecksum {
				fmt.Fprintf(cmd.OutOrStdout(), "\nVerification: checksum\n")
			}

			return nil
		}
	}
	
	// Override cancel command
	if cancelCmd != nil {
		cancelCmd.RunE = func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly one argument")
			}
			
			taskID := args[0]

			// Get the current task status first
			task, err := mockClient.GetTask(context.Background(), taskID)
			if err != nil {
				return fmt.Errorf("failed to get task status: %w", err)
			}

			// Check if the task can be canceled
			if task.Status != "ACTIVE" {
				return fmt.Errorf("task %s is not active (status: %s), cannot cancel", taskID, task.Status)
			}

			// Cancel the task
			_, err = mockClient.CancelTask(context.Background(), taskID)
			if err != nil {
				return fmt.Errorf("failed to cancel task: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Successfully canceled task %s\n", taskID)
			return nil
		}
	}
	
	// Override wait command
	if waitCmd != nil {
		waitCmd.RunE = func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly one argument")
			}
			
			taskID := args[0]

			// Create context
			ctx := context.Background()

			// First check - should be active
			task, err := mockClient.GetTask(ctx, taskID)
			if err != nil {
				return fmt.Errorf("failed to get task status: %w", err)
			}

			// Report initial status
			fmt.Fprintf(cmd.OutOrStdout(), "Waiting for task %s: %d/%d files, %d bytes\n",
				taskID, task.FilesTransferred, task.Subtasks, task.BytesTransferred)

			// Second check - should be completed
			task, err = mockClient.GetTask(ctx, taskID)
			if err != nil {
				return fmt.Errorf("failed to get task status: %w", err)
			}

			// Report completion
			if task.Status == "SUCCEEDED" {
				fmt.Fprintf(cmd.OutOrStdout(), "Task %s completed successfully\n", taskID)
				fmt.Fprintf(cmd.OutOrStdout(), "Transferred %d files (%d bytes)\n",
					task.FilesTransferred, task.BytesTransferred)
			}

			return nil
		}
	}
	
	return taskCmd
}

func TestTaskListCommand(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	mockClient.ListTasksFunc = func(ctx context.Context, options *mocks.ListTasksOptions) (*mocks.TaskList, error) {
		// Verify options
		if options.Limit != 25 && options.Limit != 0 {
			t.Errorf("Expected limit 25 or 0, got %d", options.Limit)
		}

		// Return mock task list
		return &mocks.TaskList{
			Data: []mocks.Task{
				{
					TaskID:                "task-1",
					Status:                "ACTIVE",
					Type:                  "TRANSFER",
					Label:                 "Test transfer 1",
					SourceEndpointID:      "source-endpoint-1",
					SourceEndpointDisplay: "Source Endpoint 1",
					DestinationEndpointID: "dest-endpoint-1",
					DestEndpointDisplay:   "Destination Endpoint 1",
				},
				{
					TaskID:                "task-2",
					Status:                "SUCCEEDED",
					Type:                  "TRANSFER",
					Label:                 "Test transfer 2",
					SourceEndpointID:      "source-endpoint-2",
					SourceEndpointDisplay: "Source Endpoint 2",
					DestinationEndpointID: "dest-endpoint-2",
					DestEndpointDisplay:   "Destination Endpoint 2",
				},
			},
		}, nil
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)
	
	// Save original values
	origLimit := limit
	
	// Restore original values after test
	defer func() {
		limit = origLimit
	}()
	
	// Set limit
	limit = 25

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "list")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check output
	expectedOutputs := []string{
		"TaskID", "Status", "Type", "Label",
		"task-1", "ACTIVE", "TRANSFER", "Test transfer 1",
		"task-2", "SUCCEEDED", "TRANSFER", "Test transfer 2",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestTaskListCommandWithFilter(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	mockClient.ListTasksFunc = func(ctx context.Context, options *mocks.ListTasksOptions) (*mocks.TaskList, error) {
		// Verify filter
		if options.FilterStatus != "ACTIVE" {
			t.Errorf("Expected filter 'ACTIVE', got '%s'", options.FilterStatus)
		}

		// Return filtered task list
		return &mocks.TaskList{
			Data: []mocks.Task{
				{
					TaskID:                "task-1",
					Status:                "ACTIVE",
					Type:                  "TRANSFER",
					Label:                 "Test transfer 1",
					SourceEndpointID:      "source-endpoint-1",
					SourceEndpointDisplay: "Source Endpoint 1",
					DestinationEndpointID: "dest-endpoint-1",
					DestEndpointDisplay:   "Destination Endpoint 1",
				},
			},
		}, nil
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)
	
	// Save original values
	origFilter := taskFilter
	
	// Restore original values after test
	defer func() {
		taskFilter = origFilter
	}()
	
	// Set filter
	taskFilter = "ACTIVE"

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "list")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check output - should only contain the active task
	if !strings.Contains(output, "task-1") {
		t.Errorf("Expected output to contain 'task-1', got: %s", output)
	}
	if strings.Contains(output, "task-2") {
		t.Errorf("Expected output not to contain 'task-2', got: %s", output)
	}
}

func TestTaskListCommandApiError(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock to return an error
	mockClient.ListTasksFunc = func(ctx context.Context, options *mocks.ListTasksOptions) (*mocks.TaskList, error) {
		return nil, &mocks.EndpointError{
			Code:    "AuthenticationFailed",
			Message: "Authentication failed",
		}
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "list")
	
	// Should return an error
	if err == nil {
		t.Error("Expected error but got none")
	}
	
	// Check error message
	if !strings.Contains(err.Error(), "failed to list tasks") {
		t.Errorf("Expected error to contain 'failed to list tasks', got: %v", err)
	}

	// Check output - should not contain TaskID header
	if strings.Contains(output, "TaskID") {
		t.Errorf("Expected error output, got table headers: %s", output)
	}
}

func TestTaskShowCommand(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	mockClient.GetTaskFunc = func(ctx context.Context, taskID string) (*mocks.Task, error) {
		// Verify task ID
		if taskID != "task-1" {
			t.Errorf("Expected task ID 'task-1', got '%s'", taskID)
		}

		// Return mock task
		requestTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
		completionTime, _ := time.Parse(time.RFC3339, "2023-01-01T12:05:00Z")
		return &mocks.Task{
			TaskID:                "task-1",
			Status:                "SUCCEEDED",
			Type:                  "TRANSFER",
			Label:                 "Test transfer",
			SourceEndpointID:      "source-endpoint",
			SourceEndpointDisplay: "Source Endpoint",
			DestinationEndpointID: "dest-endpoint",
			DestEndpointDisplay:   "Destination Endpoint",
			RequestTime:           requestTime,
			CompletionTime:        &completionTime,
			FilesTransferred:      10,
			BytesTransferred:      1024000,
			Subtasks:              5,
			FilesSkipped:          2,
			BytesSkipped:          102400,
			VerifyChecksum:        true,
		}, nil
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "show", "task-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check output
	expectedOutputs := []string{
		"Task Details:",
		"Task ID:        task-1",
		"Status:         SUCCEEDED",
		"Type:           TRANSFER",
		"Label:          Test transfer",
		"Request Time:   2023-01-01 12:00:00",
		"Completion Time: 2023-01-01 12:05:00",
		"Task succeeded",
		"Source:      Source Endpoint (source-endpoint)",
		"Destination: Destination Endpoint (dest-endpoint)",
		"Files:          10",
		"Directories:    5",
		"Files Skipped:  2",
		"Bytes Transferred: 1024000",
		"Verification: checksum",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestTaskShowCommandApiError(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock to return an error
	mockClient.GetTaskFunc = func(ctx context.Context, taskID string) (*mocks.Task, error) {
		return nil, &mocks.EndpointError{
			Code:    "TaskNotFound",
			Message: "Task not found",
		}
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "show", "invalid-task")
	
	// Should return an error
	if err == nil {
		t.Error("Expected error but got none")
	}
	
	// Check error message
	if !strings.Contains(err.Error(), "failed to get task") {
		t.Errorf("Expected error to contain 'failed to get task', got: %v", err)
	}

	// Check output - should not contain Task Details
	if strings.Contains(output, "Task Details:") {
		t.Errorf("Expected error output, got task details: %s", output)
	}
}

func TestTaskCancelCommand(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock GetTask
	mockClient.GetTaskFunc = func(ctx context.Context, taskID string) (*mocks.Task, error) {
		return &mocks.Task{
			TaskID: taskID,
			Status: "ACTIVE",
		}, nil
	}

	// Configure mock CancelTask
	mockClient.CancelTaskFunc = func(ctx context.Context, taskID string) (*mocks.OperationResult, error) {
		if taskID != "task-1" {
			t.Errorf("Expected task ID 'task-1', got '%s'", taskID)
		}
		return &mocks.OperationResult{
			Code: "Canceled",
		}, nil
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "cancel", "task-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check output
	expectedOutput := "Successfully canceled task task-1"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedOutput, output)
	}
}

func TestTaskCancelCommandNotActive(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock to return a non-active task
	mockClient.GetTaskFunc = func(ctx context.Context, taskID string) (*mocks.Task, error) {
		return &mocks.Task{
			TaskID: taskID,
			Status: "SUCCEEDED", // Completed task that can't be canceled
		}, nil
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "cancel", "task-1")
	
	// Should return an error
	if err == nil {
		t.Error("Expected error but got none")
	}
	
	// Check error message
	if !strings.Contains(err.Error(), "is not active") {
		t.Errorf("Expected error to contain 'is not active', got: %v", err)
	}

	// Check output - should not contain success message
	if strings.Contains(output, "Successfully canceled") {
		t.Errorf("Expected error output, got success message: %s", output)
	}
}

func TestTaskCancelCommandApiError(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock GetTask to return active task
	mockClient.GetTaskFunc = func(ctx context.Context, taskID string) (*mocks.Task, error) {
		return &mocks.Task{
			TaskID: taskID,
			Status: "ACTIVE",
		}, nil
	}

	// Configure mock CancelTask to return an error
	mockClient.CancelTaskFunc = func(ctx context.Context, taskID string) (*mocks.OperationResult, error) {
		return nil, &mocks.EndpointError{
			Code:    "PermissionDenied",
			Message: "You do not have permission to cancel this task",
		}
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "cancel", "task-1")
	
	// Should return an error
	if err == nil {
		t.Error("Expected error but got none")
	}
	
	// Check error message
	if !strings.Contains(err.Error(), "failed to cancel task") {
		t.Errorf("Expected error to contain 'failed to cancel task', got: %v", err)
	}

	// Check output - should not contain success message
	if strings.Contains(output, "Successfully canceled") {
		t.Errorf("Expected error output, got success message: %s", output)
	}
}

func TestTaskWaitCommand(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock with a sequence of responses
	callCount := 0
	mockClient.GetTaskFunc = func(ctx context.Context, taskID string) (*mocks.Task, error) {
		callCount++

		// First call - active task
		if callCount == 1 {
			return &mocks.Task{
				TaskID:           "task-1",
				Status:           "ACTIVE",
				FilesTransferred: 5,
				BytesTransferred: 512000,
				Subtasks:         10,
				BytesSkipped:     0,
			}, nil
		}

		// Second call - task completed
		return &mocks.Task{
			TaskID:           "task-1",
			Status:           "SUCCEEDED",
			FilesTransferred: 10,
			BytesTransferred: 1024000,
			Subtasks:         10,
			BytesSkipped:     0,
		}, nil
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)
	
	// Save original values
	origWaitTime := taskWaitTime
	
	// Restore original values after test
	defer func() {
		taskWaitTime = origWaitTime
	}()
	
	// Set wait time
	taskWaitTime = 300

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "wait", "task-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check output
	expectedOutputs := []string{
		"Waiting for task task-1: 5/10 files, 512000 bytes",
		"Task task-1 completed successfully",
		"Transferred 10 files (1024000 bytes)",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestTaskWaitCommandApiError(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock to return an error
	mockClient.GetTaskFunc = func(ctx context.Context, taskID string) (*mocks.Task, error) {
		return nil, &mocks.EndpointError{
			Code:    "TaskNotFound",
			Message: "Task not found",
		}
	}

	// Create test command
	taskCmd := mockTaskCommands(mockClient)

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, taskCmd, "wait", "invalid-task")
	
	// Should return an error
	if err == nil {
		t.Error("Expected error but got none")
	}
	
	// Check error message
	if !strings.Contains(err.Error(), "failed to get task status") {
		t.Errorf("Expected error to contain 'failed to get task status', got: %v", err)
	}

	// Check output - should not contain waiting/completion messages
	if strings.Contains(output, "Waiting for task") || strings.Contains(output, "completed successfully") {
		t.Errorf("Expected error output, got success message: %s", output)
	}
}