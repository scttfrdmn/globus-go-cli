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

// mockCpCommand creates a mock cp command for testing
func mockCpCommand(mockClient *mocks.MockTransferClient) *cobra.Command {
	cmd := CpCmd()

	// Override the normal RunE function
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("requires exactly 2 arguments (source and destination)")
		}

		// Parse source endpoint and path
		sourceEndpointID, sourcePath := parseEndpointAndPath(args[0])

		// Parse destination endpoint and path
		destEndpointID, destPath := parseEndpointAndPath(args[1])

		// For test purposes, skip confirmation dialog
		transferDryRun = true

		// Display transfer details
		fmt.Println("Transfer Details:")
		fmt.Printf("  Source:      %s:%s\n", sourceEndpointID, sourcePath)
		fmt.Printf("  Destination: %s:%s\n", destEndpointID, destPath)
		fmt.Printf("  Recursive:   %t\n", transferRecursive)
		fmt.Printf("  Sync Level:  %d\n", transferSync)
		fmt.Printf("  Label:       %s\n", transferLabel)

		// Create options map to pass to the transfer API
		optionsMap := map[string]interface{}{
			"recursive":       transferRecursive,
			"sync_level":      transferSync,
			"preserve_mtime":  transferPreserveTime,
			"verify_checksum": transferVerify,
		}

		// Submit the transfer
		taskResponse, err := mockClient.SubmitTransfer(
			context.Background(),
			sourceEndpointID, sourcePath,
			destEndpointID, destPath,
			transferLabel,
			optionsMap,
		)

		if err != nil {
			return fmt.Errorf("failed to submit transfer: %w", err)
		}

		// Display task information
		fmt.Printf("Task ID: %s\n", taskResponse.TaskID)
		fmt.Printf("Task submitted successfully. Run 'globus transfer task show %s' to check status.\n", taskResponse.TaskID)

		return nil
	}

	return cmd
}

func TestCpCmd_Success(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	sourceEndpointID := "source-endpoint-id"
	sourcePath := "/source/path/file.txt"
	destEndpointID := "dest-endpoint-id"
	destPath := "/dest/path/"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	var submitTransferCalled bool
	mockClient.SubmitTransferFunc = func(ctx context.Context,
		srcEndpointID, srcPath, dstEndpointID, dstPath, label string,
		options map[string]interface{}) (*mocks.TaskResponse, error) {

		submitTransferCalled = true
		// Verify parameters
		if srcEndpointID != sourceEndpointID {
			t.Errorf("Expected source endpoint ID %s, got %s", sourceEndpointID, srcEndpointID)
		}
		if srcPath != sourcePath {
			t.Errorf("Expected source path %s, got %s", sourcePath, srcPath)
		}
		if dstEndpointID != destEndpointID {
			t.Errorf("Expected destination endpoint ID %s, got %s", destEndpointID, dstEndpointID)
		}
		if dstPath != destPath {
			t.Errorf("Expected destination path %s, got %s", destPath, dstPath)
		}

		return &mocks.TaskResponse{TaskID: "mock-task-id"}, nil
	}

	// Create test command
	cmd := mockCpCommand(mockClient)

	// Save original values
	origRecursive := transferRecursive
	origSync := transferSync
	origLabel := transferLabel
	origDryRun := transferDryRun

	// Restore original values after test
	defer func() {
		transferRecursive = origRecursive
		transferSync = origSync
		transferLabel = origLabel
		transferDryRun = origDryRun
	}()

	// Configure for test
	transferRecursive = false
	transferSync = 0
	transferLabel = "test-transfer"
	transferDryRun = true

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, cmd,
		sourceEndpointID+":"+sourcePath,
		destEndpointID+":"+destPath)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify mock was called
	if !submitTransferCalled {
		t.Error("Expected SubmitTransfer to be called")
	}

	// Check output contains expected information
	expectedOutputs := []string{
		"Transfer Details:",
		fmt.Sprintf("Source:      %s:%s", sourceEndpointID, sourcePath),
		fmt.Sprintf("Destination: %s:%s", destEndpointID, destPath),
		"Task ID: mock-task-id",
		"Task submitted successfully",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestCpCmd_WithRecursive(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	sourceEndpointID := "source-endpoint-id"
	sourcePath := "/source/path/directory/"
	destEndpointID := "dest-endpoint-id"
	destPath := "/dest/path/"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	mockClient.SubmitTransferFunc = func(ctx context.Context,
		srcEndpointID, srcPath, dstEndpointID, dstPath, label string,
		options map[string]interface{}) (*mocks.TaskResponse, error) {

		// Verify options
		recursive, ok := options["recursive"].(bool)
		if !ok || !recursive {
			t.Errorf("Expected recursive option to be true, got %v", options["recursive"])
		}

		return &mocks.TaskResponse{TaskID: "mock-task-id"}, nil
	}

	// Create test command
	cmd := mockCpCommand(mockClient)

	// Save original values
	origRecursive := transferRecursive
	origDryRun := transferDryRun

	// Restore original values after test
	defer func() {
		transferRecursive = origRecursive
		transferDryRun = origDryRun
	}()

	// Set recursive flag
	transferRecursive = true
	transferDryRun = true

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, cmd,
		sourceEndpointID+":"+sourcePath,
		destEndpointID+":"+destPath)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that output indicates recursive transfer
	if !strings.Contains(output, "Recursive:   true") {
		t.Errorf("Expected output to indicate recursive transfer, got: %s", output)
	}
}

