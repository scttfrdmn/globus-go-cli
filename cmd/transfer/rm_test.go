// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers"
	"github.com/scttfrdmn/globus-go-cli/pkg/testhelpers/mocks"
	"github.com/spf13/cobra"
)

// mockRmCommand creates a mock rm command for testing
func mockRmCommand(mockClient *mocks.MockTransferClient) *cobra.Command {
	cmd := RmCmd()

	// Override the normal RunE function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least one argument")
		}

		// Parse endpoint and path
		endpointID, path := parseEndpointAndPath(args[0])

		// Check if path is valid
		if path == "" {
			return fmt.Errorf("path must be specified")
		}

		// For test purposes, skip confirmation dialog
		rmForce = true

		// Display delete operation details
		fmt.Println("Delete Operation Details:")
		fmt.Printf("  Endpoint: %s\n", endpointID)
		fmt.Printf("  Paths to delete: %d\n", 1)
		fmt.Printf("    1. %s\n", path)
		fmt.Printf("  Recursive: %t\n", rmRecursive)

		// Build delete items list
		var items []mocks.DeleteItem
		itemType := "file"
		if rmRecursive {
			itemType = "directory"
		}
		items = append(items, mocks.DeleteItem{
			DataType: itemType,
			Path:     path,
		})

		// Create delete request
		request := &mocks.DeleteTaskRequest{
			DataType:   "delete",
			EndpointID: endpointID,
			Items:      items,
		}

		// Submit delete task
		taskResponse, err := mockClient.CreateDeleteTask(context.Background(), request)
		if err != nil {
			return fmt.Errorf("failed to submit delete task: %w", err)
		}

		// Display task information
		fmt.Printf("Task ID: %s\n", taskResponse.TaskID)
		fmt.Printf("Delete task submitted successfully. Run 'globus transfer task show %s' to check status.\n", taskResponse.TaskID)

		return nil
	}

	return cmd
}