func TestCpCmd_WithSyncLevel(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	sourceEndpointID := "source-endpoint-id"
	sourcePath := "/source/path/directory/"
	destEndpointID := "dest-endpoint-id"
	destPath := "/dest/path/"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	mockClient.SubmitTransferFunc = func(ctx context.Context,
		srcEndpointID, srcPath, dstEndpointID, dstPath, label string,
		options map[string]interface{}) (*mocks.TaskResponse, error) {

		// Verify options
		syncLevel, ok := options["sync_level"].(int)
		if !ok || syncLevel != 2 {
			t.Errorf("Expected sync_level option to be 2, got %v", options["sync_level"])
		}

		return &mocks.TaskResponse{TaskID: "mock-task-id"}, nil
	}

	// Create test command
	cmd := mockCpCommand(mockClient)

	// Save original values
	origSync := transferSync
	origDryRun := transferDryRun

	// Restore original values after test
	defer func() {
		transferSync = origSync
		transferDryRun = origDryRun
	}()

	// Set sync level
	transferSync = 2
	transferDryRun = true

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, cmd,
		sourceEndpointID+":"+sourcePath,
		destEndpointID+":"+destPath)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check that output indicates sync level
	if !strings.Contains(output, "Sync Level:  2") {
		t.Errorf("Expected output to indicate sync level 2, got: %s", output)
	}
}

func TestCpCmd_InvalidArguments(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	sourceEndpointID := "source-endpoint-id"
	sourcePath := "/source/path/file.txt"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Create test command
	cmd := mockCpCommand(mockClient)

	// Save original value
	origDryRun := transferDryRun

	// Restore original value after test
	defer func() {
		transferDryRun = origDryRun
	}()

	// Configure for test
	transferDryRun = true

	// Set insufficient arguments
	cmd.SetArgs([]string{
		sourceEndpointID + ":" + sourcePath,
		// Missing destination argument
	})

	// Execute command
	err := cmd.Execute()

	// Should return an error for insufficient arguments
	if err == nil {
		t.Error("Expected error for insufficient arguments, got none")
	}

	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("Expected error about requiring 2 arguments, got: %v", err)
	}
}

func TestCpCmd_ApiError(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	sourceEndpointID := "source-endpoint-id"
	sourcePath := "/source/path/file.txt"
	destEndpointID := "dest-endpoint-id"
	destPath := "/dest/path/"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock to return an error
	mockClient.SubmitTransferFunc = func(ctx context.Context,
		srcEndpointID, srcPath, dstEndpointID, dstPath, label string,
		options map[string]interface{}) (*mocks.TaskResponse, error) {

		return nil, &mocks.EndpointError{
			Code:    "PermissionDenied",
			Message: "You do not have permission to access this endpoint",
		}
	}

	// Create test command
	cmd := mockCpCommand(mockClient)

	// Save original value
	origDryRun := transferDryRun

	// Restore original value after test
	defer func() {
		transferDryRun = origDryRun
	}()

	// Configure for test
	transferDryRun = true

	// Execute command
	output, err := testhelpers.ExecuteCommand(t, cmd,
		sourceEndpointID+":"+sourcePath,
		destEndpointID+":"+destPath)

	// Should return an error from the API
	if err == nil {
		t.Error("Expected error from API, got none")
	}

	// Check error message
	if !strings.Contains(err.Error(), "failed to submit transfer") {
		t.Errorf("Expected 'failed to submit transfer' error, got: %v", err)
	}

	// Check that task ID and success message are not in output
	if strings.Contains(output, "Task ID: mock-task-id") ||
		strings.Contains(output, "Task submitted successfully") {
		t.Errorf("Did not expect task success message in output, got: %s", output)
	}
}

func TestCpCmd_WithPreserveTimestamp(t *testing.T) {
	// Setup a temporary token file that tests can use
	_, cleanup := testhelpers.SetupTokenFile(t)
	defer cleanup()

	// Set up test parameters
	sourceEndpointID := "source-endpoint-id"
	sourcePath := "/source/path/file.txt"
	destEndpointID := "dest-endpoint-id"
	destPath := "/dest/path/"

	// Create mock client
	mockClient := &mocks.MockTransferClient{}

	// Configure mock
	mockClient.SubmitTransferFunc = func(ctx context.Context,
		srcEndpointID, srcPath, dstEndpointID, dstPath, label string,
		options map[string]interface{}) (*mocks.TaskResponse, error) {

		// Verify options
		preserveMtime, ok := options["preserve_mtime"].(bool)
		if !ok || !preserveMtime {
			t.Errorf("Expected preserve_mtime option to be true, got %v", options["preserve_mtime"])
		}

		return &mocks.TaskResponse{TaskID: "mock-task-id"}, nil
	}

	// Create test command
	cmd := mockCpCommand(mockClient)

	// Save original values
	origPreserveTime := transferPreserveTime
	origDryRun := transferDryRun

	// Restore original values after test
	defer func() {
		transferPreserveTime = origPreserveTime
		transferDryRun = origDryRun
	}()

	// Set preserve timestamp flag
	transferPreserveTime = true
	transferDryRun = true

	// Execute command
	_, err := testhelpers.ExecuteCommand(t, cmd,
		sourceEndpointID+":"+sourcePath,
		destEndpointID+":"+destPath)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