func TestRmCmd_SingleFile(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path := "/path/to/file.txt"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	var createDeleteTaskCalled bool
	mockClient.CreateDeleteTaskFunc = func(ctx context.Context, request *mocks.DeleteTaskRequest) (*mocks.TaskResponse, error) {
		createDeleteTaskCalled = true
		if request.EndpointID != endpointID {
			t.Errorf("Expected endpoint ID %s, got %s", endpointID, request.EndpointID)
		}
		if len(request.Items) != 1 {
			t.Errorf("Expected 1 delete item, got %d", len(request.Items))
		}
		if request.Items[0].Path != path {
			t.Errorf("Expected path %s, got %s", path, request.Items[0].Path)
		}
		return &mocks.TaskResponse{TaskID: "mock-delete-task-id"}, nil
	}

	// Create test command
	cmd := mockRmCommand(mockClient)

	// Save original values
	origRecursive := rmRecursive
	origForce := rmForce

	// Restore original values after test
	defer func() {
		rmRecursive = origRecursive
		rmForce = origForce
	}()

	// Configure for test
	rmRecursive = false
	rmForce = true

	// Execute the command and capture output
	output, err := testhelpers.ExecuteCommand(t, cmd, endpointID+":"+path)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify mock was called
	if !createDeleteTaskCalled {
		t.Error("Expected CreateDeleteTask to be called")
	}

	// Check output contains expected information
	expectedOutputs := []string{
		"Delete Operation Details:",
		fmt.Sprintf("Endpoint: %s", endpointID),
		"Paths to delete: 1",
		path,
		"Task ID: mock-delete-task-id",
		"Delete task submitted successfully",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestRmCmd_WithRecursive(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path := "/path/to/directory/"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	var createDeleteTaskCalled bool
	mockClient.CreateDeleteTaskFunc = func(ctx context.Context, request *mocks.DeleteTaskRequest) (*mocks.TaskResponse, error) {
		createDeleteTaskCalled = true
		if request.Items[0].DataType != "directory" {
			t.Errorf("Expected data type 'directory' for recursive delete, got %s", request.Items[0].DataType)
		}
		return &mocks.TaskResponse{TaskID: "mock-delete-task-id"}, nil
	}

	// Create test command
	cmd := mockRmCommand(mockClient)

	// Save original values
	origRecursive := rmRecursive
	origForce := rmForce

	// Restore original values after test
	defer func() {
		rmRecursive = origRecursive
		rmForce = origForce
	}()

	// Set recursive flag
	rmRecursive = true
	rmForce = true

	// Execute the command and capture output
	output, err := testhelpers.ExecuteCommand(t, cmd, endpointID+":"+path)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify mock was called
	if !createDeleteTaskCalled {
		t.Error("Expected CreateDeleteTask to be called")
	}

	// Check that output indicates recursive delete
	if !strings.Contains(output, "Recursive: true") {
		t.Errorf("Expected output to indicate recursive delete, got: %s", output)
	}
}

func TestRmCmd_InvalidPath(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Create test command
	cmd := mockRmCommand(mockClient)

	// Set command arguments with invalid path
	cmd.SetArgs([]string{"test-endpoint-id:"}) // Missing path

	// Execute the command and capture output
	var output string
	var err error
	stdout, stderr := testhelpers.CaptureOutput(func() {
		err = cmd.Execute()
	})
	output = stdout + stderr
	
	// Should return an error for missing path
	if err == nil {
		t.Error("Expected error for invalid path, got none")
	}
	
	// Check error message
	if !strings.Contains(err.Error(), "path must be specified") {
		t.Errorf("Expected 'path must be specified' error, got: %v", err)
	}

	// Check that task ID and success message are not in output
	if output != "" && (strings.Contains(output, "Task ID:") || 
						 strings.Contains(output, "Delete task submitted successfully")) {
		t.Errorf("Did not expect task success message in output, got: %s", output)
	}
}

func TestRmCmd_ApiError(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path := "/path/to/file.txt"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock to return an error
	mockClient.CreateDeleteTaskFunc = func(ctx context.Context, request *mocks.DeleteTaskRequest) (*mocks.TaskResponse, error) {
		return nil, &mocks.EndpointError{
			Code:    "PermissionDenied",
			Message: "You do not have permission to delete files on this endpoint",
		}
	}

	// Create test command
	cmd := mockRmCommand(mockClient)

	// Save original values
	origForce := rmForce

	// Restore original values after test
	defer func() {
		rmForce = origForce
	}()

	// Skip confirmation dialog
	rmForce = true

	// Execute the command
	output, err := testhelpers.ExecuteCommand(t, cmd, endpointID+":"+path)

	// Should return an error from the API
	if err == nil {
		t.Error("Expected error from API, got none")
	}

	// Check error message
	if !strings.Contains(err.Error(), "failed to submit delete task") {
		t.Errorf("Expected 'failed to submit delete task' error, got: %v", err)
	}

	// Check that task ID and success message are not in output
	if strings.Contains(output, "Task ID: mock-delete-task-id") ||
		strings.Contains(output, "Delete task submitted successfully") {
		t.Errorf("Did not expect task success message in output, got: %s", output)
	}
}

func TestRmCmd_MultipleFiles(t *testing.T) {
	// This test is simpler without using the ExecuteCommand helper
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	endpointID := "test-endpoint-id"
	path1 := "/path/to/file1.txt"
	path2 := "/path/to/file2.txt"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Track paths that were deleted
	deletedPaths := make([]string, 0)

	// Configure mock
	mockClient.CreateDeleteTaskFunc = func(ctx context.Context, request *mocks.DeleteTaskRequest) (*mocks.TaskResponse, error) {
		for _, item := range request.Items {
			deletedPaths = append(deletedPaths, item.Path)
		}
		return &mocks.TaskResponse{TaskID: "mock-delete-task-id"}, nil
	}

	// Create test commands for each file
	cmd1 := mockRmCommand(mockClient)
	cmd1.SetArgs([]string{endpointID+":"+path1})

	cmd2 := mockRmCommand(mockClient)
	cmd2.SetArgs([]string{endpointID+":"+path2})

	// Save original values
	origForce := rmForce

	// Restore original values after test
	defer func() {
		rmForce = origForce
	}()

	// Skip confirmation dialog
	rmForce = true

	// Execute commands for different files
	var err1, err2 error

	testhelpers.CaptureOutput(func() {
		err1 = cmd1.Execute()
		err2 = cmd2.Execute()
	})

	// Check for errors
	if err1 != nil {
		t.Errorf("Unexpected error for first file: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Unexpected error for second file: %v", err2)
	}

	// Verify all paths were passed
	expectedPaths := []string{path1, path2}
	if len(deletedPaths) != len(expectedPaths) {
		t.Errorf("Expected %d delete paths, got %d", len(expectedPaths), len(deletedPaths))
	}

	for _, expected := range expectedPaths {
		found := false
		for _, actual := range deletedPaths {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Path %s was not deleted", expected)
		}
	}
}